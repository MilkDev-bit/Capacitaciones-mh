// Package router registra todas las rutas HTTP del gateway.
// main.go solo inyecta dependencias y llama a router.New().
package router

import (
	"net/http"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/config"
	"Prueba-Go/gateway/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Deps agrupa todos los handlers y middlewares que necesita el router.
type Deps struct {
	Cfg          *config.Config
	AuthH        *handler.AuthHandler
	UsuariosH    *handler.UsuariosHandler
	CursosH      *handler.CursosHandler
	LeccionesH   *handler.LeccionesHandler
	ExamenesH    *handler.ExamenesHandler
	ForosH       *handler.ForosHandler
	MensajesH    *handler.MensajesHandler
	WsH          *handler.WsHandler
	PresignH     *handler.PresignHandler
	AuthMW       gin.HandlerFunc
	InstructorMW gin.HandlerFunc
	AdminMW      gin.HandlerFunc
}

// New construye y devuelve el engine de Gin con todas las rutas registradas.
func New(d Deps) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20

	_ = r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     d.Cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Cabeceras de seguridad
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		if !strings.HasPrefix(c.Request.URL.Path, "/uploads/documents/") {
			c.Header("X-Frame-Options", "SAMEORIGIN")
		}
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	// Frontend estático (SPA)
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		for _, prefix := range []string{"/.git", "/.env", "/actuator"} {
			if strings.HasPrefix(p, prefix) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
		}
		c.File("./frontend/dist/index.html")
	})

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

		// ── Auth (público) ────────────────────────────────────────────────────
		api.POST("/register", d.AuthH.Register)
		api.POST("/login", d.AuthH.Login)
		api.POST("/logout", d.AuthH.Logout)
		api.POST("/forgot-password", d.AuthH.ForgotPassword)
		api.POST("/reset-password", d.AuthH.ResetPassword)

		// ── Público ───────────────────────────────────────────────────────────
		api.GET("/preview-curso/:codigo", d.CursosH.PreviewCurso)
		api.GET("/cursos-publicos", d.CursosH.ListCursosPublicos)
		api.GET("/cursos-publicos/:id", d.CursosH.GetCursoPublico)
		// Stripe webhook — debe ser público y sin AuthMW
		api.POST("/webhooks/stripe", d.CursosH.StripeWebhook)

		// ── Autenticado ───────────────────────────────────────────────────────
		auth := api.Group("/")
		auth.Use(d.AuthMW)
		{
			// Perfil
			auth.GET("/perfil", d.UsuariosH.GetPerfil)
			auth.PUT("/perfil", d.UsuariosH.UpdatePerfil)
			auth.POST("/perfil/avatar", d.UsuariosH.UploadAvatar)
			auth.POST("/perfil/cover", d.UsuariosH.UploadCover)
			auth.POST("/perfil/become-instructor", d.UsuariosH.BecomeInstructor)
			auth.GET("/usuarios/:id/perfil", d.UsuariosH.GetPublicPerfil)
			auth.GET("/usuarios/search", d.UsuariosH.SearchUsers)

			// Notificaciones
			auth.GET("/notificaciones", d.UsuariosH.ListNotificaciones)
			auth.POST("/notificaciones/marcar-leidas", d.UsuariosH.MarcarNotificacionesLeidas)

			// Cursos y capacitaciones
			auth.GET("/mis-capacitaciones", d.CursosH.ListMisCapacitaciones)
			auth.GET("/capacitaciones/:id", d.CursosH.GetCurso)
			auth.POST("/cursos/:id/inscripciones", d.CursosH.Inscribirse)
			auth.POST("/inscripciones", d.CursosH.UnirseConCodigo)
			auth.POST("/inscripciones-licencia", d.CursosH.UnirseConLicencia)
			auth.POST("/checkout-session", d.CursosH.CreateCheckoutSession)
			auth.POST("/checkout-session-b2b-direct", d.CursosH.CreateCheckoutSessionB2BDirect)
			auth.POST("/verify-checkout-session", d.CursosH.VerifyCheckoutSession)
			auth.GET("/licencias-compradas", d.CursosH.ListLicenciasCompradas)
			auth.GET("/licencias/:id/invoice", d.CursosH.GetLicenciaInvoicePDF)

			// Lecciones
			auth.GET("/capacitaciones/:id/lecciones", d.LeccionesH.GetLeccionesConProgreso)
			auth.POST("/lecciones/:leccion_id/completar", d.LeccionesH.MarcarLeccionCompleta)
			auth.GET("/capacitaciones/:id/intermedias", d.LeccionesH.GetPreguntasIntermedias)
			auth.POST("/capacitaciones/:id/intermedias/submit", d.LeccionesH.SubmitPreguntasIntermedias)

			// Exámenes
			auth.GET("/mis-examenes", d.ExamenesH.ListMisExamenes)
			auth.GET("/examenes/:id", d.ExamenesH.GetExamen)
			auth.POST("/examenes/:id/submit", d.ExamenesH.SubmitExamen)

			// Presign
			auth.GET("/presign", d.PresignH.PresignUpload)

			// Foros
			auth.GET("/lecciones/:leccion_id/foro", d.ForosH.ListForoPosts)
			auth.POST("/lecciones/:leccion_id/foro", d.ForosH.CreateForoPost)
			auth.DELETE("/foro/posts/:post_id", d.ForosH.DeleteForoPost)
			auth.GET("/foro/posts/:post_id/comentarios", d.ForosH.ListForoComentarios)
			auth.POST("/foro/posts/:post_id/comentarios", d.ForosH.CreateForoComentario)
			auth.POST("/foro/posts/:post_id/reactions", d.ForosH.ToggleForoPostReaction)
			auth.POST("/foro/comentarios/:comentario_id/reactions", d.ForosH.ToggleForoComentarioReaction)

			// WebSocket tiempo real
			auth.GET("/ws", d.WsH.Handle)

			// Mensajes directos y grupos
			auth.GET("/mensajes/no-leidos", d.MensajesH.NoLeidos)
			auth.GET("/mensajes/conversaciones", d.MensajesH.ListConversaciones)
			auth.POST("/mensajes/leido/:msg_id", d.MensajesH.MarcarLeido)
			auth.POST("/mensajes/grupos", d.MensajesH.CreateGroup)
			auth.POST("/mensajes/grupos/:grupo_id/members", d.MensajesH.AddGroupMembers)
			auth.GET("/mensajes/grupos/:grupo_id/members", d.MensajesH.GetGroupMembers)
			auth.GET("/mensajes/:peer_id", d.MensajesH.GetMensajes)
			auth.POST("/mensajes/:peer_id", d.MensajesH.SendMensaje)

			// ── Instructor ────────────────────────────────────────────────────
			inst := auth.Group("/instructor")
			inst.Use(d.InstructorMW)
			{
				inst.GET("/capacitaciones", d.CursosH.InstructorListCapacitaciones)
				inst.POST("/capacitaciones", d.CursosH.InstructorCreateCapacitacion)
				inst.PUT("/capacitaciones/:id", d.CursosH.InstructorUpdateCapacitacion)
				inst.DELETE("/capacitaciones/:id", d.CursosH.InstructorDeleteCapacitacion)

				inst.GET("/capacitaciones/:id/lecciones", d.LeccionesH.InstructorListLecciones)
				inst.POST("/capacitaciones/:id/lecciones", d.LeccionesH.InstructorCreateLeccion)
				inst.PUT("/capacitaciones/:id/lecciones/:leccion_id", d.LeccionesH.InstructorUpdateLeccion)
				inst.DELETE("/capacitaciones/:id/lecciones/:leccion_id", d.LeccionesH.InstructorDeleteLeccion)
				inst.PUT("/capacitaciones/:id/lecciones/reorder", d.LeccionesH.InstructorReorderLecciones)

				inst.PATCH("/capacitaciones/:id/toggle-public", d.CursosH.InstructorTogglePublic)
				inst.POST("/capacitaciones/:id/reset-codigo", d.CursosH.InstructorResetCodigo)
				inst.GET("/estudiantes", d.CursosH.InstructorListEstudiantes)
				inst.POST("/asignar", d.CursosH.InstructorAsignar)

				inst.GET("/capacitaciones/:id/intermedias", d.LeccionesH.InstructorListPreguntasIntermedias)
				inst.POST("/capacitaciones/:id/intermedias", d.LeccionesH.InstructorCreatePreguntaIntermedia)
				inst.DELETE("/capacitaciones/:id/intermedias/:pregunta_id", d.LeccionesH.InstructorDeletePreguntaIntermedia)

				inst.GET("/examenes", d.ExamenesH.InstructorListExamenes)
				inst.POST("/examenes", d.ExamenesH.InstructorCreateExamen)
				inst.DELETE("/examenes/:id", d.ExamenesH.InstructorDeleteExamen)
				inst.GET("/examenes/:id/resultados", d.ExamenesH.InstructorGetResultados)
				inst.GET("/examenes/:id/resultados/:user_id", d.ExamenesH.InstructorGetRespuestasUsuario)
			}

			// ── Admin ─────────────────────────────────────────────────────────
			adm := auth.Group("/admin")
			adm.Use(d.AdminMW)
			{
				adm.GET("/users", d.UsuariosH.ListUsers)
				adm.POST("/users/:id/revoke-sessions", d.UsuariosH.RevokeUserSessions)

				adm.GET("/capacitaciones", d.CursosH.AdminListCapacitaciones)
				adm.POST("/asignar", d.CursosH.AdminAsignar)
				adm.DELETE("/asignar/:id", d.CursosH.AdminDesAsignar)

				adm.POST("/capacitaciones", d.CursosH.AdminCreateCapacitacion)
				adm.PUT("/capacitaciones/:id", d.CursosH.AdminUpdateCapacitacion)
				adm.DELETE("/capacitaciones/:id", d.CursosH.AdminDeleteCapacitacion)
				adm.GET("/asignaciones", d.CursosH.AdminListAsignaciones)

				adm.GET("/examenes", d.ExamenesH.AdminListExamenes)
				adm.POST("/examenes", d.ExamenesH.AdminCreateExamen)
				adm.DELETE("/examenes/:id", d.ExamenesH.AdminDeleteExamen)
			}
		}
	}

	return r
}
