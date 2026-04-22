package handlers

import (
	"net/http"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
)

type createExamenRequest struct {
	Title       string            `json:"title" binding:"required"`
	Description string            `json:"description"`
	Preguntas   []preguntaRequest `json:"preguntas" binding:"required,min=1"`
}

type preguntaRequest struct {
	Texto    string          `json:"texto" binding:"required"`
	Valor    float64         `json:"valor" binding:"required,gt=0"`
	Orden    int             `json:"orden"`
	Opciones []opcionRequest `json:"opciones" binding:"required,min=2"`
}

type opcionRequest struct {
	Texto      string `json:"texto" binding:"required"`
	EsCorrecta bool   `json:"es_correcta"`
}

func CreateExamen(c *gin.Context) {
	var req createExamenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	var examenID string
	err = tx.QueryRow(
		`INSERT INTO examenes(title, description) VALUES($1,$2) RETURNING id`,
		req.Title, req.Description,
	).Scan(&examenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, p := range req.Preguntas {
		var preguntaID string
		err = tx.QueryRow(
			`INSERT INTO preguntas(examen_id, texto, valor, orden) VALUES($1,$2,$3,$4) RETURNING id`,
			examenID, p.Texto, p.Valor, p.Orden,
		).Scan(&preguntaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, o := range p.Opciones {
			_, err = tx.Exec(
				`INSERT INTO opciones(pregunta_id, texto, es_correcta) VALUES($1,$2,$3)`,
				preguntaID, o.Texto, o.EsCorrecta,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": examenID})
}

func ListExamenes(c *gin.Context) {
	rows, err := db.DB.Query(`SELECT id, title, description, created_at FROM examenes ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.Examen{}
	for rows.Next() {
		var e models.Examen
		rows.Scan(&e.ID, &e.Title, &e.Description, &e.CreatedAt)
		result = append(result, e)
	}
	c.JSON(http.StatusOK, result)
}

func GetExamen(c *gin.Context) {
	id := c.Param("id")
	role, _ := c.Get("role")

	var examen models.Examen
	err := db.DB.QueryRow(
		`SELECT id, title, description, created_at FROM examenes WHERE id=$1`, id,
	).Scan(&examen.ID, &examen.Title, &examen.Description, &examen.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}

	rows, _ := db.DB.Query(`SELECT id, texto, valor, orden FROM preguntas WHERE examen_id=$1 ORDER BY orden`, id)
	defer rows.Close()
	for rows.Next() {
		var p models.Pregunta
		rows.Scan(&p.ID, &p.Texto, &p.Valor, &p.Orden)

		opRows, _ := db.DB.Query(`SELECT id, texto, es_correcta FROM opciones WHERE pregunta_id=$1`, p.ID)
		for opRows.Next() {
			var o models.Opcion
			opRows.Scan(&o.ID, &o.Texto, &o.EsCorrecta)
			// Solo admin ve cuál es correcta
			if role != "admin" {
				o.EsCorrecta = false
			}
			p.Opciones = append(p.Opciones, o)
		}
		opRows.Close()
		examen.Preguntas = append(examen.Preguntas, p)
	}

	c.JSON(http.StatusOK, examen)
}

func DeleteExamen(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec(`DELETE FROM examenes WHERE id=$1`, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Usuario: solo exámenes asignados
func ListExamenesUsuario(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT e.id, e.title, e.description, e.created_at
		FROM examenes e
		INNER JOIN asignaciones a ON a.examen_id = e.id
		WHERE a.user_id = $1
		ORDER BY e.created_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.Examen{}
	for rows.Next() {
		var e models.Examen
		rows.Scan(&e.ID, &e.Title, &e.Description, &e.CreatedAt)
		result = append(result, e)
	}
	c.JSON(http.StatusOK, result)
}

func SubmitExamen(c *gin.Context) {
	examenID := c.Param("id")
	userID, _ := c.Get("user_id")

	var respuestas []models.Respuesta
	if err := c.ShouldBindJSON(&respuestas); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var puntaje, puntajeMax float64

	for _, r := range respuestas {
		// Verificar que la pregunta pertenece al examen
		var valor float64
		err := db.DB.QueryRow(
			`SELECT valor FROM preguntas WHERE id=$1 AND examen_id=$2`, r.PreguntaID, examenID,
		).Scan(&valor)
		if err != nil {
			continue
		}
		puntajeMax += valor

		// Verificar si la opción es correcta
		var esCorrecta bool
		db.DB.QueryRow(
			`SELECT es_correcta FROM opciones WHERE id=$1 AND pregunta_id=$2`, r.OpcionID, r.PreguntaID,
		).Scan(&esCorrecta)
		if esCorrecta {
			puntaje += valor
		}

		// Guardar respuesta (upsert)
		db.DB.Exec(`
			INSERT INTO respuestas(user_id, examen_id, pregunta_id, opcion_id)
			VALUES($1,$2,$3,$4)
			ON CONFLICT(user_id, examen_id, pregunta_id) DO UPDATE SET opcion_id=$4, respondido_at=NOW()
		`, userID, examenID, r.PreguntaID, r.OpcionID)
	}

	porcentaje := 0.0
	if puntajeMax > 0 {
		porcentaje = (puntaje / puntajeMax) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"puntaje":     puntaje,
		"puntaje_max": puntajeMax,
		"porcentaje":  porcentaje,
	})
}
