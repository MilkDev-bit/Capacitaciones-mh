package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/config"
	"Prueba-Go/gateway/internal/handler"
	"Prueba-Go/gateway/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// ── 1. Config (Fail Fast) ─────────────────────────────────────────────────
	cfg := config.Load()

	// ── 2. Logging ────────────────────────────────────────────────────────────
	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	// ── 3. Modo Gin ───────────────────────────────────────────────────────────
	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	} else if cfg.RailwayEnvironment != "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── 4. Conexiones gRPC a los microservicios ───────────────────────────────
	svc, err := clients.Dial(cfg)
	if err != nil {
		slog.Error("Error conectando a microservicios", "error", err)
		os.Exit(1)
	}
	defer svc.Close()
	slog.Info("Conexiones gRPC establecidas")

	// ── 5. Handlers HTTP (DI: clients → handler) ──────────────────────────────
	authH := handler.NewAuthHandler(svc, cfg)
	usuariosH := handler.NewUsuariosHandler(svc)
	cursosH := handler.NewCursosHandler(svc)
	leccionesH := handler.NewLeccionesHandler(svc)
	examenesH := handler.NewExamenesHandler(svc)
	forosH := handler.NewForosHandler(svc)

	// ── 6. Middlewares de autenticación ───────────────────────────────────────
	authMW := middleware.AuthRequired(svc)
	instructorMW := middleware.InstructorRequired(svc)
	adminMW := middleware.AdminRequired(svc)

	// ── 7. Router ─────────────────────────────────────────────────────────────
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20

	if err := r.SetTrustedProxies(nil); err != nil {
		slog.Warn("SetTrustedProxies", "error", err)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
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

	// Servir frontend compilado
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
		api.POST("/register", authH.Register)
		api.POST("/login", authH.Login)
		api.POST("/logout", authH.Logout)
		api.POST("/forgot-password", authH.ForgotPassword)
		api.POST("/reset-password", authH.ResetPassword)

		// ── Público ───────────────────────────────────────────────────────────
		api.GET("/preview-curso/:codigo", cursosH.PreviewCurso)
		api.GET("/cursos-publicos", cursosH.ListCursosPublicos)

		// ── Autenticado ───────────────────────────────────────────────────────
		auth := api.Group("/")
		auth.Use(authMW)
		{
			// Perfil
			auth.GET("/perfil", usuariosH.GetPerfil)
			auth.PUT("/perfil", usuariosH.UpdatePerfil)
			auth.POST("/perfil/avatar", usuariosH.UploadAvatar)
			auth.POST("/perfil/cover", usuariosH.UploadCover)
			auth.POST("/perfil/become-instructor", usuariosH.BecomeInstructor)
			auth.GET("/usuarios/:id/perfil", usuariosH.GetPublicPerfil)

			// Cursos
			auth.GET("/mis-capacitaciones", cursosH.ListMisCapacitaciones)
			auth.GET("/capacitaciones/:id", cursosH.GetCurso)
			auth.POST("/cursos/:id/inscripciones", cursosH.Inscribirse)
			auth.POST("/inscripciones", cursosH.UnirseConCodigo)

			// Lecciones
			auth.GET("/capacitaciones/:id/lecciones", leccionesH.GetLeccionesConProgreso)
			auth.POST("/lecciones/:leccion_id/completar", leccionesH.MarcarLeccionCompleta)
			auth.GET("/capacitaciones/:id/intermedias", leccionesH.GetPreguntasIntermedias)
			auth.POST("/capacitaciones/:id/intermedias/submit", leccionesH.SubmitPreguntasIntermedias)

			// Exámenes
			auth.GET("/mis-examenes", examenesH.ListMisExamenes)
			auth.GET("/examenes/:id", examenesH.GetExamen)
			auth.POST("/examenes/:id/submit", examenesH.SubmitExamen)

			// Foros
			auth.GET("/lecciones/:leccion_id/foro", forosH.ListForoPosts)
			auth.POST("/lecciones/:leccion_id/foro", forosH.CreateForoPost)
			auth.DELETE("/foro/posts/:post_id", forosH.DeleteForoPost)
			auth.GET("/foro/posts/:post_id/comentarios", forosH.ListForoComentarios)
			auth.POST("/foro/posts/:post_id/comentarios", forosH.CreateForoComentario)
			auth.POST("/foro/posts/:post_id/like", forosH.ToggleForoPostLike)

			// ── Instructor ────────────────────────────────────────────────────
			inst := auth.Group("/instructor")
			inst.Use(instructorMW)
			{
				inst.GET("/capacitaciones", cursosH.InstructorListCapacitaciones)
				inst.POST("/capacitaciones", cursosH.InstructorCreateCapacitacion)
				inst.PUT("/capacitaciones/:id", cursosH.InstructorUpdateCapacitacion)
				inst.DELETE("/capacitaciones/:id", cursosH.InstructorDeleteCapacitacion)

				inst.GET("/capacitaciones/:id/lecciones", leccionesH.InstructorListLecciones)
				inst.POST("/capacitaciones/:id/lecciones", leccionesH.InstructorCreateLeccion)
				inst.PUT("/capacitaciones/:id/lecciones/:leccion_id", leccionesH.InstructorUpdateLeccion)
				inst.DELETE("/capacitaciones/:id/lecciones/:leccion_id", leccionesH.InstructorDeleteLeccion)
				inst.PUT("/capacitaciones/:id/lecciones/reorder", leccionesH.InstructorReorderLecciones)

				inst.GET("/examenes", examenesH.InstructorListExamenes)
				inst.DELETE("/examenes/:id", examenesH.InstructorDeleteExamen)
				inst.GET("/examenes/:id/resultados", examenesH.InstructorGetResultados)
			}

			// ── Admin ─────────────────────────────────────────────────────────
			adm := auth.Group("/admin")
			adm.Use(adminMW)
			{
				adm.GET("/users", usuariosH.ListUsers)
				adm.POST("/users/:id/revoke-sessions", usuariosH.RevokeUserSessions)

				adm.GET("/capacitaciones", cursosH.AdminListCapacitaciones)
				adm.POST("/asignar", cursosH.AdminAsignar)
				adm.DELETE("/asignar/:id", cursosH.AdminDesAsignar)

				adm.GET("/examenes", examenesH.AdminListExamenes)
				adm.DELETE("/examenes/:id", examenesH.AdminDeleteExamen)
			}
		}
	}

	// ── 8. Servidor HTTP ──────────────────────────────────────────────────────
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		slog.Info("Gateway iniciado", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Apagando gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Shutdown", "error", err)
	}
	slog.Info("Gateway apagado")
}
