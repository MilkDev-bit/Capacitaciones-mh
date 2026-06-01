FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY gen/go.mod gen/go.mod
COPY gen/go.sum* gen/
COPY services/mensajes/go.mod services/mensajes/go.sum* ./services/mensajes/

WORKDIR /app/services/mensajes
RUN GOWORK=off go mod download

WORKDIR /app
COPY gen/              gen/
COPY services/mensajes/ services/mensajes/

WORKDIR /app/services/mensajes
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /mensajes-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /mensajes-service /mensajes-service
EXPOSE 50057
ENTRYPOINT ["/mensajes-service"]
