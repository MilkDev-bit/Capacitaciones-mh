FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work go.work
COPY gen/go.mod gen/go.mod
COPY gateway/go.mod gateway/go.sum* ./gateway/

WORKDIR /app/gateway
RUN go mod download

WORKDIR /app
COPY gen/     gen/
COPY gateway/ gateway/
COPY frontend/dist/ frontend/dist/

WORKDIR /app/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gateway ./cmd/server

FROM gcr.io/distroless/static-debian12

COPY --from=builder /gateway /gateway
COPY --from=builder /app/frontend/dist /frontend/dist

EXPOSE 8080

ENTRYPOINT ["/gateway"]
