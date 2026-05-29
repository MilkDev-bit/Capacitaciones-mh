FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar workspace y módulos necesarios
COPY go.work go.work
COPY gen/go.mod gen/go.mod
COPY services/auth/go.mod services/auth/go.sum* ./services/auth/

# Descargar dependencias antes de copiar el código fuente (capa de caché)
WORKDIR /app/services/auth
RUN go mod download

# Copiar el código fuente
WORKDIR /app
COPY gen/ gen/
COPY services/auth/ services/auth/

# Compilar el binario estático
WORKDIR /app/services/auth
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /auth-service ./cmd/server

# ── Imagen final (mínima) ─────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12

COPY --from=builder /auth-service /auth-service

EXPOSE 50051

ENTRYPOINT ["/auth-service"]
