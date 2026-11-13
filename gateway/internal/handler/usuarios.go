package handler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	"Prueba-Go/gateway/internal/storage"
	authpb "Prueba-Go/gen/auth"
	cursospb "Prueba-Go/gen/cursos"
	examenespb "Prueba-Go/gen/examenes"
	leccionespb "Prueba-Go/gen/lecciones"
	usuariospb "Prueba-Go/gen/usuarios"

	"github.com/gin-gonic/gin"
)

type UsuariosHandler struct{ c *clients.Clients }

func NewUsuariosHandler(c *clients.Clients) *UsuariosHandler { return &UsuariosHandler{c: c} }

// GET /api/perfil
//
// Devuelve QUIÉN es el usuario, y nada más. Es la llamada que hace el layout en
// cada página para pintar el nombre y el avatar, así que tiene que ser barata:
// una sola consulta a usuarios-service.
//
// Las estadísticas se fueron a /api/perfil/estadisticas. Estaban aquí, y eso
// convertía cada carga de página en cinco llamadas gRPC en serie más una por
// curso inscrito. Medido en producción: 16 y 17 segundos. Solo dos pantallas
// —el perfil del alumno y el del instructor— leen esos números; el resto de la
// plataforma los pagaba sin usarlos.
//
// `stats` sigue viajando, vacío, para no romper a quien lo lea sin haberse
// actualizado: un cliente antiguo verá ceros en lugar de reventar.
func (h *UsuariosHandler) GetPerfil(ctx *gin.Context) {
	resp, err := h.c.Usuarios.GetPerfil(ctx.Request.Context(), &usuariospb.GetPerfilRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"user": resp, "stats": gin.H{}})
}

// GET /api/perfil/estadisticas
//
// Los contadores del perfil, repartidos entre cuatro servicios con su propia
// base de datos. Solo lo piden las dos pantallas que los muestran.
//
// Las consultas van en PARALELO. En serie, el coste era la suma de todas —cinco
// viajes más uno por curso— y en Railway cada viaje ronda los 200 ms. En
// paralelo el coste es el de la más lenta.
func (h *UsuariosHandler) GetEstadisticasPerfil(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxUserID)
	// Contexto propio: sobrevive a las goroutines aunque gin recicle el suyo.
	c, cancelar := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancelar()

	var (
		mu    sync.Mutex
		wg    sync.WaitGroup
		stats = gin.H{
			"cursos_inscritos": 0, "lecciones_completadas": 0, "total_lecciones": 0,
			"cursos_creados": 0, "estudiantes_total": 0, "examenes_creados": 0,
		}
	)
	guardar := func(clave string, valor int32) {
		mu.Lock()
		stats[clave] = valor
		mu.Unlock()
	}

	// Cada contador que un fallo deje sin resolver se queda en cero y no corta
	// la respuesta: media pantalla de estadísticas es mejor que ninguna.
	wg.Add(3)
	go func() {
		defer wg.Done()
		if r, err := h.c.Cursos.InstructorListCapacitaciones(c, &cursospb.UserRequest{UserId: userID}); err == nil {
			guardar("cursos_creados", int32(len(r.Cursos)))
		}
	}()
	go func() {
		defer wg.Done()
		if r, err := h.c.Cursos.InstructorListEstudiantes(c, &cursospb.UserRequest{UserId: userID}); err == nil {
			guardar("estudiantes_total", int32(len(r.Estudiantes)))
		}
	}()
	go func() {
		defer wg.Done()
		if r, err := h.c.Examenes.InstructorListExamenes(c, &examenespb.UserRequest{UserId: userID}); err == nil {
			guardar("examenes_creados", int32(len(r.Examenes)))
		}
	}()

	// El avance NO se lee de los cursos que llegan de cursos-service: esa base
	// devuelve `0 as total_lecciones, 0 as lecciones_completadas` literales,
	// porque las lecciones y su progreso viven en la base de lecciones. Sumarlas
	// daba cero siempre y el perfil mostraba "0%" aunque el curso estuviera
	// terminado.
	if cursos, err := h.c.Cursos.ListMisCapacitaciones(c, &cursospb.UserRequest{UserId: userID}); err == nil {
		guardar("cursos_inscritos", int32(len(cursos.Cursos)))

		var comp, total int32
		var muAvance sync.Mutex
		var wgAvance sync.WaitGroup
		for _, curso := range cursos.Cursos {
			wgAvance.Add(1)
			go func(cursoID string) {
				defer wgAvance.Done()
				res, errA := h.c.Lecciones.ResumenAvanceCurso(c, &leccionespb.ResumenAvanceRequest{
					CursoId: cursoID,
					UserIds: []string{userID},
				})
				if errA != nil || res == nil {
					return
				}
				muAvance.Lock()
				defer muAvance.Unlock()
				// TotalLecciones viene aunque el alumno no tenga ni una fila de
				// progreso; sin él, un curso recién empezado contaría 0 de 0, que
				// se lee como completado.
				total += res.TotalLecciones
				if len(res.Avances) > 0 {
					comp += res.Avances[0].Completadas
				}
			}(curso.Id)
		}
		wgAvance.Wait()
		guardar("lecciones_completadas", comp)
		guardar("total_lecciones", total)
	}

	wg.Wait()
	ctx.JSON(http.StatusOK, stats)
}

// PUT /api/perfil
func (h *UsuariosHandler) UpdatePerfil(ctx *gin.Context) {
	var body struct {
		Name      string `json:"name"`
		Bio       string `json:"bio"`
		Phone     string `json:"phone"`
		Specialty string `json:"specialty"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Usuarios.UpdatePerfil(ctx.Request.Context(), &usuariospb.UpdatePerfilRequest{
		UserId:    ctx.GetString(middleware.CtxUserID),
		Name:      body.Name,
		Bio:       body.Bio,
		Phone:     body.Phone,
		Specialty: body.Specialty,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"user": resp})
}

// POST /api/perfil/avatar
func (h *UsuariosHandler) UploadAvatar(ctx *gin.Context) {
	url, err := uploadFileToR2(ctx, "avatars")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Usuarios.UpdateAvatarURL(ctx.Request.Context(), &usuariospb.UpdateMediaURLRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
		Url:    url,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": resp.AvatarUrl})
}

// POST /api/perfil/cover
func (h *UsuariosHandler) UploadCover(ctx *gin.Context) {
	url, err := uploadFileToR2(ctx, "covers")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Usuarios.UpdateCoverURL(ctx.Request.Context(), &usuariospb.UpdateMediaURLRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
		Url:    url,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": resp.CoverUrl})
}

// GET /api/usuarios/:id/perfil
func (h *UsuariosHandler) GetPublicPerfil(ctx *gin.Context) {
	resp, err := h.c.Usuarios.GetPublicPerfil(ctx.Request.Context(), &usuariospb.UserIDRequest{
		UserId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"user": resp})
}

// GET /api/admin/users
func (h *UsuariosHandler) ListUsers(ctx *gin.Context) {
	resp, err := h.c.Usuarios.ListUsers(ctx.Request.Context(), &usuariospb.ListUsersRequest{
		Role: ctx.Query("role"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Users)
}

// POST /api/admin/users/:id/revoke-sessions
func (h *UsuariosHandler) RevokeUserSessions(ctx *gin.Context) {
	_, err := h.c.Auth.RevokeUserSessions(ctx.Request.Context(), &authpb.RevokeRequest{
		UserId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "sesiones revocadas"})
}

// GET /api/usuarios/search
//
// Se compone de dos llamadas porque el dato está repartido: quién comparte
// capacitación conmigo lo sabe cursos-service, dueño de `inscripciones`,
// `asignaciones` y `capacitaciones`; los nombres y correos los tiene
// usuarios-service. Antes usuarios-service consultaba las tres tablas por su
// cuenta y en producción, con una base por servicio, respondía 500.
func (h *UsuariosHandler) SearchUsers(ctx *gin.Context) {
	q := ctx.Query("q")
	if q == "" {
		ctx.JSON(http.StatusOK, []interface{}{})
		return
	}
	userID := ctx.GetString(middleware.CtxUserID)
	rol := ctx.GetString(middleware.CtxUserRole)

	// Admin e instructor buscan sin restricción, así que se ahorran la llamada
	// a cursos. usuarios-service vuelve a comprobar el rol contra la base: el
	// del token sirve para decidir si hace falta la lista, no para autorizar.
	var soloIDs []string
	if rol != "admin" && rol != "instructor" {
		companeros, err := h.c.Cursos.CompanerosDeCurso(ctx.Request.Context(), &cursospb.CompanerosRequest{
			UserId: userID,
			// Un curso masivo puede tener miles de compañeros y el buscador
			// muestra diez. Se pide un techo amplio para que el filtro por
			// nombre siga teniendo de dónde elegir, no uno igual al de salida.
			Limite: 500,
		})
		if err != nil {
			// Se corta aquí en vez de seguir sin lista. Continuar dejaría
			// soloIDs vacío y, si algún día alguien invierte ese significado en
			// usuarios-service, la caída de cursos-service se convertiría en
			// una fuga del directorio completo.
			grpcToHTTP(ctx, err)
			return
		}
		if len(companeros.UserIds) == 0 {
			ctx.JSON(http.StatusOK, []interface{}{})
			return
		}
		soloIDs = companeros.UserIds
	}

	resp, err := h.c.Usuarios.SearchUsers(ctx.Request.Context(), &usuariospb.SearchUsersRequest{
		Query:       q,
		Limit:       10,
		RequesterId: userID,
		SoloIds:     soloIDs,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Users)
}

// GET /api/notificaciones
func (h *UsuariosHandler) ListNotificaciones(ctx *gin.Context) {
	resp, err := h.c.Usuarios.ListNotificaciones(ctx.Request.Context(), &usuariospb.UserIDRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Notificaciones)
}

// POST /api/notificaciones/marcar-leidas
func (h *UsuariosHandler) MarcarNotificacionesLeidas(ctx *gin.Context) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.c.Usuarios.MarkNotificacionesRead(ctx.Request.Context(), &usuariospb.MarkNotificacionesReadRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
		Ids:    body.IDs,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// uploadFileToR2 lee el multipart del contexto y lo sube a R2.
func uploadFileToR2(ctx *gin.Context, folder string) (string, error) {
	fh, err := ctx.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("archivo requerido")
	}
	return storage.UploadMultipart(ctx.Request.Context(), fh, folder)
}

// PATCH /api/admin/users/:id/role
func (h *UsuariosHandler) AdminUpdateRole(ctx *gin.Context) {
	targetID := ctx.Param("id")
	var body struct {
		Role string `json:"role"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Usuarios.AdminUpdateRole(ctx.Request.Context(), &usuariospb.AdminUpdateRoleRequest{
		AdminId:      ctx.GetString(middleware.CtxUserID),
		TargetUserId: targetID,
		NewRole:      body.Role,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
