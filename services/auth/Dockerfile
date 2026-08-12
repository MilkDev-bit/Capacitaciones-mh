FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY gen/go.mod gen/go.mod
COPY gen/go.sum* gen/
# pkg/mailer es un módulo local referenciado por 'replace': su go.mod debe
# existir antes del 'go mod download' o la resolución falla.
COPY pkg/mailer/go.mod pkg/mailer/go.sum* ./pkg/mailer/
COPY services/auth/go.mod services/auth/go.sum* ./services/auth/

WORKDIR /app/services/auth
RUN GOWORK=off go mod download

WORKDIR /app
COPY gen/ gen/
COPY pkg/mailer/ pkg/mailer/
COPY services/auth/ services/auth/

WORKDIR /app/services/auth
RUN GOMAXPROCS=1 GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /auth-service ./cmd/server

# ── Imagen final (mínima) ─────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12

COPY --from=builder /auth-service /auth-service

EXPOSE 50051

ENTRYPOINT ["/auth-service"]
