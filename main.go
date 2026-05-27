package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/handlers"
	"Prueba-Go/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	for _, dir := range []string{
		"uploads/videos", "uploads/documents",
		"uploads/thumbnails", "uploads/avatars", "uploads/covers",
		"uploads/foro",
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("warning: no se pudo crear %s: %v", dir, err)
		}
	}

	if os.Getenv("GIN_MODE") != "" {
		gin.SetMode(os.Getenv("GIN_MODE"))
	} else if os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		gin.SetMode(gin.ReleaseMode)
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("[SECURITY] JWT_SECRET no definido — configura esta variable antes de arrancar")
	}

	r := gin.Default()

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{allowedOrigin},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:       12 * time.Hour,
	}))

	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		if !strings.HasPrefix(c.Request.URL.Path, "/uploads/documents/") {
			c.Header("X-Frame-Options", "SAMEORIGIN")
		}
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	r.Static("/uploads", "./uploads")
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

		loginLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
		registerLimiter := middleware.NewRateLimiter(5, time.Hour)
		forgotLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
		resetLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
		api.POST("/register", registerLimiter.Middleware(), handlers.Register)
		api.POST("/login", loginLimiter.Middleware(), handlers.Login)
		api.POST("/forgot-password", forgotLimiter.Middleware(), handlers.ForgotPassword)
		api.POST("/reset-password", resetLimiter.Middleware(), handlers.ResetPassword)
		api.GET("/preview-curso/:codigo", handlers.PreviewCurso)

		auth := api.Group("/")
		auth.Use(middleware.AuthRequired())
		{
			auth.GET("/mis-capacitaciones", handlers.ListCapacitacionesUsuario)
			auth.GET("/mis-examenes", handlers.ListExamenesUsuario)
			auth.GET("/examenes/:id", handlers.GetExamen)
			auth.POST("/examenes/:id/submit", handlers.SubmitExamen)
			auth.GET("/capacitaciones/:id", handlers.GetCapacitacion)

			auth.GET("/capacitaciones/:id/lecciones", handlers.GetLeccionesConProgreso)
			auth.POST("/lecciones/:leccion_id/completar", handlers.MarcarLeccionCompleta)

			auth.GET("/capacitaciones/:id/intermedias", handlers.GetPreguntasIntermedias)
			auth.POST("/capacitaciones/:id/intermedias/submit", handlers.SubmitPreguntasIntermedias)

			auth.GET("/lecciones/:leccion_id/foro", handlers.ListForoPosts)
			auth.POST("/lecciones/:leccion_id/foro", handlers.CreateForoPost)
			auth.DELETE("/foro/posts/:post_id", handlers.DeleteForoPost)
			auth.GET("/foro/posts/:post_id/comentarios", handlers.ListForoComentarios)
			auth.POST("/foro/posts/:post_id/comentarios", handlers.CreateForoComentario)
			auth.POST("/foro/posts/:post_id/like", handlers.ToggleForoPostLike)

			auth.GET("/perfil", handlers.GetPerfil)
			auth.PUT("/perfil", handlers.UpdatePerfil)
			auth.POST("/perfil/avatar", handlers.UploadAvatar)
			auth.POST("/perfil/cover", handlers.UploadCover)
			auth.GET("/usuarios/:id/perfil", handlers.GetPublicPerfil)

			auth.GET("/cursos-publicos", handlers.ListCursosPublicos)
			auth.POST("/inscribirse/:id", handlers.Inscribirse)
			auth.POST("/unirse-con-codigo", handlers.UnirseConCodigo)

			instructor := auth.Group("/instructor")
			instructor.Use(middleware.InstructorRequired())
			{
				instructor.GET("/capacitaciones", handlers.InstructorListCapacitaciones)
				instructor.POST("/capacitaciones", handlers.InstructorCreateCapacitacion)
				instructor.PUT("/capacitaciones/:id", handlers.InstructorUpdateCapacitacion)
				instructor.DELETE("/capacitaciones/:id", handlers.InstructorDeleteCapacitacion)
				instructor.PATCH("/capacitaciones/:id/toggle-public", handlers.InstructorTogglePublic)
				instructor.POST("/capacitaciones/:id/reset-codigo", handlers.InstructorResetCodigo)

				instructor.GET("/capacitaciones/:id/lecciones", handlers.InstructorListLecciones)
				instructor.POST("/capacitaciones/:id/lecciones", handlers.InstructorCreateLeccion)
				instructor.PUT("/capacitaciones/:id/lecciones/:leccion_id", handlers.InstructorUpdateLeccion)
				instructor.DELETE("/capacitaciones/:id/lecciones/:leccion_id", handlers.InstructorDeleteLeccion)
				instructor.PUT("/capacitaciones/:id/lecciones/reorder", handlers.InstructorReorderLecciones)

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

				instructor.GET("/users", handlers.ListUsers)
			}

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

	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Error en r.Run: %v", err)
		}
	}()

	db.Connect()
	db.Migrate()

	select {}
}
