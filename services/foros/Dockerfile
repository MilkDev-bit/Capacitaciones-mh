FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY gen/go.mod gen/go.mod
COPY gen/go.sum* gen/
COPY services/foros/go.mod services/foros/go.sum* ./services/foros/

WORKDIR /app/services/foros
RUN GOWORK=off go mod download

WORKDIR /app
COPY gen/            gen/
COPY services/foros/ services/foros/

WORKDIR /app/services/foros
RUN GOMAXPROCS=1 GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /foros-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /foros-service /foros-service
EXPOSE 50056
ENTRYPOINT ["/foros-service"]
