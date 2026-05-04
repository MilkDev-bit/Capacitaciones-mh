package handlers

import (
	"net/http"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
)

type preguntaIntermediaReq struct {
	DespuesDeLeccionID *string        `json:"despues_de_leccion_id"`
	Texto              string         `json:"texto" binding:"required"`
	Tipo               string         `json:"tipo" binding:"required"`
	Orden              int            `json:"orden"`
	Opciones           []opcionIntReq `json:"opciones"`
}

type opcionIntReq struct {
	Texto      string `json:"texto" binding:"required"`
	EsCorrecta bool   `json:"es_correcta"`
}

func InstructorListPreguntasIntermedias(c *gin.Context) {
	capID := c.Param("id")
	instructorID, _ := c.Get("user_id")

	var owner string
	if err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM capacitaciones WHERE id=$1`, capID).Scan(&owner); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
		return
	}
	if owner != instructorID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permisos"})
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, capacitacion_id, despues_de_leccion_id, texto, tipo, orden
		FROM preguntas_intermedias WHERE capacitacion_id=$1 ORDER BY orden`, capID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	result := []models.PreguntaIntermedia{}
	for rows.Next() {
		var p models.PreguntaIntermedia
		rows.Scan(&p.ID, &p.CapacitacionID, &p.DespuesDeLeccionID, &p.Texto, &p.Tipo, &p.Orden)

		opRows, _ := db.DB.Query(`SELECT id, pregunta_id, texto, es_correcta FROM opciones_intermedias WHERE pregunta_id=$1`, p.ID)
		for opRows.Next() {
			var o models.OpcionIntermedia
			opRows.Scan(&o.ID, &o.PreguntaID, &o.Texto, &o.EsCorrecta)
			p.Opciones = append(p.Opciones, o)
		}
		opRows.Close()
		result = append(result, p)
	}
	c.JSON(http.StatusOK, result)
}

func InstructorCreatePreguntaIntermedia(c *gin.Context) {
	capID := c.Param("id")
	instructorID, _ := c.Get("user_id")

	var owner string
	if err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM capacitaciones WHERE id=$1`, capID).Scan(&owner); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
		return
	}
	if owner != instructorID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permisos"})
		return
	}

	var req preguntaIntermediaReq
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

	var pID string
	err = tx.QueryRow(`
		INSERT INTO preguntas_intermedias(capacitacion_id, despues_de_leccion_id, texto, tipo, orden)
		VALUES($1,$2,$3,$4,$5) RETURNING id`,
		capID, req.DespuesDeLeccionID, req.Texto, req.Tipo, req.Orden,
	).Scan(&pID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, o := range req.Opciones {
		tx.Exec(`INSERT INTO opciones_intermedias(pregunta_id, texto, es_correcta) VALUES($1,$2,$3)`,
			pID, o.Texto, o.EsCorrecta)
	}

	tx.Commit()
	c.JSON(http.StatusCreated, gin.H{"id": pID})
}

func InstructorDeletePreguntaIntermedia(c *gin.Context) {
	capID := c.Param("id")
	pID := c.Param("pregunta_id")
	instructorID, _ := c.Get("user_id")

	var owner string
	if err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM capacitaciones WHERE id=$1`, capID).Scan(&owner); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
		return
	}
	if owner != instructorID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permisos"})
		return
	}

	db.DB.Exec(`DELETE FROM preguntas_intermedias WHERE id=$1 AND capacitacion_id=$2`, pID, capID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Usuario: responder preguntas intermedias ────────────────────────────────

func GetPreguntasIntermedias(c *gin.Context) {
	capID := c.Param("id")
	leccionIDParam := c.Query("despues_de_leccion_id")
	userID, _ := c.Get("user_id")

	var rows interface {
		Next() bool
		Scan(...interface{}) error
		Close() error
	}
	var err error

	if leccionIDParam == "" {
		rows, err = db.DB.Query(`
			SELECT p.id, p.capacitacion_id, p.despues_de_leccion_id, p.texto, p.tipo, p.orden
			FROM preguntas_intermedias p
			WHERE p.capacitacion_id=$1 AND p.despues_de_leccion_id IS NULL
			  AND NOT EXISTS (SELECT 1 FROM respuestas_intermedias r WHERE r.pregunta_id=p.id AND r.user_id=$2)
			ORDER BY p.orden`, capID, userID)
	} else {
		rows, err = db.DB.Query(`
			SELECT p.id, p.capacitacion_id, p.despues_de_leccion_id, p.texto, p.tipo, p.orden
			FROM preguntas_intermedias p
			WHERE p.capacitacion_id=$1 AND p.despues_de_leccion_id=$2
			  AND NOT EXISTS (SELECT 1 FROM respuestas_intermedias r WHERE r.pregunta_id=p.id AND r.user_id=$3)
			ORDER BY p.orden`, capID, leccionIDParam, userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	result := []models.PreguntaIntermedia{}
	for rows.Next() {
		var p models.PreguntaIntermedia
		rows.Scan(&p.ID, &p.CapacitacionID, &p.DespuesDeLeccionID, &p.Texto, &p.Tipo, &p.Orden)

		opRows, _ := db.DB.Query(`SELECT id, pregunta_id, texto FROM opciones_intermedias WHERE pregunta_id=$1`, p.ID)
		for opRows.Next() {
			var o models.OpcionIntermedia
			opRows.Scan(&o.ID, &o.PreguntaID, &o.Texto)
			p.Opciones = append(p.Opciones, o)
		}
		opRows.Close()
		result = append(result, p)
	}
	c.JSON(http.StatusOK, result)
}

func SubmitPreguntasIntermedias(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var respuestas []struct {
		PreguntaID     string  `json:"pregunta_id"`
		OpcionID       *string `json:"opcion_id"`
		RespuestaTexto string  `json:"respuesta_texto"`
	}
	if err := c.ShouldBindJSON(&respuestas); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results := []gin.H{}
	for _, r := range respuestas {
		var esCorrecta *bool

		if r.OpcionID != nil && *r.OpcionID != "" {
			var ec bool
			db.DB.QueryRow(`SELECT es_correcta FROM opciones_intermedias WHERE id=$1 AND pregunta_id=$2`,
				r.OpcionID, r.PreguntaID).Scan(&ec)
			esCorrecta = &ec
		}

		db.DB.Exec(`
			INSERT INTO respuestas_intermedias(user_id, pregunta_id, opcion_id, respuesta_texto, es_correcta)
			VALUES($1,$2,$3,$4,$5)
			ON CONFLICT(user_id, pregunta_id) DO UPDATE
			  SET opcion_id=$3, respuesta_texto=$4, es_correcta=$5`,
			userID, r.PreguntaID, r.OpcionID, r.RespuestaTexto, esCorrecta)

		results = append(results, gin.H{"pregunta_id": r.PreguntaID, "es_correcta": esCorrecta})
	}
	c.JSON(http.StatusOK, results)
}
