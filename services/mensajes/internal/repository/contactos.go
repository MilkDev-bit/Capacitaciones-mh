package repository

import (
	"context"
	"fmt"

	cursospb "Prueba-Go/gen/cursos"

	"github.com/jmoiron/sqlx"
)

// Este archivo concentra la regla de visibilidad entre usuarios:
// "solo puedes escribir a quien comparte curso contigo".
//
// Por qué vive en este servicio y no en el gateway: la regla es una restricción
// de dominio, no de transporte. Si se validara solo en el gateway o en el
// frontend, cualquiera podría hacer POST /api/mensajes/<uuid-arbitrario> y
// saltársela — el filtro de búsqueda de usuarios que ya existía era exactamente
// eso, cosmético. Al validarla en el servicio, ninguna ruta de entrada puede
// evitarla.
//
// Qué se consulta y dónde:
//
//	quién comparte curso conmigo → cursos-service, por gRPC
//	quién pertenece a un grupo   → esta base de datos
//
// Antes lo primero también salía de aquí, leyendo `inscripciones`,
// `asignaciones` y `capacitaciones` con SQL directo. Esas tablas son de
// cursos-service. En desarrollo colaba porque docker-compose da el mismo
// DATABASE_URL a los siete contenedores; en producción cada servicio tiene su
// propia base, la consulta fallaba con "relation does not exist" y el error se
// traducía a "no se pudo validar el destinatario": ningún mensaje directo podía
// enviarse, y tampoco se podían crear grupos.

// ContactosRepository resuelve a quién puede contactar un usuario.
type ContactosRepository interface {
	// FiltrarContactables devuelve el subconjunto de targetIDs que el
	// solicitante tiene permitido contactar. Nunca devuelve al propio
	// solicitante.
	FiltrarContactables(ctx context.Context, requesterID string, targetIDs []string) ([]string, error)
	// PuedeContactar es el caso de un solo destinatario.
	PuedeContactar(ctx context.Context, requesterID, targetID string) (bool, error)
	// EsMiembroDeGrupo indica si el usuario pertenece al grupo.
	EsMiembroDeGrupo(ctx context.Context, userID, grupoID string) (bool, error)
	// AdminDeGrupo devuelve el admin_id del grupo, o "" si no existe.
	AdminDeGrupo(ctx context.Context, grupoID string) (string, error)
}

type postgresContactosRepository struct {
	db     *sqlx.DB
	cursos cursospb.CursosServiceClient
}

func NewContactosRepository(db *sqlx.DB, cursos cursospb.CursosServiceClient) ContactosRepository {
	return &postgresContactosRepository{db: db, cursos: cursos}
}

func (r *postgresContactosRepository) FiltrarContactables(ctx context.Context, requesterID string, targetIDs []string) ([]string, error) {
	if requesterID == "" || len(targetIDs) == 0 {
		return nil, nil
	}

	// Se mandan los candidatos para que cursos-service devuelva solo el
	// subconjunto permitido, en vez de traerse la lista entera y cruzarla aquí.
	// En un curso masivo la diferencia es entre unos pocos identificadores y
	// varios miles por cada mensaje enviado.
	//
	// Sin límite: el conjunto ya viene acotado por los candidatos, y recortarlo
	// descartaría en silencio a destinatarios legítimos.
	resp, err := r.cursos.CompanerosDeCurso(ctx, &cursospb.CompanerosRequest{
		UserId:     requesterID,
		Candidatos: targetIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("contactos: consultar compañeros: %w", err)
	}

	// El propio solicitante nunca sale: cursos-service ya lo excluye, pero la
	// promesa es de esta interfaz y se cumple aquí también por si esa consulta
	// cambia.
	permitidos := make([]string, 0, len(resp.UserIds))
	for _, id := range resp.UserIds {
		if id != requesterID {
			permitidos = append(permitidos, id)
		}
	}
	return permitidos, nil
}

func (r *postgresContactosRepository) PuedeContactar(ctx context.Context, requesterID, targetID string) (bool, error) {
	permitidos, err := r.FiltrarContactables(ctx, requesterID, []string{targetID})
	if err != nil {
		return false, err
	}
	return len(permitidos) == 1, nil
}

// Las dos de abajo se quedan en SQL: `grupos` y `grupo_miembros` son tablas de
// este servicio, en esta misma base.

func (r *postgresContactosRepository) EsMiembroDeGrupo(ctx context.Context, userID, grupoID string) (bool, error) {
	var existe bool
	err := r.db.GetContext(ctx, &existe,
		`SELECT EXISTS (SELECT 1 FROM grupo_miembros WHERE grupo_id = $1 AND usuario_id = $2)`,
		grupoID, userID)
	if err != nil {
		return false, fmt.Errorf("contactos: verificar membresía: %w", err)
	}
	return existe, nil
}

func (r *postgresContactosRepository) AdminDeGrupo(ctx context.Context, grupoID string) (string, error) {
	var adminID string
	err := r.db.GetContext(ctx, &adminID, `SELECT admin_id FROM grupos WHERE id = $1`, grupoID)
	if err != nil {
		return "", fmt.Errorf("contactos: obtener admin del grupo: %w", err)
	}
	return adminID, nil
}
