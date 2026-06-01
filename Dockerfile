# syntax=docker/dockerfile:1
FROM node:22-alpine AS frontend
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci --prefer-offline

COPY frontend/ ./
# Las variables VITE_* se inyectan como secrets de BuildKit en tiempo de build.
# Los valores NUNCA quedan almacenados en capas de imagen.
RUN --mount=type=secret,id=VITE_RECAPTCHA_SITE_KEY \
    --mount=type=secret,id=VITE_SENTRY_DSN \
    VITE_RECAPTCHA_SITE_KEY="$(cat /run/secrets/VITE_RECAPTCHA_SITE_KEY 2>/dev/null || true)" \
    VITE_SENTRY_DSN="$(cat /run/secrets/VITE_SENTRY_DSN 2>/dev/null || true)" \
    npm run build-only

FROM golang:1.26-alpine AS backend
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

FROM alpine:3.21
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=backend /app/server ./server

COPY --from=frontend /app/frontend/dist ./frontend/dist

RUN mkdir -p ./uploads/videos ./uploads/documents

EXPOSE 8080

CMD ["./server"]