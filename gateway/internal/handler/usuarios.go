package handler

import (
	"fmt"
	"net/http"

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
func (h *UsuariosHandler) GetPerfil(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxUserID)
	resp, err := h.c.Usuarios.GetPerfil(ctx.Request.Context(), &usuariospb.GetPerfilRequest{
		UserId: userID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	// Enriquecer stats desde los microservicios correspondientes
	// (el servicio de usuarios puede no tener acceso a esas tablas en producción)
	stats := gin.H{
		"cursos_inscritos":      resp.CursosInscritos,
		"lecciones_completadas": resp.LeccionesCompletadas,
		"total_lecciones":       resp.TotalLecciones,
		"cursos_creados":        resp.CursosCreados,
		"estudiantes_total":     resp.EstudiantesTotal,
		"examenes_creados":      resp.ExamenesCreados,
	}

	// Cursos creados
	if cursosResp, errC := h.c.Cursos.InstructorListCapacitaciones(ctx.Request.Context(), &cursospb.UserRequest{UserId: userID}); errC == nil {
		stats["cursos_creados"] = int32(len(cursosResp.Cursos))
	}
	// Estudiantes totales
	if estResp, errEst := h.c.Cursos.InstructorListEstudiantes(ctx.Request.Context(), &cursospb.UserRequest{UserId: userID}); errEst == nil {
		stats["estudiantes_total"] = int32(len(estResp.Estudiantes))
	}
	// Exámenes creados
	if examResp, errE := h.c.Examenes.InstructorListExamenes(ctx.Request.Context(), &examenespb.UserRequest{UserId: userID}); errE == nil {
		stats["examenes_creados"] = int32(len(examResp.Examenes))
	}
	// Cursos inscritos y lecciones completadas
	if cursosResp, errC := h.c.Cursos.ListMisCapacitaciones(ctx.Request.Context(), &cursospb.UserRequest{UserId: userID}); errC == nil {
		stats["cursos_inscritos"] = int32(len(cursosResp.Cursos))

		// El avance NO se lee de los cursos que acaban de llegar.
		//
		// cursos-service devuelve esas dos columnas como `0 as total_lecciones,
		// 0 as lecciones_completadas` literales: las lecciones y su progreso
		// viven en la base de lecciones, no en la suya. Sumarlas daba cero
		// siempre, y el perfil mostraba "0%" a todo el mundo aunque hubiera
		// terminado el curso. El handler de /api/mis-capacitaciones sí
		// enriquece; este, que llama al gRPC directo, se lo saltaba.
		var comp, total int32
		for _, c := range cursosResp.Cursos {
			res, errA := h.c.Lecciones.ResumenAvanceCurso(ctx.Request.Context(), &leccionespb.ResumenAvanceRequest{
				CursoId: c.Id,
				UserIds: []string{userID},
			})
			if errA != nil || res == nil {
				continue
			}
			// TotalLecciones viene aunque el alumno no tenga ni una fila de
			// progreso; sin él, un curso recién empezado contaría 0 de 0, que
			// se lee como completado.
			total += res.TotalLecciones
			if len(res.Avances) > 0 {
				comp += res.Avances[0].Completadas
			}
		}
		stats["lecciones_completadas"] = comp
		stats["total_lecciones"] = total
	}

	ctx.JSON(http.StatusOK, gin.H{"user": resp, "stats": stats})
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
