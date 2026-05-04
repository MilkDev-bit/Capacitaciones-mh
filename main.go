package main

import (
	"log"
	"net/http"
	"os"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/handlers"
	"Prueba-Go/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
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
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)
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
				instructor.DELETE("/capacitaciones/:id", handlers.InstructorDeleteCapacitacion)
				instructor.PATCH("/capacitaciones/:id/toggle-public", handlers.InstructorTogglePublic)
				instructor.POST("/capacitaciones/:id/reset-codigo", handlers.InstructorResetCodigo)

				instructor.GET("/examenes", handlers.InstructorListExamenes)
				instructor.POST("/examenes", handlers.InstructorCreateExamen)
				instructor.DELETE("/examenes/:id", handlers.InstructorDeleteExamen)

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
