package db

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
`

func Migrate() {
	if _, err := DB.Exec(Schema); err != nil {
		panic("Error ejecutando migración: " + err.Error())
	}
}
