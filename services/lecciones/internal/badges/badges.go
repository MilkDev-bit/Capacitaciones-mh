// Package badges define el catálogo de insignias y las reglas para desbloquearlas.
// Es la fuente de verdad para el frontend (metadata de insignias) y para el
// servicio que evalúa si se debe otorgar una insignia tras un evento.
package badges

import leccionespb "Prueba-Go/gen/lecciones"

// ─────────────────────────────────────────────────────────────────────────────
// Catálogo de insignias
// ─────────────────────────────────────────────────────────────────────────────

type BadgeDef struct {
	Slug        string
	Name        string
	Description string
	IconURL     string // Ruta relativa o URL absoluta
	Color       string // Color hex de acento
}

// Catalog es la lista maestra de todas las insignias disponibles.
// Agregar una nueva insignia aquí la hace disponible en todo el sistema.
var Catalog = []BadgeDef{
	// ── Primeras veces ────────────────────────────────────────────────────────
	{
		Slug:        "primera_leccion",
		Name:        "Primer Paso",
		Description: "Completaste tu primera lección.",
		IconURL:     "/badges/primera_leccion.svg",
		Color:       "#6366f1",
	},
	{
		Slug:        "primer_juego",
		Name:        "Jugador Novato",
		Description: "Completaste tu primer minijuego.",
		IconURL:     "/badges/primer_juego.svg",
		Color:       "#8b5cf6",
	},

	// ── Memorama ──────────────────────────────────────────────────────────────
	{
		Slug:        "memorama_perfecto",
		Name:        "Memoria de Elefante",
		Description: "Completaste un memorama sin errores.",
		IconURL:     "/badges/memorama.svg",
		Color:       "#f59e0b",
	},
	{
		Slug:        "memorama_veloz",
		Name:        "Flash Mental",
		Description: "Completaste un memorama en menos de 30 segundos.",
		IconURL:     "/badges/memorama_veloz.svg",
		Color:       "#fbbf24",
	},

	// ── Drag & Drop ───────────────────────────────────────────────────────────
	{
		Slug:        "dragdrop_perfecto",
		Name:        "Clasificador Experto",
		Description: "Clasificaste todos los elementos correctamente al primer intento.",
		IconURL:     "/badges/dragdrop.svg",
		Color:       "#10b981",
	},

	// ── Sopa de letras ────────────────────────────────────────────────────────
	{
		Slug:        "wordsearch_completo",
		Name:        "Cazador de Palabras",
		Description: "Encontraste todas las palabras en la sopa de letras.",
		IconURL:     "/badges/wordsearch.svg",
		Color:       "#3b82f6",
	},

	// ── Completar espacios ────────────────────────────────────────────────────
	{
		Slug:        "fillblank_perfecto",
		Name:        "Lingüista",
		Description: "Completaste todos los espacios correctamente.",
		IconURL:     "/badges/fillblank.svg",
		Color:       "#ec4899",
	},

	// ── Ordenar ───────────────────────────────────────────────────────────────
	{
		Slug:        "order_perfecto",
		Name:        "Maestro del Orden",
		Description: "Ordenaste la secuencia perfectamente.",
		IconURL:     "/badges/order.svg",
		Color:       "#f97316",
	},

	// ── Puntos acumulados ─────────────────────────────────────────────────────
	{
		Slug:        "puntos_100",
		Name:        "Centurión",
		Description: "Acumulaste 100 puntos en un curso.",
		IconURL:     "/badges/puntos_100.svg",
		Color:       "#64748b",
	},
	{
		Slug:        "puntos_500",
		Name:        "Guerrero del Conocimiento",
		Description: "Acumulaste 500 puntos en un curso.",
		IconURL:     "/badges/puntos_500.svg",
		Color:       "#a78bfa",
	},
	{
		Slug:        "puntos_1000",
		Name:        "Leyenda",
		Description: "Acumulaste 1,000 puntos en un curso.",
		IconURL:     "/badges/puntos_1000.svg",
		Color:       "#fbbf24",
	},

	// ── Racha ─────────────────────────────────────────────────────────────────
	{
		Slug:        "streak_5",
		Name:        "En Racha",
		Description: "Completaste 5 actividades seguidas.",
		IconURL:     "/badges/streak.svg",
		Color:       "#ef4444",
	},
	{
		Slug:        "top_leaderboard",
		Name:        "Campeón",
		Description: "Llegaste al #1 del leaderboard de un curso.",
		IconURL:     "/badges/champion.svg",
		Color:       "#f59e0b",
	},
}

// CatalogMap permite lookup O(1) por slug.
var CatalogMap = func() map[string]BadgeDef {
	m := make(map[string]BadgeDef, len(Catalog))
	for _, b := range Catalog {
		m[b.Slug] = b
	}
	return m
}()

// ─────────────────────────────────────────────────────────────────────────────
// Evento de juego — contiene toda la info para evaluar reglas
// ─────────────────────────────────────────────────────────────────────────────

type GameEvent struct {
	LessonType       string // valor del enum LessonType
	Points           int32  // puntos obtenidos en este intento
	PointsReward     int32  // máximo posible en esta lección
	TimeSecs         int32  // tiempo del intento
	TotalCoursePoints int32 // puntos acumulados en el curso (ya incluye este intento)
	IsFirstGame      bool   // primer minijuego completado por el usuario (ever)
}

// EvaluateRules devuelve los slugs de insignias que se deben otorgar por este evento.
// Solo devuelve insignias que aún NO tiene el usuario (el caller debe filtrar).
func EvaluateRules(ev GameEvent) []string {
	var earned []string

	add := func(slug string) { earned = append(earned, slug) }

	isPerfect := ev.PointsReward > 0 && ev.Points >= ev.PointsReward

	// ── Primer juego ever ─────────────────────────────────────────────────────
	if ev.IsFirstGame {
		add("primer_juego")
	}

	// ── Reglas por tipo de juego ──────────────────────────────────────────────
	switch ev.LessonType {
	case leccionespb.LessonType_LESSON_TYPE_GAME_MEMORY.String():
		if isPerfect {
			add("memorama_perfecto")
		}
		if ev.TimeSecs > 0 && ev.TimeSecs < 30 {
			add("memorama_veloz")
		}
	case leccionespb.LessonType_LESSON_TYPE_GAME_DRAGDROP.String():
		if isPerfect {
			add("dragdrop_perfecto")
		}
	case leccionespb.LessonType_LESSON_TYPE_GAME_WORDSEARCH.String():
		if isPerfect {
			add("wordsearch_completo")
		}
	case leccionespb.LessonType_LESSON_TYPE_GAME_FILLBLANK.String():
		if isPerfect {
			add("fillblank_perfecto")
		}
	case leccionespb.LessonType_LESSON_TYPE_GAME_ORDER.String():
		if isPerfect {
			add("order_perfecto")
		}
	}

	// ── Puntos acumulados en el curso ─────────────────────────────────────────
	switch {
	case ev.TotalCoursePoints >= 1000:
		add("puntos_1000")
	case ev.TotalCoursePoints >= 500:
		add("puntos_500")
	case ev.TotalCoursePoints >= 100:
		add("puntos_100")
	}

	return earned
}

// ToProto convierte una definición de insignia al proto de respuesta.
func (b BadgeDef) ToProto(unlockedAt string) *leccionespb.Badge {
	// Note: leccionespb.Badge es del proto de usuarios, lo dejamos como string map
	// para no importar el paquete de usuarios aquí (evitar ciclo de dependencias).
	return &leccionespb.Badge{
		Slug:        b.Slug,
		Name:        b.Name,
		Description: b.Description,
		IconUrl:     b.IconURL,
		Color:       b.Color,
		UnlockedAt:  unlockedAt,
	}
}
