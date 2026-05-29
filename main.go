package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"Prueba-Go/internal/config"
	"Prueba-Go/internal/db"
	"Prueba-Go/internal/handlers"
	"Prueba-Go/internal/middleware"
	"Prueba-Go/internal/repository"
	"Prueba-Go/internal/service"
	"Prueba-Go/internal/storage"

	sentry "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// ── 1. Configuración centralizada (Fail Fast) ─────────────────────────────
	// config.Load() lee todas las variables de entorno y llama a os.Exit(1)
	// si falta alguna variable crítica (JWT_SECRET, etc.).
	config.Load()

	// ── 2. Logging estructurado ───────────────────────────────────────────────
	logLevel := slog.LevelInfo
	if config.C.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(log.Writer(), &slog.HandlerOptions{
		Level: logLevel,
	})))

	// ── 3. Modo Gin ───────────────────────────────────────────────────────────
	if config.C.GinMode != "" {
		gin.SetMode(config.C.GinMode)
	} else if config.C.RailwayEnvironment != "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── 4. Middleware JWT — secret viene del config ───────────────────────────
	middleware.SetSecret(config.C.JWTSecret)

	// ── 5. Base de datos + migraciones ────────────────────────────────────────
	db.Connect()
	defer db.DB.Close()
	db.Migrate()

	// ── 6. Almacenamiento (R2) ────────────────────────────────────────────────
	storage.Init()

	// ── 5. Sentry ─────────────────────────────────────────────────────────────
	if config.C.SentryDSN != "" {
		env := config.C.RailwayEnvironment
		if env == "" {
			env = "development"
		}
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              config.C.SentryDSN,
			Environment:      env,
			EnableTracing:    true,
			TracesSampleRate: 0.1,
		}); err != nil {
			slog.Warn("Sentry init failed", "error", err)
		} else {
			slog.Info("Sentry inicializado", "environment", env)
			defer sentry.Flush(2 * time.Second)
		}
	}

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MB en memoria; el resto va a disco temporal

	// Railway usa un proxy interno — confiar solo en él para X-Forwarded-For.
	// nil = no confiar en ningún proxy (usar IP directa); ajustar si Railway
	// proporciona IPs de proxy específicas.
	if err := r.SetTrustedProxies(nil); err != nil {
		slog.Warn("SetTrustedProxies falló", "error", err)
	}

	// ALLOWED_ORIGIN ya fue procesado en config.Load()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.C.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true, // necesario para cookies HttpOnly
		MaxAge:           12 * time.Hour,
	}))

	// Captura panics y errores con Sentry (no-op si Sentry no está configurado)
	if config.C.SentryDSN != "" {
		r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	}

	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		if !strings.HasPrefix(c.Request.URL.Path, "/uploads/documents/") {
			c.Header("X-Frame-Options", "SAMEORIGIN")
		}
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		// Bloquear prefijos sensibles — devuelven 404 real
		for _, prefix := range []string{"/.git", "/.env", "/.vite", "/actuator", "/api/"} {
			if strings.HasPrefix(p, prefix) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
		}
		// Todo lo demás es una ruta de la SPA — Vue Router maneja el 404 internamente.
		c.File("./frontend/dist/index.html")
	})

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

		// ── Inyección de dependencias: Repositorio → Servicio → Handler ──────
		// db.DB ya fue conectado arriba — siempre tiene valor aquí.
		userRepo := repository.NewUserRepository(db.DB)
		authSvc := service.NewAuthService(userRepo)
		authHandler := handlers.NewAuthHandler(authSvc)

		loginLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
		registerLimiter := middleware.NewRateLimiter(5, time.Hour)
		forgotLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
		resetLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
		api.POST("/register", registerLimiter.Middleware(), authHandler.Register)
		api.POST("/login", loginLimiter.Middleware(), authHandler.Login)
		api.POST("/logout", handlers.Logout)
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
			auth.POST("/perfil/become-instructor", handlers.BecomeInstructor)
			auth.GET("/usuarios/:id/perfil", handlers.GetPublicPerfil)

			auth.GET("/presign", handlers.PresignUpload)

			auth.GET("/cursos-publicos", handlers.ListCursosPublicos)
			// RESTful: sustantivos + método HTTP indican la acción
			// Antes: POST /inscribirse/:id  →  Ahora: POST /cursos/:id/inscripciones
			// Antes: POST /unirse-con-codigo →  Ahora: POST /inscripciones  (código en body)
			auth.POST("/cursos/:id/inscripciones", handlers.Inscribirse)
			auth.POST("/inscripciones", handlers.UnirseConCodigo)

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
				admin.POST("/users/:id/revoke-sessions", handlers.RevokeUserSessions)
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

				admin.POST("/migrate-local-to-r2", handlers.MigrateLocalToR2)
			}
		}
	}

	srv := &http.Server{
		Addr:    ":" + config.C.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Servidor iniciado en http://localhost:%s", config.C.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error iniciando servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Apagando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Shutdown: %v", err)
	}
	log.Println("Servidor apagado correctamente")
}
