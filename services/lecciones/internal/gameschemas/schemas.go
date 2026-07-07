// Package gameschemas documenta los contratos JSON para cada tipo de minijuego.
//
// El campo game_config_json en la tabla lecciones almacena un objeto JSON
// cuya estructura depende del lesson_type. Este archivo NO se compila en
// producción; es documentación ejecutable para desarrolladores.
//
// Cada schema es 100% editable por el instructor desde el panel de administración.
// El frontend valida el JSON antes de enviarlo; el backend lo almacena sin parsear
// (responsabilidad del frontend renderizarlo correctamente).
package gameschemas

// ─────────────────────────────────────────────────────────────────────────────
// LESSON_TYPE_GAME_MEMORY — Memorama de pares
// ─────────────────────────────────────────────────────────────────────────────
// El instructor define N pares. El frontend baraja las tarjetas y el alumno
// debe encontrar todos los pares. La puntuación se basa en el número de
// intentos y el tiempo.
//
// Ejemplo:
// {
//   "instruction": "Encuentra los pares de conceptos relacionados",
//   "pairs": [
//     { "front": "HTTP",       "back": "Protocolo de transferencia de hipertexto" },
//     { "front": "DNS",        "back": "Sistema de nombres de dominio",
//                              "image_url": "https://..." },
//     { "front": "TCP/IP",     "back": "Protocolo de control de transmisión" }
//   ],
//   "max_time_secs": 120,
//   "show_labels": true
// }
type MemoryConfig struct {
	Instruction string       `json:"instruction"` // Texto de instrucción para el alumno
	Pairs       []MemoryPair `json:"pairs"`       // Mínimo 2, máximo 24 pares
	MaxTimeSecs int          `json:"max_time_secs,omitempty"` // 0 = sin límite
	ShowLabels  bool         `json:"show_labels"` // Muestra texto debajo de las imágenes
}

type MemoryPair struct {
	Front    string `json:"front"`             // Texto de la cara A
	Back     string `json:"back"`              // Texto de la cara B
	ImageURL string `json:"image_url,omitempty"` // Imagen opcional (cara A)
}

// ─────────────────────────────────────────────────────────────────────────────
// LESSON_TYPE_GAME_DRAGDROP — Arrastrar y soltar
// ─────────────────────────────────────────────────────────────────────────────
// El alumno arrastra cada ítem a su categoría correcta.
//
// Ejemplo:
// {
//   "instruction": "Clasifica cada elemento en su categoría",
//   "categories": ["Frontend", "Backend", "Base de datos"],
//   "items": [
//     { "text": "React",      "correct_category": "Frontend" },
//     { "text": "PostgreSQL", "correct_category": "Base de datos" },
//     { "text": "Go",         "correct_category": "Backend" },
//     { "text": "Vue.js",     "correct_category": "Frontend",
//       "image_url": "https://..." }
//   ]
// }
type DragDropConfig struct {
	Instruction string         `json:"instruction"`
	Categories  []string       `json:"categories"` // Zonas de destino
	Items       []DragDropItem `json:"items"`
}

type DragDropItem struct {
	Text            string `json:"text"`
	CorrectCategory string `json:"correct_category"` // Debe coincidir exacto con una categoría
	ImageURL        string `json:"image_url,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// LESSON_TYPE_GAME_WORDSEARCH — Sopa de letras
// ─────────────────────────────────────────────────────────────────────────────
// El instructor define las palabras; el frontend genera la cuadrícula
// automáticamente. Las palabras se pueden ocultar horizontal, vertical y
// diagonalmente según la dificultad.
//
// Ejemplo:
// {
//   "instruction": "Encuentra los conceptos de redes en la sopa de letras",
//   "words": ["HTTP", "DNS", "TCP", "IP", "ROUTER", "SWITCH"],
//   "grid_size": 12,
//   "difficulty": "medium",
//   "show_word_list": true
// }
type WordSearchConfig struct {
	Instruction   string   `json:"instruction"`
	Words         []string `json:"words"`               // Máximo 20 palabras
	GridSize      int      `json:"grid_size"`           // Tamaño de la cuadrícula N×N (8-20)
	Difficulty    string   `json:"difficulty"`          // "easy" | "medium" | "hard"
	ShowWordList  bool     `json:"show_word_list"`      // Muestra la lista de palabras a buscar
}

// ─────────────────────────────────────────────────────────────────────────────
// LESSON_TYPE_GAME_FILLBLANK — Completar el espacio en blanco
// ─────────────────────────────────────────────────────────────────────────────
// El alumno escribe o selecciona la palabra que falta. Soporta dos modos:
// - "type": el alumno escribe la respuesta (texto libre, case-insensitive)
// - "select": el alumno elige entre opciones (opciones se mezclan automáticamente)
//
// El marcador ___ (triple guion bajo) en el texto indica dónde va la respuesta.
//
// Ejemplo:
// {
//   "instruction": "Completa las definiciones",
//   "mode": "select",
//   "sentences": [
//     {
//       "text": "El protocolo ___ es el estándar para transferir páginas web.",
//       "answer": "HTTP",
//       "options": ["HTTP", "FTP", "SSH", "DNS"]
//     },
//     {
//       "text": "Una dirección IP versión 4 tiene ___ octetos.",
//       "answer": "4"
//     }
//   ]
// }
type FillBlankConfig struct {
	Instruction string          `json:"instruction"`
	Mode        string          `json:"mode"` // "type" | "select"
	Sentences   []FillBlankItem `json:"sentences"`
}

type FillBlankItem struct {
	Text    string   `json:"text"`              // Texto con ___ como marcador de posición
	Answer  string   `json:"answer"`            // Respuesta correcta (case-insensitive en "type")
	Options []string `json:"options,omitempty"` // Solo para mode="select"
}

// ─────────────────────────────────────────────────────────────────────────────
// LESSON_TYPE_GAME_ORDER — Ordenar secuencia / pasos
// ─────────────────────────────────────────────────────────────────────────────
// El alumno ordena una lista de ítems en la secuencia correcta. Útil para
// procesos, algoritmos, líneas de tiempo, etc.
//
// Ejemplo:
// {
//   "instruction": "Ordena los pasos del proceso de desarrollo",
//   "items": [
//     { "text": "Definir requerimientos",  "correct_order": 1 },
//     { "text": "Diseñar la arquitectura", "correct_order": 2 },
//     { "text": "Implementar el código",   "correct_order": 3 },
//     { "text": "Realizar pruebas",        "correct_order": 4 },
//     { "text": "Desplegar a producción",  "correct_order": 5 }
//   ],
//   "show_numbers": false
// }
type OrderConfig struct {
	Instruction string      `json:"instruction"`
	Items       []OrderItem `json:"items"`
	ShowNumbers bool        `json:"show_numbers"` // Muestra números de posición como pistas
}

type OrderItem struct {
	Text         string `json:"text"`
	CorrectOrder int    `json:"correct_order"` // 1-indexed
	ImageURL     string `json:"image_url,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Cálculo de puntos por minijuego
// ─────────────────────────────────────────────────────────────────────────────
// El frontend calcula los puntos obtenidos (0..points_reward) y los envía
// via POST /api/lecciones/:id/game-score. El backend NO re-valida la lógica
// del juego; asume que el frontend es la fuente de verdad (los valores se
// sanitizan a max=points_reward).
//
// Fórmula sugerida por tipo:
//
//   MEMORY:    points = points_reward × (1 - intentos_extras/total_pares×0.5)
//              Bonus de velocidad: +10% si tiempo < max_time_secs/2
//
//   DRAGDROP:  points = (correctos/total) × points_reward
//              Sin penalización por intentos adicionales
//
//   WORDSEARCH: points = (palabras_encontradas/total_palabras) × points_reward
//               Bonus: +20% si se encuentran todas antes del tiempo límite
//
//   FILLBLANK: points = (correctos/total) × points_reward
//              Modo "type": tolerancia de 1 error tipográfico (Levenshtein ≤ 1)
//
//   ORDER:     points = (posiciones_correctas/total) × points_reward
//              Puntuación parcial: cada ítem en posición correcta suma
