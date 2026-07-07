package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	authpb "Prueba-Go/gen/auth"
	"Prueba-Go/services/auth/internal/config"
	"Prueba-Go/services/auth/internal/handler"
	"Prueba-Go/services/auth/internal/repository"
	"Prueba-Go/services/auth/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

func main() {
	// ── 1. Config (Fail Fast) ─────────────────────────────────────────────────
	cfg := config.Load()

	// ── 2. Logging estructurado ───────────────────────────────────────────────
	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	// ── 3. Base de datos ──────────────────────────────────────────────────────
	db, err := sqlx.Connect("pgx", cfg.DatabaseURL)
	if err != nil {
		slog.Error("DB connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	slog.Info("Database conectada")

	// ── 4. Migraciones automáticas ────────────────────────────────────────────
	if err := runMigrations(db); err != nil {
		slog.Error("Migraciones fallidas", "error", err)
		os.Exit(1)
	}

	// ── 5. Inyección de dependencias: Repo → Service → Handler ───────────────
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, cfg)
	authHandler := handler.NewAuthHandler(authSvc)

	// ── 5. Servidor gRPC ──────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("net.Listen failed", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}

	grpcSrv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              10 * time.Second,
			Timeout:           3 * time.Second,
		}),
		grpc.ChainUnaryInterceptor(loggingInterceptor),
	)
	authpb.RegisterAuthServiceServer(grpcSrv, authHandler)
	reflection.Register(grpcSrv) // permite grpcurl en desarrollo

	slog.Info("auth-service iniciado", "grpc_port", cfg.GRPCPort)

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			slog.Error("gRPC Serve failed", "error", err)
			os.Exit(1)
		}
	}()

	// ── 6. Graceful shutdown ──────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Apagando auth-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = ctx

	grpcSrv.GracefulStop()
	slog.Info("auth-service apagado")
}

// loggingInterceptor registra cada RPC con su duración.
func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	slog.Info("gRPC", "method", info.FullMethod, "duration_ms", time.Since(start).Milliseconds(), "error", err)
	return resp, err
}

// runMigrations crea las tablas y columnas necesarias de forma idempotente.
func runMigrations(db *sqlx.DB) error {
	migrations := []string{
		// Tabla principal de usuarios
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(120) NOT NULL,
			email VARCHAR(200) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role VARCHAR(10) NOT NULL DEFAULT 'user',
			token_version INT NOT NULL DEFAULT 1,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		// Columnas opcionales que pueden no existir en BDs antiguas
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 1`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_token TEXT`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_token_expires TIMESTAMPTZ`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(30) DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS specialty TEXT DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS cover_url TEXT DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS points_total INT NOT NULL DEFAULT 0`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			// Loguear el error completo para diagnóstico
			slog.Error("Migración fallida", "sql", m[:min(60, len(m))], "error", err)
			return fmt.Errorf("migración fallida: %w", err)
		}
	}
	slog.Info("Migraciones aplicadas correctamente")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
