# ── Stage 1: Build frontend ───────────────────────────────────────────────────
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci --prefer-offline

COPY frontend/ .

# VITE_RECAPTCHA_SITE_KEY se inyecta en tiempo de build via build arg
ARG VITE_RECAPTCHA_SITE_KEY=""
ENV VITE_RECAPTCHA_SITE_KEY=$VITE_RECAPTCHA_SITE_KEY

RUN npm run build-only

# ── Stage 2: Build gateway (Go) ───────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY gen/go.mod gen/go.mod
COPY gen/go.sum* gen/
COPY gateway/go.mod gateway/go.sum* ./gateway/

WORKDIR /app/gateway
RUN GOWORK=off go mod download

WORKDIR /app
COPY gen/     gen/
COPY gateway/ gateway/

WORKDIR /app/gateway
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gateway ./cmd/server

# ── Stage 3: Imagen final mínima ─────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12

COPY --from=builder        /gateway          /gateway
COPY --from=frontend-builder /app/frontend/dist /frontend/dist

EXPOSE 8080

ENTRYPOINT ["/gateway"]
