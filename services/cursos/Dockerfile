FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY gen/go.mod gen/go.mod
COPY gen/go.sum* gen/
COPY services/cursos/go.mod services/cursos/go.sum* ./services/cursos/

WORKDIR /app/services/cursos
RUN GOWORK=off go mod download

WORKDIR /app
COPY gen/             gen/
COPY services/cursos/ services/cursos/

WORKDIR /app/services/cursos
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cursos-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /cursos-service /cursos-service
EXPOSE 50053
ENTRYPOINT ["/cursos-service"]
