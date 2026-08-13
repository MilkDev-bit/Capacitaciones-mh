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
# pkg/mailer y pkg/dc3 son módulos locales referenciados por 'replace': sus
# go.mod deben existir antes del 'go mod download' o la resolución falla.
#
# El build corre con GOWORK=off, así que go.work NO los une aquí: lo único que
# los enlaza son los 'replace' del go.mod del gateway. Añadir un módulo local
# nuevo obliga siempre a tocar los dos sitios, este y el go.mod.
COPY pkg/mailer/go.mod pkg/mailer/go.sum* ./pkg/mailer/
COPY pkg/dc3/go.mod    pkg/dc3/go.sum*    ./pkg/dc3/
COPY gateway/go.mod gateway/go.sum* ./gateway/

WORKDIR /app/gateway
RUN GOWORK=off go mod download && GOWORK=off go mod tidy

WORKDIR /app
COPY gen/         gen/
COPY pkg/mailer/  pkg/mailer/
# Incluye plantilla/dc3.docx, que va embebida en el binario con go:embed.
COPY pkg/dc3/     pkg/dc3/
COPY gateway/     gateway/

WORKDIR /app/gateway
RUN GOMAXPROCS=1 GOWORK=off CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gateway ./cmd/server

# ── Stage 3: Imagen final mínima ─────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12

COPY --from=builder        /gateway          /gateway
COPY --from=frontend-builder /app/frontend/dist /frontend/dist

EXPOSE 8080

ENTRYPOINT ["/gateway"]
