FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY gen/go.mod gen/go.mod
COPY gen/go.sum* gen/
COPY services/usuarios/go.mod services/usuarios/go.sum* ./services/usuarios/

WORKDIR /app/services/usuarios
RUN GOWORK=off go mod download

WORKDIR /app
COPY gen/               gen/
COPY services/usuarios/ services/usuarios/

WORKDIR /app/services/usuarios
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /usuarios-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /usuarios-service /usuarios-service
EXPOSE 50052
ENTRYPOINT ["/usuarios-service"]
