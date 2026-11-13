package handler

import (
	"net/http"

	"Prueba-Go/gateway/internal/middleware"
	cursospb "Prueba-Go/gen/cursos"
	usuariospb "Prueba-Go/gen/usuarios"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────────────────────────────────
// Instructores asignados a un curso para revisarlo
//
// `capacitaciones.instructor_id` es el dueño: lo creó y es el único que puede
// editarlo o borrarlo. Esto es otra cosa: instructores que el administrador suma
// para que revisen el avance de los alumnos y los resultados de los exámenes.
//
// Se compone aquí porque los nombres viven en la base de auth y la asignación en
// la de cursos. El gateway es el único que habla con las dos.
// ─────────────────────────────────────────────────────────────────────────────

// GET /api/admin/capacitaciones/:id/instructores
func (h *CursosHandler) ListInstructoresDeCurso(ctx *gin.Context) {
	resp, err := h.c.Cursos.ListInstructoresDeCurso(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	ids := make([]string, 0, len(resp.Instructores))
	for _, i := range resp.Instructores {
		ids = append(ids, i.InstructorId)
	}
	perfiles := h.resolverNombres(ctx.Request.Context(), ids)

	filas := make([]gin.H, 0, len(resp.Instructores))
	for _, i := range resp.Instructores {
		u := perfiles[i.InstructorId]
		filas = append(filas, gin.H{
			"instructor_id": i.InstructorId,
			// "Cuenta eliminada" y no un hueco: una asignación que apunta a un
			// usuario borrado sigue dando acceso a nada, pero el administrador
			// tiene que verla para poder quitarla.
			"nombre":      nombreOr(u, "Cuenta eliminada"),
			"email":       correoDe(u),
			"avatar_url":  avatarDe(u),
			"asignado_at": i.AsignadoAt,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"instructores": filas})
}

// POST /api/admin/capacitaciones/:id/instructores
func (h *CursosHandler) AsignarInstructorACurso(ctx *gin.Context) {
	var body struct {
		InstructorID string `json:"instructor_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "falta instructor_id"})
		return
	}

	// Que la persona exista Y sea instructor se comprueba ANTES de asignar.
	//
	// cursos-service no puede hacerlo: `users` vive en la base de auth. Sin esta
	// comprobación se podría asignar un identificador inventado, o el de un
	// alumno, y la fila quedaría ahí dando un permiso que nadie entiende de
	// dónde salió.
	perfil, err := h.c.Usuarios.GetPublicPerfil(ctx.Request.Context(),
		&usuariospb.UserIDRequest{UserId: body.InstructorID})
	if err != nil || perfil == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ese usuario no existe"})
		return
	}
	if perfil.Role != "instructor" && perfil.Role != "admin" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "solo se puede asignar a alguien con perfil de instructor",
		})
		return
	}

	if _, err := h.c.Cursos.AdminAsignarInstructor(ctx.Request.Context(), &cursospb.AsignarInstructorRequest{
		CapacitacionId: ctx.Param("id"),
		InstructorId:   body.InstructorID,
		// Se guarda quién asignó. Es la primera pregunta cuando alguien ve datos
		// de alumnos que no esperaba tener delante.
		AsignadoPor: ctx.GetString(middleware.CtxUserID),
	}); err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"ok": true})
}

// DELETE /api/admin/capacitaciones/:id/instructores/:instructor_id
func (h *CursosHandler) QuitarInstructorDeCurso(ctx *gin.Context) {
	if _, err := h.c.Cursos.AdminQuitarInstructor(ctx.Request.Context(), &cursospb.AsignarInstructorRequest{
		CapacitacionId: ctx.Param("id"),
		InstructorId:   ctx.Param("instructor_id"),
	}); err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
