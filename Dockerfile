# ── Etapa 1: Build del frontend Vue ──────────────────────────────────────────
FROM node:22-alpine AS frontend
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci --prefer-offline

COPY frontend/ ./
RUN npm run build-only

# ── Etapa 2: Build del backend Go ────────────────────────────────────────────
FROM golang:1.26-alpine AS backend
WORKDIR /app

# Descargar dependencias primero (aprovechar cache de capas)
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# ── Etapa 3: Imagen final mínima ─────────────────────────────────────────────
FROM alpine:3.21
WORKDIR /app

# Certificados para conexiones TLS (necesario para Railway PostgreSQL)
RUN apk add --no-cache ca-certificates

# Binario Go
COPY --from=backend /app/server ./server

# Frontend compilado (lo sirve el backend desde ./frontend/dist)
COPY --from=frontend /app/frontend/dist ./frontend/dist

# Directorio de uploads (Railway puede montar un volumen aquí)
RUN mkdir -p ./uploads/videos ./uploads/documents

EXPOSE 8080

CMD ["./server"]
