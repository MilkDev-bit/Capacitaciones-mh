# Capacitaciones MH

Panel de capacitaciones con panel admin y panel de usuario.

## Stack

- **Backend**: Go + Gin + PostgreSQL
- **Frontend**: Vue 3 + Pinia + Vue Router

## Requisitos previos

1. PostgreSQL instalado y corriendo
2. Node.js >= 18
3. Go >= 1.21

## Configuración inicial

### 1. Crear la base de datos en PostgreSQL

```sql
CREATE DATABASE capacitaciones;
```

### 2. Configurar variables de entorno

Edita el archivo `.env` en la raíz del proyecto:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=tu_password
DB_NAME=capacitaciones
JWT_SECRET=cambia_este_secreto_minimo_32_chars!!
PORT=8080
```

### 3. Instalar dependencias del frontend

```bash
cd frontend
npm install
```

## Desarrollo (dos terminales)

**Terminal 1 — Backend:**

```bash
# En la raíz del proyecto
$env:DB_PASSWORD="tu_password"; go run main.go
```

**Terminal 2 — Frontend:**

```bash
cd frontend
npm run dev
```

Abre http://localhost:5173

## Producción

```bash
# Compilar frontend
cd frontend && npm run build

# El backend sirve el frontend compilado desde ./frontend/dist
go run main.go
```

Abre http://localhost:8080

## Crear primer usuario administrador

Llama al endpoint de registro con rol admin:

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin","email":"admin@empresa.com","password":"secreto123","role":"admin"}'
```

## Estructura del proyecto

```
├── main.go                    # Punto de entrada, rutas
├── internal/
│   ├── db/                    # Conexión y migraciones PostgreSQL
│   ├── handlers/              # Controladores HTTP
│   ├── middleware/            # JWT auth
│   └── models/                # Estructuras de datos
├── uploads/
│   ├── videos/                # Videos subidos
│   └── documents/             # Documentos PDF subidos
└── frontend/                  # Vue 3 SPA
    └── src/
        ├── views/
        │   ├── admin/         # Panel administrador
        │   └── user/          # Panel usuario
        ├── stores/            # Pinia (auth)
        └── api.ts             # Cliente axios
```

## Funcionalidades

- **Admin**: subir capacitaciones (video/PDF/texto), crear exámenes con preguntas de valor personalizado, asignar contenido a usuarios
- **Usuario**: ver sus capacitaciones asignadas, responder exámenes, ver resultado con puntaje y porcentaje
