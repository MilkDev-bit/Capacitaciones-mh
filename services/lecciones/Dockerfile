FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.work go.work
COPY gen/go.mod gen/go.mod
COPY services/lecciones/go.mod services/lecciones/go.sum* ./services/lecciones/
WORKDIR /app/services/lecciones
RUN go mod download
WORKDIR /app
COPY gen/               gen/
COPY services/lecciones/ services/lecciones/
WORKDIR /app/services/lecciones
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /lecciones-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /lecciones-service /lecciones-service
EXPOSE 50054
ENTRYPOINT ["/lecciones-service"]
