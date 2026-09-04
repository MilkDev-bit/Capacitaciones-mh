# Capacitaciones MH

Plataforma de capacitación en línea para el mercado mexicano: catálogo de cursos,
seguimiento de progreso, exámenes y **emisión de constancias DC-3** conforme al
formato de la Secretaría del Trabajo y Previsión Social (STPS).

Construida sobre microservicios en Go con gRPC y un frontend Vue 3.

[![CI](https://github.com/MilkDev-bit/Capacitaciones-mh/actions/workflows/ci.yml/badge.svg)](https://github.com/MilkDev-bit/Capacitaciones-mh/actions/workflows/ci.yml)

---

## Qué hace

- Catálogo de cursos con compra individual o por suscripción
- Contenido en video, documentos y lecturas, con progreso por lección
- Exámenes con calificación automática
- Foros y mensajería entre participantes
- Actividades con entrega de archivos
- Licencias corporativas: una empresa compra cupos y reparte accesos
- **Constancia DC-3 en PDF**, con folio y verificación pública

### La constancia DC-3

Es la parte más específica del proyecto. La constancia se genera a partir de la
plantilla oficial en `.docx`, se convierte a **PDF** mediante
[Gotenberg](https://gotenberg.dev/) y lleva un **folio verificable** desde una
página pública, sin necesidad de cuenta.

El PDF no es un capricho: en `.docx` la plantilla es editable, y cualquiera podría
emitir constancias a nombre de la organización sin haber cursado nada.

---

## Arquitectura

```
                        ┌──────────────┐
   Navegador  ────────▶ │   Frontend   │   Vue 3 + Vite (SPA, Nginx)
                        └──────┬───────┘
                               │ REST/JSON
                        ┌──────▼───────┐
                        │   Gateway    │   Gin · JWT · webhooks de Stripe
                        └──────┬───────┘
                               │ gRPC
   ┌───────────┬───────────┬───┴───────┬───────────┬───────────┬───────────┐
   ▼           ▼           ▼           ▼           ▼           ▼           ▼
 auth      usuarios     cursos    lecciones   examenes     foros      mensajes
```

El **gateway** es la única pieza expuesta a internet. Traduce REST a gRPC, valida
el JWT y es quien habla con los servicios externos.

**Cada microservicio tiene su propia base de datos.** No hay `JOIN` entre
servicios: cuando un servicio necesita datos que pertenecen a otro, el gateway los
resuelve por gRPC. Es la restricción que más condiciona el diseño de cualquier
funcionalidad nueva.

### Stack

| Capa | Tecnología |
|---|---|
| Backend | Go 1.26 · gRPC · Protocol Buffers ([buf](https://buf.build)) |
| API | Gin (REST hacia el navegador) |
| Frontend | Vue 3 · TypeScript · Pinia · Vite |
| Base de datos | PostgreSQL 16 |
| Pagos | Stripe — tarjeta, OXXO, transferencia SPEI, suscripciones |
| PDF | Gotenberg (LibreOffice headless) |
| Correo | Resend |
| Archivos | Cloudflare R2 |
| Despliegue | Docker · Railway |

---

## Puesta en marcha

Requisitos: Docker y Docker Compose.

```bash
git clone https://github.com/MilkDev-bit/Capacitaciones-mh.git
cd Capacitaciones-mh

cp .env.example .env     # rellena los valores antes de continuar
make up                  # levanta todo el stack
make logs                # sigue los registros
```

- Frontend: <http://localhost>
- API: <http://localhost:8080>

Para parar: `make down`.

La configuración se toma de variables de entorno. [`.env.example`](.env.example)
documenta cada una, cuáles son obligatorias y qué implica activarlas.

---

## Desarrollo

```bash
make build     # compila todos los módulos Go
make test      # tests de Go
make tidy      # sincroniza dependencias del workspace
make clean     # limpia binarios
```

Frontend:

```bash
cd frontend
npm install
npm run dev          # servidor de desarrollo
npm test             # Vitest
npm run type-check   # vue-tsc
```

### Protobuf: el paso que no se puede olvidar

Los contratos gRPC viven en [`proto/`](proto/) y el código generado en `gen/`,
**versionado en el repositorio**. Al tocar cualquier `.proto`:

```bash
make generate    # buf generate
make tidy
git add gen/     # ← imprescindible
```

Olvidarlo produce errores del tipo `unknown field X` repartidos entre varios
servicios, y solo se manifiestan al compilar. Por eso CI ejecuta `buf generate` y
falla si `gen/` no coincide con los `.proto`.

```bash
make lint-proto   # estilo de los contratos
make breaking     # detecta cambios incompatibles
```

### Migraciones

Cada servicio aplica sus migraciones idempotentes al arrancar. Los cambios que
afectan a varios servicios, o que tocan tablas creadas fuera del arranque, viven
en [`migrations/`](migrations/) y se ejecutan a mano. Cada archivo indica **en qué
base** debe correrse y trae consultas de verificación al final.

---

## Tests

```bash
make test                     # Go
cd frontend && npm test       # Vitest
```

CI ejecuta en cada push: sincronización de protos, build y tests de Go, y lint más
tests del frontend.

La mayoría de los tests son **de regresión**: existen porque el fallo ocurrió, y
cada uno documenta en su comentario qué se rompió y por qué. Si vas a tocar el
formateo de importes, el saneado de Markdown o la selección de métodos de pago,
esos comentarios explican el terreno minado.

---

## Estructura

```
├── gateway/          API REST, JWT, webhooks, orquestación gRPC
├── services/         auth · usuarios · cursos · lecciones · examenes · foros · mensajes
├── frontend/         SPA en Vue 3
├── pkg/              módulos compartidos: dc3 · mailer · money
├── proto/            contratos gRPC (fuente de verdad)
├── gen/              código generado por buf (versionado)
├── migrations/       SQL que se aplica a mano
├── gotenberg/        imagen del conversor a PDF, con las fuentes de la DC-3
└── scripts/
```

Un detalle sobre `pkg/money`: **los importes se manejan en centavos, como
enteros**, de extremo a extremo. La división entre 100 ocurre una sola vez, al
formatear para la pantalla. Sumar pesos en coma flotante pierde centavos, y eso se
nota al totalizar un año de ventas.

---

## Licencia

Este repositorio no incluye un archivo de licencia. Sin una licencia explícita,
el código queda bajo derechos de autor reservados y no se concede permiso de uso,
copia ni distribución.

Software desarrollado para **MH Soluciones Empresariales, S.A.S. de C.V.**
