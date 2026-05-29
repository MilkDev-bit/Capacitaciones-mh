FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY gen/go.mod gen/go.mod
COPY gen/go.sum* gen/
COPY services/examenes/go.mod services/examenes/go.sum* ./services/examenes/

WORKDIR /app/services/examenes
RUN GOWORK=off go mod download

WORKDIR /app
COPY gen/               gen/
COPY services/examenes/ services/examenes/

WORKDIR /app/services/examenes
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /examenes-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /examenes-service /examenes-service
EXPOSE 50055
ENTRYPOINT ["/examenes-service"]
