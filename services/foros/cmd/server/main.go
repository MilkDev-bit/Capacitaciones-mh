// foros-service: gestiona posts, comentarios y likes de foros por lección.
package main

import (
	"log/slog"
	"net"
	"os"

	forospb "Prueba-Go/gen/foros"
	"Prueba-Go/services/foros/internal/handler"
	"Prueba-Go/services/foros/internal/repository"
	"Prueba-Go/services/foros/internal/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	db, err := sqlx.Connect("postgres", requireEnv("DATABASE_URL"))
	if err != nil {
		slog.Error("DB", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewForosRepository(db)
	svc := service.NewForosService(repo)
	h := handler.NewForosHandler(svc)

	lis, _ := net.Listen("tcp", ":"+getEnvOr("GRPC_PORT", "50056"))
	srv := grpc.NewServer()
	forospb.RegisterForosServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("foros-service iniciado", "port", getEnvOr("GRPC_PORT", "50056"))
	if err := srv.Serve(lis); err != nil {
		slog.Error("Serve", "error", err)
		os.Exit(1)
	}
}

func requireEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		slog.Error("variable requerida", "key", k)
		os.Exit(1)
	}
	return v
}

func getEnvOr(k, fb string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fb
}
