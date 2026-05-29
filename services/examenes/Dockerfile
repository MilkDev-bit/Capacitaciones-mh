FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.work go.work
COPY gen/go.mod gen/go.mod
COPY services/examenes/go.mod services/examenes/go.sum* ./services/examenes/
WORKDIR /app/services/examenes
RUN go mod download
WORKDIR /app
COPY gen/              gen/
COPY services/examenes/ services/examenes/
WORKDIR /app/services/examenes
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /examenes-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /examenes-service /examenes-service
EXPOSE 50055
ENTRYPOINT ["/examenes-service"]
