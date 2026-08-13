# ─────────────────────────────────────────────────────────────
# Makefile — Gestión del workspace de microservicios
# ─────────────────────────────────────────────────────────────
# Requisitos para `make generate`:
#   go install github.com/bufbuild/buf/cmd/buf@latest
#
# buf gestiona la generación de código protobuf de forma
# estandarizada sin depender de una instalación global de protoc.
# Los plugins se descargan automáticamente desde buf.build.
# ─────────────────────────────────────────────────────────────

SERVICES := auth usuarios cursos lecciones examenes foros
# Módulos compartidos que no son servicios pero sí forman parte del workspace.
SHARED := pkg/dc3 pkg/mailer pkg/money

# ─────────────────────────────────────────────────────────────
# Toolchain de Go
#
# make ejecuta cada receta con /bin/sh, que NO lee la configuración de tu shell
# interactivo (fish, zsh). Si `go version` funciona al teclearlo pero make dice
# "go: orden no encontrada", esa es la causa.
#
# Se busca en las rutas de instalación habituales; puedes forzarla con:
#   make build GO=/usr/local/go/bin/go
# ─────────────────────────────────────────────────────────────
GO ?= $(shell command -v go 2>/dev/null \
	|| ls /usr/local/go/bin/go 2>/dev/null \
	|| ls $$HOME/go/bin/go 2>/dev/null \
	|| ls $$HOME/.local/go/bin/go 2>/dev/null \
	|| ls /usr/lib/go-*/bin/go 2>/dev/null | sort -V | tail -1)

.PHONY: check-go generate lint-proto breaking tidy build up down logs clean test video-check video-fix jitsi-secrets

## check-go: verifica que el toolchain de Go esté disponible
#
# GOTOOLCHAIN=local es importante: sin él, `go version` lee el `go 1.26.2` de
# go.work y Go intenta descargar ese toolchain. Aquí solo queremos saber si el
# binario existe y corre, no resolver versiones.
check-go:
	@if [ -z "$(GO)" ] || ! GOTOOLCHAIN=local "$(GO)" version >/dev/null 2>&1; then \
		echo "ERROR: no se encontró el compilador de Go."; \
		echo ""; \
		echo "  Si Go NO está instalado:"; \
		echo "    https://go.dev/dl/  ·  o bien:  sudo snap install go --classic"; \
		echo ""; \
		echo "  Si SÍ lo tienes pero make no lo ve (PATH definido solo en fish/zsh),"; \
		echo "  indícale la ruta explícitamente:"; \
		echo "    make $(MAKECMDGOALS) GO=\$$(command -v go)"; \
		echo ""; \
		exit 1; \
	fi
	@echo "==> Usando $$(GOTOOLCHAIN=local "$(GO)" version)"

## generate: genera código Go desde los .proto usando buf
generate:
	@echo "==> Generando código protobuf con buf..."
	buf generate
	@echo "==> Hecho. Ejecuta 'make tidy' para sincronizar dependencias."

## lint-proto: verifica estilo y convenciones de los .proto
lint-proto:
	buf lint

## breaking: detecta cambios incompatibles contra el origen remoto
breaking:
	buf breaking --against '.git#branch=main'

## tidy: ejecuta 'go mod tidy' en todos los módulos del workspace
tidy: check-go
	@for d in . gen $(SHARED) $(addprefix services/,$(SERVICES)) gateway; do \
		echo "==> go mod tidy: $$d"; \
		(cd $$d && "$(GO)" mod tidy) || exit 1; \
	done

## build: compila todos los servicios (requiere haber ejecutado make generate)
build: check-go
	@for svc in $(SERVICES); do \
		echo "==> Building services/$$svc ..."; \
		(cd services/$$svc && "$(GO)" build ./cmd/server/...) || exit 1; \
	done
	@echo "==> Building gateway ..."
	@(cd gateway && "$(GO)" build ./cmd/server/...)

## up: levanta todo con docker compose
up:
	docker compose up --build -d

## down: detiene y elimina los contenedores
down:
	docker compose down

## logs: sigue los logs de todos los servicios
logs:
	docker compose logs -f

## test: ejecuta tests de todos los módulos
test: check-go
	@for d in . $(SHARED) $(addprefix services/,$(SERVICES)) gateway; do \
		echo "==> Testing $$d ..."; \
		(cd $$d && "$(GO)" test ./...) || exit 1; \
	done

## jitsi-secrets: genera los secretos de videollamada y los imprime para .env
##                Nunca los escribe en disco: cópialos tú al .env.
jitsi-secrets:
	@echo "# Pega estas líneas en tu .env (reemplazando las existentes):"
	@echo "JITSI_APP_SECRET=$$(openssl rand -hex 32)"
	@echo "JICOFO_AUTH_PASSWORD=$$(openssl rand -hex 16)"
	@echo "JVB_AUTH_PASSWORD=$$(openssl rand -hex 16)"
	@echo "JICOFO_COMPONENT_SECRET=$$(openssl rand -hex 16)"
	@echo ""
	@echo "# Tras cambiarlos: docker compose down -v jitsi-prosody jitsi-jicofo jitsi-jvb"
	@echo "# (los contenedores cachean la config en sus volúmenes)"

## video-check: diagnostica si un MP4 arranca lento (atom moov al final)
##              uso: make video-check FILE=ruta/o/url.mp4
video-check:
	@test -n "$(FILE)" || { echo "Falta FILE=. Ej: make video-check FILE=curso.mp4"; exit 1; }
	@./scripts/video_faststart.sh check "$(FILE)"

## video-fix: reescribe un MP4 con -movflags +faststart (sin recodificar)
##            uso: make video-fix FILE=entrada.mp4 [OUT=salida.mp4]
video-fix:
	@test -n "$(FILE)" || { echo "Falta FILE=. Ej: make video-fix FILE=curso.mp4"; exit 1; }
	@./scripts/video_faststart.sh fix "$(FILE)" $(OUT)

## clean: elimina el código generado (gen/auth, gen/usuarios, etc.)
clean:
	@rm -rf $(addprefix gen/,$(SERVICES))
	@echo "==> Código generado eliminado. Vuelve a ejecutar 'make generate'."
