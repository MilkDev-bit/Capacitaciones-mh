package db

import "log"

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

CREATE TABLE IF NOT EXISTS inscripciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
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
    titulo VARCHAR(200) NOT NULL,
    contenido TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS foro_comentarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES foro_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contenido TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
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
`

func Migrate() {
	if _, err := DB.Exec(Schema); err != nil {
		log.Fatalf("Error ejecutando migración: %v", err)
	}
	log.Println("Migración ejecutada correctamente")
}
