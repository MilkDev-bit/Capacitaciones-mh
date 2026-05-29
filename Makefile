# ─────────────────────────────────────────────────────────────
# Makefile — Gestión del workspace de microservicios
# ─────────────────────────────────────────────────────────────
# Requisitos para `make generate`:
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#   sudo apt install -y protobuf-compiler  (Linux/WSL)
#   brew install protobuf                  (macOS)
#   choco install protoc                   (Windows)
# ─────────────────────────────────────────────────────────────

PROTO_FILES := $(shell find proto -name '*.proto')
SERVICES    := auth usuarios cursos lecciones examenes foros

.PHONY: generate tidy build up down logs clean test

## generate: compila los .proto → gen/**/*.pb.go y gen/**/*_grpc.pb.go
generate:
	@echo "==> Generando código protobuf..."
	@protoc \
		--go_out=gen     --go_opt=module=Prueba-Go/gen \
		--go-grpc_out=gen --go-grpc_opt=module=Prueba-Go/gen \
		-I proto \
		$(PROTO_FILES)
	@echo "==> Hecho. Ejecuta 'make tidy' para sincronizar dependencias."

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
