package handler

import (
	"net/http"

	"Prueba-Go/gateway/internal/middleware"
	cursospb "Prueba-Go/gen/cursos"
	examenespb "Prueba-Go/gen/examenes"
	leccionespb "Prueba-Go/gen/lecciones"
	usuariospb "Prueba-Go/gen/usuarios"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────────────────────────────────
// Seguimiento de alumnos por curso (instructor)
//
// Se compone aquí y no en un servicio porque los datos están repartidos entre
// cuatro bases distintas: las inscripciones en cursos, el avance y los puntos en
// lecciones, las calificaciones en examenes y los nombres en auth. El gateway es
// el único que habla con los cuatro.
// ─────────────────────────────────────────────────────────────────────────────

// GET /api/instructor/capacitaciones/:id/inscritos
//
// Lista de seguimiento: un alumno por fila, con su avance y sus puntos.
//
// Incluye a quien se inscribió y no ha empezado. Filtrarlo por no tener progreso
// dejaría fuera precisamente a quien hay que perseguir.
func (h *CursosHandler) InstructorListInscritos(ctx *gin.Context) {
	cursoID := ctx.Param("id")
	instructorID := ctx.GetString(middleware.CtxUserID)

	// El servicio comprueba que el curso sea de este instructor. El middleware
	// solo valida el ROL, así que sin esa comprobación bastaría cambiar el id de
	// la URL para leer los alumnos de un curso ajeno.
	inscritos, err := h.c.Cursos.InstructorListInscritos(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: cursoID,
		UserId:  instructorID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	ids := make([]string, 0, len(inscritos.Inscritos))
	for _, i := range inscritos.Inscritos {
		ids = append(ids, i.UserId)
	}

	// Avance de todos en una sola llamada, no una por alumno.
	avancePorUsuario := map[string]*leccionespb.AvanceAlumno{}
	totalLecciones := int32(0)
	if len(ids) > 0 {
		res, errA := h.c.Lecciones.ResumenAvanceCurso(ctx.Request.Context(), &leccionespb.ResumenAvanceRequest{
			CursoId: cursoID,
			UserIds: ids,
		})
		if errA == nil && res != nil {
			totalLecciones = res.TotalLecciones
			for _, a := range res.Avances {
				avancePorUsuario[a.UserId] = a
			}
		}
	}

	perfiles := h.resolverNombres(ctx.Request.Context(), ids)

	filas := make([]gin.H, 0, len(inscritos.Inscritos))
	for _, i := range inscritos.Inscritos {
		u := perfiles[i.UserId]
		a := avancePorUsuario[i.UserId]

		completadas := int32(0)
		puntos := int32(0)
		ultima := ""
		if a != nil {
			completadas, puntos, ultima = a.Completadas, a.Puntos, a.UltimaActividad
		}

		filas = append(filas, gin.H{
			"user_id":          i.UserId,
			"nombre":           nombreOr(u, "Cuenta eliminada"),
			"email":            correoDe(u),
			"avatar_url":       avatarDe(u),
			"inscrito_at":      i.InscritoAt,
			"por_licencia":     i.LicenciaId != "",
			"completadas":      completadas,
			"total_lecciones":  totalLecciones,
			"ultima_actividad": ultima,
			"puntos":           puntos,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"inscritos":       filas,
		"total_lecciones": totalLecciones,
	})
}

// GET /api/instructor/capacitaciones/:id/inscritos/:user_id
//
// Detalle de un alumno: lección por lección, sus exámenes y sus puntos.
func (h *CursosHandler) InstructorDetalleAlumno(ctx *gin.Context) {
	cursoID := ctx.Param("id")
	alumnoID := ctx.Param("user_id")
	instructorID := ctx.GetString(middleware.CtxUserID)

	// Se reutiliza la comprobación de propiedad del listado: si el curso no es
	// de este instructor, aquí termina.
	if _, err := h.c.Cursos.InstructorListInscritos(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: cursoID,
		UserId:  instructorID,
	}); err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	// ── Lecciones con el progreso del ALUMNO ─────────────────────────────
	//
	// El RPC acepta un user_id; la ruta del alumno lo fija al del token, pero
	// aquí el instructor ya demostró ser dueño del curso.
	lecciones := []gin.H{}
	completadas := 0
	if res, err := h.c.Lecciones.GetLeccionesConProgreso(ctx.Request.Context(), &leccionespb.CursoUserRequest{
		CursoId: cursoID,
		UserId:  alumnoID,
	}); err == nil && res != nil {
		for _, l := range res.Lecciones {
			if l.Completada {
				completadas++
			}
			lecciones = append(lecciones, gin.H{
				"id":           l.Id,
				"title":        l.Title,
				"tipo":         l.LessonType.String(),
				"orden":        l.Orden,
				"duracion_min": l.DuracionMin,
				"completada":   l.Completada,
			})
		}
	}

	// ── Exámenes del curso con el resultado del alumno ───────────────────
	//
	// No hay un "resultados por curso": se listan los exámenes del instructor,
	// se filtran por capacitación y se busca a este alumno en cada uno. Son
	// pocas llamadas porque un curso tiene pocos exámenes.
	examenes := []gin.H{}
	if lista, err := h.c.Examenes.InstructorListExamenes(ctx.Request.Context(),
		&examenespb.UserRequest{UserId: instructorID}); err == nil && lista != nil {
		for _, ex := range lista.Examenes {
			if ex.CapacitacionId != cursoID {
				continue
			}
			fila := gin.H{
				"id":         ex.Id,
				"title":      ex.Title,
				"presentado": false,
			}
			if res, errR := h.c.Examenes.InstructorGetResultados(ctx.Request.Context(),
				&examenespb.ExamenRequest{ExamenId: ex.Id}); errR == nil && res != nil {
				for _, r := range res.Resultados {
					if r.UserId != alumnoID {
						continue
					}
					fila["presentado"] = true
					fila["puntaje"] = r.Puntaje
					fila["porcentaje"] = r.Porcentaje
					fila["submitted_at"] = r.SubmittedAt
					break
				}
			}
			examenes = append(examenes, fila)
		}
	}

	// ── Puntos y datos del alumno ────────────────────────────────────────
	puntos := int32(0)
	ultima := ""
	if res, err := h.c.Lecciones.ResumenAvanceCurso(ctx.Request.Context(), &leccionespb.ResumenAvanceRequest{
		CursoId: cursoID,
		UserIds: []string{alumnoID},
	}); err == nil && res != nil && len(res.Avances) > 0 {
		puntos, ultima = res.Avances[0].Puntos, res.Avances[0].UltimaActividad
	}

	perfiles := h.resolverNombres(ctx.Request.Context(), []string{alumnoID})
	u := perfiles[alumnoID]

	ctx.JSON(http.StatusOK, gin.H{
		"user_id":    alumnoID,
		"nombre":     nombreOr(u, "Cuenta eliminada"),
		"email":      correoDe(u),
		"avatar_url": avatarDe(u),
		// El curso se da por terminado cuando no queda ninguna lección
		// pendiente. Con cero lecciones NO se considera completado: un curso
		// vacío no acredita nada, y darlo por bueno emitiría constancias de un
		// contenido que no existe.
		"completado":       len(lecciones) > 0 && completadas == len(lecciones),
		"completadas":      completadas,
		"total_lecciones":  len(lecciones),
		"ultima_actividad": ultima,
		"puntos":           puntos,
		"lecciones":        lecciones,
		"examenes":         examenes,
	})
}

// avatarDe saca el avatar de un perfil que puede no haberse podido resolver.
func avatarDe(u *usuariospb.PerfilResponse) string {
	if u == nil {
		return ""
	}
	return u.AvatarUrl
}
