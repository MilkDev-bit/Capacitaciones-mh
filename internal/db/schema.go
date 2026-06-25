package db

import (
	"log/slog"
	"os"
)

const Schema = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    email VARCHAR(200) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(10) NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS capacitaciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL,
    file_path TEXT,
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS examenes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS preguntas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    examen_id UUID NOT NULL REFERENCES examenes(id) ON DELETE CASCADE,
    texto TEXT NOT NULL,
    valor NUMERIC(5,2) NOT NULL DEFAULT 1,
    orden INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS opciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pregunta_id UUID NOT NULL REFERENCES preguntas(id) ON DELETE CASCADE,
    texto TEXT NOT NULL,
    es_correcta BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS asignaciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capacitacion_id UUID REFERENCES capacitaciones(id) ON DELETE CASCADE,
    examen_id UUID REFERENCES examenes(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, capacitacion_id),
    CHECK (capacitacion_id IS NOT NULL OR examen_id IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS respuestas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    examen_id UUID NOT NULL REFERENCES examenes(id) ON DELETE CASCADE,
    pregunta_id UUID NOT NULL REFERENCES preguntas(id) ON DELETE CASCADE,
    opcion_id UUID NOT NULL REFERENCES opciones(id) ON DELETE CASCADE,
    respondido_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, examen_id, pregunta_id)
);

CREATE TABLE IF NOT EXISTS curso_licencias (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
    nombre VARCHAR(100) NOT NULL,
    precio NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    capacidad_maxima INT NOT NULL DEFAULT 0,
    usadas INT NOT NULL DEFAULT 0,
    codigo_acceso VARCHAR(50) UNIQUE,
    stripe_product_id VARCHAR(100),
    stripe_price_id VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inscripciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
    licencia_id UUID REFERENCES curso_licencias(id) ON DELETE SET NULL,
    inscrito_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, capacitacion_id)
);

ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS instructor_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS codigo_acceso VARCHAR(12) UNIQUE;
ALTER TABLE examenes ADD COLUMN IF NOT EXISTS instructor_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Perfil de usuario
ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(30) DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS specialty TEXT DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS cover_url TEXT DEFAULT '';

-- Lecciones dentro de un curso (contenido estructurado en serie)
CREATE TABLE IF NOT EXISTS lecciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    type VARCHAR(20) NOT NULL DEFAULT 'video',
    file_path TEXT DEFAULT '',
    content TEXT DEFAULT '',
    orden INT NOT NULL DEFAULT 0,
    duracion_min INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Progreso del usuario: lecciones completadas
CREATE TABLE IF NOT EXISTS progreso_lecciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leccion_id UUID NOT NULL REFERENCES lecciones(id) ON DELETE CASCADE,
    completado_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, leccion_id)
);

-- Tipos de pregunta en exámenes (multiple_choice | true_false | open_text)
ALTER TABLE preguntas ADD COLUMN IF NOT EXISTS tipo VARCHAR(30) NOT NULL DEFAULT 'multiple_choice';

-- Respuestas abiertas en exámenes
ALTER TABLE respuestas ADD COLUMN IF NOT EXISTS respuesta_texto TEXT;

-- Enlace examen → curso
ALTER TABLE examenes ADD COLUMN IF NOT EXISTS capacitacion_id UUID REFERENCES capacitaciones(id) ON DELETE SET NULL;

-- Foros por lección
CREATE TABLE IF NOT EXISTS foro_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    leccion_id UUID NOT NULL REFERENCES lecciones(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    licencia_id UUID REFERENCES curso_licencias(id) ON DELETE CASCADE,
    titulo VARCHAR(200) NOT NULL,
    contenido TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS foro_comentarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES foro_posts(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES foro_comentarios(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contenido TEXT NOT NULL,
    media_url TEXT DEFAULT '',
    media_type VARCHAR(20) DEFAULT '',
    is_private BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS foro_post_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES foro_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(post_id, user_id, emoji)
);

CREATE TABLE IF NOT EXISTS foro_comentario_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comentario_id UUID NOT NULL REFERENCES foro_comentarios(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(comentario_id, user_id, emoji)
);

-- Preguntas intermedias entre lecciones de un curso
CREATE TABLE IF NOT EXISTS preguntas_intermedias (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
    despues_de_leccion_id UUID REFERENCES lecciones(id) ON DELETE CASCADE,
    texto TEXT NOT NULL,
    tipo VARCHAR(30) NOT NULL DEFAULT 'multiple_choice',
    orden INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS opciones_intermedias (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pregunta_id UUID NOT NULL REFERENCES preguntas_intermedias(id) ON DELETE CASCADE,
    texto TEXT NOT NULL,
    es_correcta BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS respuestas_intermedias (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pregunta_id UUID NOT NULL REFERENCES preguntas_intermedias(id) ON DELETE CASCADE,
    opcion_id UUID REFERENCES opciones_intermedias(id) ON DELETE CASCADE,
    respuesta_texto TEXT,
    es_correcta BOOLEAN,
    respondido_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, pregunta_id)
);

-- Curso: mensaje de bienvenida y portada
ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS welcome_message TEXT DEFAULT '';
ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS thumbnail_url TEXT DEFAULT '';
ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS color VARCHAR(7) DEFAULT '#f97316';
ALTER TABLE capacitaciones ALTER COLUMN color TYPE TEXT;
ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS codigo_acceso VARCHAR(12);

-- Perfil instructor: especialidad
ALTER TABLE users ADD COLUMN IF NOT EXISTS specialty VARCHAR(255) DEFAULT '';

-- Archivos adjuntos en posts del foro
ALTER TABLE foro_posts ADD COLUMN IF NOT EXISTS media_url TEXT DEFAULT '';
ALTER TABLE foro_posts ADD COLUMN IF NOT EXISTS media_type VARCHAR(20) DEFAULT '';

-- Permitir respuestas abiertas sin opcion_id
ALTER TABLE respuestas ALTER COLUMN opcion_id DROP NOT NULL;

-- Likes en posts del foro
CREATE TABLE IF NOT EXISTS foro_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES foro_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(post_id, user_id)
);

-- Tokens de recuperación de contraseña (códigos de 1 solo uso, 15 min de vida)
CREATE TABLE IF NOT EXISTS password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_password_resets_email ON password_resets(email);

CREATE INDEX IF NOT EXISTS idx_preguntas_examen_id ON preguntas(examen_id);
CREATE INDEX IF NOT EXISTS idx_opciones_pregunta_id ON opciones(pregunta_id);
CREATE INDEX IF NOT EXISTS idx_respuestas_user_id ON respuestas(user_id);
CREATE INDEX IF NOT EXISTS idx_respuestas_examen_id ON respuestas(examen_id);
CREATE INDEX IF NOT EXISTS idx_respuestas_pregunta_id ON respuestas(pregunta_id);

CREATE TABLE IF NOT EXISTS notificaciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tipo VARCHAR(50) NOT NULL,
    titulo VARCHAR(200) NOT NULL,
    mensaje TEXT NOT NULL,
    leida BOOLEAN NOT NULL DEFAULT false,
    enlace TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notificaciones_user_id ON notificaciones(user_id);
CREATE INDEX IF NOT EXISTS idx_lecciones_capacitacion_id ON lecciones(capacitacion_id);
CREATE INDEX IF NOT EXISTS idx_progreso_user_id ON progreso_lecciones(user_id);
CREATE INDEX IF NOT EXISTS idx_progreso_leccion_id ON progreso_lecciones(leccion_id);
CREATE INDEX IF NOT EXISTS idx_inscripciones_user_id ON inscripciones(user_id);
CREATE INDEX IF NOT EXISTS idx_inscripciones_capacitacion_id ON inscripciones(capacitacion_id);
CREATE INDEX IF NOT EXISTS idx_asignaciones_user_id ON asignaciones(user_id);
CREATE INDEX IF NOT EXISTS idx_asignaciones_examen_id ON asignaciones(examen_id);
CREATE INDEX IF NOT EXISTS idx_foro_posts_leccion_id ON foro_posts(leccion_id);
CREATE INDEX IF NOT EXISTS idx_foro_posts_user_id ON foro_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_foro_comentarios_post_id ON foro_comentarios(post_id);
CREATE INDEX IF NOT EXISTS idx_preguntas_int_capacitacion_id ON preguntas_intermedias(capacitacion_id);
CREATE INDEX IF NOT EXISTS idx_respuestas_int_user_id ON respuestas_intermedias(user_id);
CREATE INDEX IF NOT EXISTS idx_respuestas_int_pregunta_id ON respuestas_intermedias(pregunta_id);
CREATE INDEX IF NOT EXISTS idx_foro_post_reac_post_id ON foro_post_reactions(post_id);
CREATE INDEX IF NOT EXISTS idx_foro_coment_reac_com_id ON foro_comentario_reactions(comentario_id);
CREATE INDEX IF NOT EXISTS idx_foro_comentarios_parent_id ON foro_comentarios(parent_id);

-- Soft deletes: los recursos borrados se ocultan pero se preserva el historial de estudiantes
ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE examenes       ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE lecciones      ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
CREATE INDEX IF NOT EXISTS idx_capacitaciones_deleted_at ON capacitaciones(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_examenes_deleted_at       ON examenes(deleted_at)       WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lecciones_deleted_at      ON lecciones(deleted_at)      WHERE deleted_at IS NULL;

-- Revocación de JWT: token_version permite invalidar tokens activos al cambiar contraseña o al banear usuario
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 1;
`

func Migrate() {
	if _, err := DB.Exec(Schema); err != nil {
		slog.Error("Error ejecutando migración", "error", err)
		os.Exit(1)
	}
	slog.Info("Migración ejecutada correctamente")
}
