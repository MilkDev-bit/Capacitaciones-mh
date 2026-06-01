package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/config"
	"Prueba-Go/gateway/internal/handler"
	"Prueba-Go/gateway/internal/hub"
	"Prueba-Go/gateway/internal/middleware"
	"Prueba-Go/gateway/internal/router"
	"Prueba-Go/gateway/internal/storage"

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
	slog.Info(fmt.Sprintf("gRPC addrs: auth=%s usuarios=%s cursos=%s lecciones=%s examenes=%s foros=%s mensajes=%s",
		cfg.AuthAddr, cfg.UsuariosAddr, cfg.CursosAddr, cfg.LeccionesAddr, cfg.ExamenesAddr, cfg.ForosAddr, cfg.MensajesAddr,
	))
	svc, err := clients.Dial(cfg)
	if err != nil {
		slog.Error("Error conectando a microservicios", "error", err)
		os.Exit(1)
	}
	defer svc.Close()
	slog.Info("Conexiones gRPC establecidas")

	// ── 4b. Hub WebSocket ─────────────────────────────────────────────────────
	h := hub.New()

	// ── 4c. Storage R2 ───────────────────────────────────────────────────────
	storage.Init(cfg)

	// ── 5. Inyección de dependencias ──────────────────────────────────────────
	r := router.New(router.Deps{
		Cfg:          cfg,
		AuthH:        handler.NewAuthHandler(svc, cfg),
		UsuariosH:    handler.NewUsuariosHandler(svc),
		CursosH:      handler.NewCursosHandler(svc),
		LeccionesH:   handler.NewLeccionesHandler(svc),
		ExamenesH:    handler.NewExamenesHandler(svc),
		ForosH:       handler.NewForosHandler(svc),
		MensajesH:    handler.NewMensajesHandler(svc, h),
		WsH:          handler.NewWsHandler(h),
		PresignH:     handler.NewPresignHandler(),
		AuthMW:       middleware.AuthRequired(svc),
		InstructorMW: middleware.InstructorRequired(svc),
		AdminMW:      middleware.AdminRequired(svc),
	})

	// ── 6. Servidor HTTP con graceful shutdown ────────────────────────────────
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
