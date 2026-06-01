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

.PHONY: generate lint-proto breaking tidy build up down logs clean test

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
tidy:
	@for d in . gen $(addprefix services/,$(SERVICES)) gateway; do \
		echo "==> go mod tidy: $$d"; \
		(cd $$d && go mod tidy); \
	done

## build: compila todos los servicios (requiere haber ejecutado make generate)
build:
	@for svc in $(SERVICES); do \
		echo "==> Building services/$$svc ..."; \
		(cd services/$$svc && go build ./cmd/server/...); \
	done
	@echo "==> Building gateway ..."
	@(cd gateway && go build ./cmd/server/...)

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
test:
	@for d in . $(addprefix services/,$(SERVICES)) gateway; do \
		echo "==> Testing $$d ..."; \
		(cd $$d && go test ./...); \
	done

## clean: elimina el código generado (gen/auth, gen/usuarios, etc.)
clean:
	@rm -rf $(addprefix gen/,$(SERVICES))
	@echo "==> Código generado eliminado. Vuelve a ejecutar 'make generate'."
