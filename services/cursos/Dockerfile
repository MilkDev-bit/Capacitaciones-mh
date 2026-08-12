FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY gen/go.mod gen/go.mod
COPY gen/go.sum* gen/
# pkg/money se resuelve por 'replace' local: su go.mod debe existir antes del
# 'go mod download' o la resolución de módulos falla.
COPY pkg/money/go.mod pkg/money/go.sum* ./pkg/money/
COPY services/cursos/go.mod services/cursos/go.sum* ./services/cursos/

WORKDIR /app/services/cursos
RUN GOWORK=off go mod download

WORKDIR /app
COPY gen/             gen/
COPY pkg/money/       pkg/money/
COPY services/cursos/ services/cursos/

WORKDIR /app/services/cursos
RUN GOMAXPROCS=1 GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cursos-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /cursos-service /cursos-service
EXPOSE 50053
ENTRYPOINT ["/cursos-service"]
