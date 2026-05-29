package main

import (
	"context"
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

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		slog.Error("DB connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	slog.Info("Database conectada")

	// ── 4. Inyección de dependencias: Repo → Service → Handler ───────────────
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
