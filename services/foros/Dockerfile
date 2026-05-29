FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.work go.work
COPY gen/go.mod gen/go.mod
COPY services/foros/go.mod services/foros/go.sum* ./services/foros/
WORKDIR /app/services/foros
RUN go mod download
WORKDIR /app
COPY gen/           gen/
COPY services/foros/ services/foros/
WORKDIR /app/services/foros
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /foros-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /foros-service /foros-service
EXPOSE 50056
ENTRYPOINT ["/foros-service"]
