package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/handlers"
	"Prueba-Go/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Asegurar que existan los subdirectorios de uploads (necesario en Railway con volumen vacío)
	for _, dir := range []string{
		"uploads/videos", "uploads/documents",
		"uploads/thumbnails", "uploads/avatars", "uploads/covers",
		"uploads/foro",
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("warning: no se pudo crear %s: %v", dir, err)
		}
	}

	// A02: modo release en producción para no exponer rutas/stack traces
	if os.Getenv("GIN_MODE") != "" {
		gin.SetMode(os.Getenv("GIN_MODE"))
	} else if os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// A04/A07: verificar JWT_SECRET al inicio
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("[SECURITY] JWT_SECRET no definido — la clave de firma es conocida. Configura esta variable en producción.")
	}

	r := gin.Default()

	// A02: CORS restringido al origen permitido (no wildcard)
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173" // solo desarrollo
	}
	r.Use(func(c *gin.Context) {
		// A02: security headers
		c.Header("X-Content-Type-Options", "nosniff")
		// Necesario para permitir previsualizaciones internas (PDF/documentos) en iframes del mismo origen.
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// CORS
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Servir archivos subidos
	r.Static("/uploads", "./uploads")
	// Servir frontend Vue compilado
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	api := r.Group("/api")
	{
		// Health check (sin auth, para Railway)
		api.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

		// Auth
		// A07: rate limiting en endpoints de autenticación
		loginLimiter := middleware.NewRateLimiter(10, 15*time.Minute) // 10 intentos / 15 min por IP
		registerLimiter := middleware.NewRateLimiter(5, time.Hour)    // 5 registros / hora por IP
		api.POST("/register", registerLimiter.Middleware(), handlers.Register)
		api.POST("/login", loginLimiter.Middleware(), handlers.Login)
		// Preview público de curso por código (sin auth, para página de invitación)
		api.GET("/preview-curso/:codigo", handlers.PreviewCurso)

		auth := api.Group("/")
		auth.Use(middleware.AuthRequired())
		{
			// Usuario: sus capacitaciones y exámenes asignados
			auth.GET("/mis-capacitaciones", handlers.ListCapacitacionesUsuario)
			auth.GET("/mis-examenes", handlers.ListExamenesUsuario)
			auth.GET("/examenes/:id", handlers.GetExamen)
			auth.POST("/examenes/:id/submit", handlers.SubmitExamen)
			auth.GET("/capacitaciones/:id", handlers.GetCapacitacion)

			// Lecciones con progreso
			auth.GET("/capacitaciones/:id/lecciones", handlers.GetLeccionesConProgreso)
			auth.POST("/lecciones/:leccion_id/completar", handlers.MarcarLeccionCompleta)

			// Preguntas intermedias
			auth.GET("/capacitaciones/:id/intermedias", handlers.GetPreguntasIntermedias)
			auth.POST("/capacitaciones/:id/intermedias/submit", handlers.SubmitPreguntasIntermedias)

			// Foros
			auth.GET("/lecciones/:leccion_id/foro", handlers.ListForoPosts)
			auth.POST("/lecciones/:leccion_id/foro", handlers.CreateForoPost)
			auth.DELETE("/foro/posts/:post_id", handlers.DeleteForoPost)
			auth.GET("/foro/posts/:post_id/comentarios", handlers.ListForoComentarios)
			auth.POST("/foro/posts/:post_id/comentarios", handlers.CreateForoComentario)
			auth.POST("/foro/posts/:post_id/like", handlers.ToggleForoPostLike)

			// Perfil
			auth.GET("/perfil", handlers.GetPerfil)
			auth.PUT("/perfil", handlers.UpdatePerfil)
			auth.POST("/perfil/avatar", handlers.UploadAvatar)
			auth.POST("/perfil/cover", handlers.UploadCover)
			auth.GET("/usuarios/:id/perfil", handlers.GetPublicPerfil)

			// Cursos públicos (cualquier usuario autenticado)
			auth.GET("/cursos-publicos", handlers.ListCursosPublicos)
			auth.POST("/inscribirse/:id", handlers.Inscribirse)
			auth.POST("/unirse-con-codigo", handlers.UnirseConCodigo)

			// Instructor
			instructor := auth.Group("/instructor")
			instructor.Use(middleware.InstructorRequired())
			{
				instructor.GET("/capacitaciones", handlers.InstructorListCapacitaciones)
				instructor.POST("/capacitaciones", handlers.InstructorCreateCapacitacion)
				instructor.PUT("/capacitaciones/:id", handlers.InstructorUpdateCapacitacion)
				instructor.DELETE("/capacitaciones/:id", handlers.InstructorDeleteCapacitacion)
				instructor.PATCH("/capacitaciones/:id/toggle-public", handlers.InstructorTogglePublic)
				instructor.POST("/capacitaciones/:id/reset-codigo", handlers.InstructorResetCodigo)

				// Lecciones de un curso
				instructor.GET("/capacitaciones/:id/lecciones", handlers.InstructorListLecciones)
				instructor.POST("/capacitaciones/:id/lecciones", handlers.InstructorCreateLeccion)
				instructor.PUT("/capacitaciones/:id/lecciones/:leccion_id", handlers.InstructorUpdateLeccion)
				instructor.DELETE("/capacitaciones/:id/lecciones/:leccion_id", handlers.InstructorDeleteLeccion)
				instructor.PUT("/capacitaciones/:id/lecciones/reorder", handlers.InstructorReorderLecciones)

				// Preguntas intermedias de un curso
				instructor.GET("/capacitaciones/:id/intermedias", handlers.InstructorListPreguntasIntermedias)
				instructor.POST("/capacitaciones/:id/intermedias", handlers.InstructorCreatePreguntaIntermedia)
				instructor.DELETE("/capacitaciones/:id/intermedias/:pregunta_id", handlers.InstructorDeletePreguntaIntermedia)

				instructor.GET("/examenes", handlers.InstructorListExamenes)
				instructor.POST("/examenes", handlers.InstructorCreateExamen)
				instructor.DELETE("/examenes/:id", handlers.InstructorDeleteExamen)
				instructor.GET("/examenes/:id/resultados", handlers.InstructorGetResultados)
				instructor.GET("/examenes/:id/resultados/:user_id", handlers.InstructorGetRespuestasUsuario)

				instructor.GET("/estudiantes", handlers.InstructorListEstudiantes)
				instructor.POST("/asignar", handlers.InstructorAsignar)

				// Listar todos los usuarios para poder asignar
				instructor.GET("/users", handlers.ListUsers)
			}

			// Admin
			admin := auth.Group("/admin")
			admin.Use(middleware.AdminRequired())
			{
				admin.GET("/users", handlers.ListUsers)
				admin.POST("/asignar", handlers.Asignar)
				admin.DELETE("/asignar/:id", handlers.DesAsignar)
				admin.GET("/asignaciones", handlers.ListAsignaciones)

				admin.GET("/capacitaciones", handlers.ListCapacitaciones)
				admin.POST("/capacitaciones", handlers.CreateCapacitacion)
				admin.PUT("/capacitaciones/:id", handlers.UpdateCapacitacion)
				admin.DELETE("/capacitaciones/:id", handlers.DeleteCapacitacion)

				admin.GET("/examenes", handlers.ListExamenes)
				admin.POST("/examenes", handlers.CreateExamen)
				admin.DELETE("/examenes/:id", handlers.DeleteExamen)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor iniciado en http://localhost:%s", port)

	// Arrancar el servidor HTTP en segundo plano para que el healthcheck
	// de Railway pueda responder mientras la BD termina de conectar.
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Error en r.Run: %v", err)
		}
	}()

	// Conectar la base de datos (reintenta hasta 10 veces × 3 s = 30 s)
	db.Connect()
	db.Migrate()

	// Mantener el proceso vivo
	select {}
}
