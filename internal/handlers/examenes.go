package handlers

import (
	"log"
	"net/http"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createExamenRequest struct {
	Title          string            `json:"title" binding:"required"`
	Description    string            `json:"description"`
	CapacitacionID *string           `json:"capacitacion_id"`
	Preguntas      []preguntaRequest `json:"preguntas" binding:"required,min=1"`
}

type preguntaRequest struct {
	Texto    string          `json:"texto" binding:"required"`
	Tipo     string          `json:"tipo"`
	Valor    float64         `json:"valor" binding:"required,gt=0"`
	Orden    int             `json:"orden"`
	Opciones []opcionRequest `json:"opciones"`
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
		log.Printf("[ERROR] CreateExamen tx.Begin: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer tx.Rollback()

	var examenID string
	err = tx.QueryRow(
		`INSERT INTO examenes(title, description, capacitacion_id) VALUES($1,$2,$3) RETURNING id`,
		req.Title, req.Description, req.CapacitacionID,
	).Scan(&examenID)
	if err != nil {
		log.Printf("[ERROR] CreateExamen insert examen: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	for _, p := range req.Preguntas {
		tipo := p.Tipo
		if tipo == "" {
			tipo = "multiple_choice"
		}
		var preguntaID string
		err = tx.QueryRow(
			`INSERT INTO preguntas(examen_id, texto, tipo, valor, orden) VALUES($1,$2,$3,$4,$5) RETURNING id`,
			examenID, p.Texto, tipo, p.Valor, p.Orden,
		).Scan(&preguntaID)
		if err != nil {
			log.Printf("[ERROR] CreateExamen insert pregunta: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
			return
		}
		if tipo != "open_text" {
			for _, o := range p.Opciones {
				_, err = tx.Exec(
					`INSERT INTO opciones(pregunta_id, texto, es_correcta) VALUES($1,$2,$3)`,
					preguntaID, o.Texto, o.EsCorrecta,
				)
				if err != nil {
					log.Printf("[ERROR] CreateExamen insert opcion: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
					return
				}
			}
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("[ERROR] CreateExamen commit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": examenID})
}

func ListExamenes(c *gin.Context) {
	limit, offset, page := parsePagination(c)
	var total int
	db.DB.QueryRow(`SELECT COUNT(*) FROM examenes WHERE deleted_at IS NULL`).Scan(&total)
	rows, err := db.DB.Query(`SELECT id, title, description, created_at FROM examenes WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		log.Printf("[ERROR] ListExamenes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()
	result := []models.Examen{}
	for rows.Next() {
		var e models.Examen
		rows.Scan(&e.ID, &e.Title, &e.Description, &e.CreatedAt)
		result = append(result, e)
	}
	c.JSON(http.StatusOK, gin.H{"data": result, "total": total, "page": page, "limit": limit})
}

func GetExamen(c *gin.Context) {
	id := c.Param("id")
	role, _ := c.Get("role")

	var examen models.Examen
	err := db.DB.QueryRow(
		`SELECT e.id, e.title, e.description, e.created_at,
		        e.capacitacion_id, COALESCE(cap.title,'')
		 FROM examenes e
		 LEFT JOIN capacitaciones cap ON cap.id = e.capacitacion_id
		 WHERE e.id=$1 AND e.deleted_at IS NULL`, id,
	).Scan(&examen.ID, &examen.Title, &examen.Description, &examen.CreatedAt,
		&examen.CapacitacionID, &examen.CapacitacionNombre)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}

	rows, _ := db.DB.Query(`SELECT id, texto, COALESCE(tipo,'multiple_choice'), valor, orden FROM preguntas WHERE examen_id=$1 ORDER BY orden`, id)
	var preguntaIDs []string
	for rows.Next() {
		var p models.Pregunta
		rows.Scan(&p.ID, &p.Texto, &p.Tipo, &p.Valor, &p.Orden)
		examen.Preguntas = append(examen.Preguntas, p)
		if p.Tipo != "open_text" {
			preguntaIDs = append(preguntaIDs, p.ID)
		}
	}
	rows.Close()

	if len(preguntaIDs) > 0 {
		opcionesMap := make(map[string][]models.Opcion)
		opRows, _ := db.DB.Query(`SELECT id, texto, es_correcta, pregunta_id FROM opciones WHERE pregunta_id = ANY($1)`, pq.Array(preguntaIDs))
		for opRows.Next() {
			var o models.Opcion
			opRows.Scan(&o.ID, &o.Texto, &o.EsCorrecta, &o.PreguntaID)
			if role != "admin" && role != "instructor" {
				o.EsCorrecta = false
			}
			opcionesMap[o.PreguntaID] = append(opcionesMap[o.PreguntaID], o)
		}
		opRows.Close()
		for i := range examen.Preguntas {
			examen.Preguntas[i].Opciones = opcionesMap[examen.Preguntas[i].ID]
		}
	}

	if role == "user" {
		uid, _ := c.Get("user_id")
		userID := uid.(string)

		// Verifica que el usuario tenga acceso: asignación directa O inscripción al curso vinculado.
		var access int
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM (
				SELECT 1 FROM asignaciones WHERE user_id=$1 AND examen_id=$2
				UNION ALL
				SELECT 1 FROM inscripciones i
				INNER JOIN examenes e ON e.capacitacion_id = i.capacitacion_id
				WHERE i.user_id=$1 AND e.id=$2
			) t
		`, userID, id).Scan(&access)
		if access == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "no tienes acceso a este examen"})
			return
		}

		yaRespondido, puntaje, puntajeMax, porcentaje := getExamenResultadoUsuario(id, userID)
		c.JSON(http.StatusOK, gin.H{
			"id": examen.ID, "title": examen.Title, "description": examen.Description,
			"created_at": examen.CreatedAt, "capacitacion_id": examen.CapacitacionID,
			"capacitacion_nombre": examen.CapacitacionNombre, "preguntas": examen.Preguntas,
			"ya_respondido": yaRespondido, "porcentaje": porcentaje,
			"puntaje": puntaje, "puntaje_max": puntajeMax,
		})
		return
	}
	c.JSON(http.StatusOK, examen)
}

func getExamenResultadoUsuario(examenID, userID string) (yaRespondido bool, puntaje, puntajeMax, porcentaje float64) {
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM respuestas WHERE user_id=$1 AND examen_id=$2`, userID, examenID).Scan(&count)
	if count == 0 {
		return
	}
	yaRespondido = true
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN o.es_correcta THEN p.valor ELSE 0 END),0),
		       COALESCE(SUM(p.valor),0)
		FROM respuestas r
		INNER JOIN preguntas p ON p.id = r.pregunta_id
		LEFT  JOIN opciones  o ON o.id = r.opcion_id
		WHERE r.user_id=$1 AND r.examen_id=$2
	`, userID, examenID).Scan(&puntaje, &puntajeMax)
	if puntajeMax > 0 {
		porcentaje = puntaje / puntajeMax * 100
	}
	return
}

func DeleteExamen(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec(`UPDATE examenes SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListExamenesUsuario(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT e.id, e.title, e.description, e.created_at,
		       COALESCE(e.capacitacion_id::text,''),
		       -- ya_respondido
		       EXISTS(
		           SELECT 1 FROM respuestas r WHERE r.examen_id=e.id AND r.user_id=$1
		       ) AS ya_respondido,
		       -- porcentaje obtenido
		       COALESCE((
		           SELECT CASE WHEN SUM(p2.valor)>0
		                       THEN SUM(CASE WHEN o2.es_correcta THEN p2.valor ELSE 0 END)/SUM(p2.valor)*100
		                       ELSE 0 END
		           FROM respuestas r2
		           INNER JOIN preguntas p2 ON p2.id=r2.pregunta_id
		           LEFT  JOIN opciones  o2 ON o2.id=r2.opcion_id
		           WHERE r2.examen_id=e.id AND r2.user_id=$1
		       ),0) AS porcentaje,
		       -- bloqueado: enlazado a un curso y el usuario no lo completó todo
		       CASE WHEN e.capacitacion_id IS NULL THEN false
		            ELSE (
		                EXISTS(
		                    SELECT 1 FROM lecciones l
		                    WHERE l.capacitacion_id=e.capacitacion_id
		                    AND NOT EXISTS(
		                        SELECT 1 FROM progreso_lecciones pl
		                        WHERE pl.leccion_id=l.id AND pl.user_id=$1
		                    )
		                )
		                OR EXISTS(
		                    SELECT 1 FROM preguntas_intermedias pi
		                    WHERE pi.capacitacion_id=e.capacitacion_id
		                    AND NOT EXISTS(
		                        SELECT 1 FROM respuestas_intermedias ri
		                        WHERE ri.pregunta_id=pi.id AND ri.user_id=$1
		                    )
		                )
		            )
		       END AS bloqueado
		FROM examenes e
		INNER JOIN asignaciones a ON a.examen_id=e.id
		WHERE a.user_id=$1
		ORDER BY e.created_at DESC
	`, userID)
	if err != nil {
		log.Printf("[ERROR] ListExamenesUsuario: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()

	type ExamenUsuario struct {
		ID             string  `json:"id"`
		Title          string  `json:"title"`
		Description    string  `json:"description"`
		CreatedAt      string  `json:"created_at"`
		CapacitacionID string  `json:"capacitacion_id"`
		YaRespondido   bool    `json:"ya_respondido"`
		Porcentaje     float64 `json:"porcentaje"`
		Bloqueado      bool    `json:"bloqueado"`
	}
	result := []ExamenUsuario{}
	for rows.Next() {
		var e ExamenUsuario
		var createdAt interface{}
		rows.Scan(&e.ID, &e.Title, &e.Description, &createdAt, &e.CapacitacionID,
			&e.YaRespondido, &e.Porcentaje, &e.Bloqueado)
		if t, ok := createdAt.(string); ok {
			e.CreatedAt = t
		}
		result = append(result, e)
	}
	c.JSON(http.StatusOK, result)
}

func InstructorGetResultados(c *gin.Context) {
	examenID := c.Param("id")
	instructorID, _ := c.Get("user_id")

	var owner string
	if err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM examenes WHERE id=$1`, examenID).Scan(&owner); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "examen no encontrado"})
		return
	}
	if owner != instructorID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permisos"})
		return
	}

	rows, err := db.DB.Query(`
		SELECT u.id, u.name, u.email,
		       COALESCE(SUM(CASE WHEN o.es_correcta THEN p.valor ELSE 0 END), 0),
		       COALESCE(SUM(p.valor), 0),
		       COUNT(r.id),
		       MAX(r.respondido_at)
		FROM respuestas r
		INNER JOIN users u ON u.id = r.user_id
		INNER JOIN preguntas p ON p.id = r.pregunta_id
		LEFT  JOIN opciones o ON o.id = r.opcion_id
		WHERE r.examen_id = $1
		GROUP BY u.id, u.name, u.email
		ORDER BY MAX(r.respondido_at) DESC
	`, examenID)
	if err != nil {
		log.Printf("[ERROR] InstructorGetResultados: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()

	type Resultado struct {
		UserID           string  `json:"user_id"`
		Nombre           string  `json:"nombre"`
		Email            string  `json:"email"`
		Puntaje          float64 `json:"puntaje"`
		PuntajeMax       float64 `json:"puntaje_max"`
		Porcentaje       float64 `json:"porcentaje"`
		TotalRespondidas int     `json:"total_respondidas"`
		RespondidoAt     string  `json:"respondido_at"`
	}
	var result []Resultado
	for rows.Next() {
		var r Resultado
		rows.Scan(&r.UserID, &r.Nombre, &r.Email, &r.Puntaje, &r.PuntajeMax, &r.TotalRespondidas, &r.RespondidoAt)
		if r.PuntajeMax > 0 {
			r.Porcentaje = (r.Puntaje / r.PuntajeMax) * 100
		}
		result = append(result, r)
	}
	if result == nil {
		result = []Resultado{}
	}
	c.JSON(http.StatusOK, result)
}

func InstructorGetRespuestasUsuario(c *gin.Context) {
	examenID := c.Param("id")
	userID := c.Param("user_id")
	instructorID, _ := c.Get("user_id")

	var owner string
	if err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM examenes WHERE id=$1`, examenID).Scan(&owner); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "examen no encontrado"})
		return
	}
	if owner != instructorID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permisos"})
		return
	}

	rows, err := db.DB.Query(`
		SELECT p.id, p.texto, COALESCE(p.tipo,'multiple_choice'), p.valor,
		       COALESCE(o_sel.texto, COALESCE(r.respuesta_texto,''), '') AS respuesta_dada,
		       COALESCE(o_sel.es_correcta, false)                         AS es_correcta,
		       COALESCE(o_cor.texto, '')                                   AS respuesta_correcta
		FROM preguntas p
		LEFT JOIN respuestas r
		       ON r.pregunta_id = p.id AND r.user_id = $2 AND r.examen_id = $1
		LEFT JOIN opciones o_sel ON o_sel.id = r.opcion_id
		LEFT JOIN opciones o_cor
		       ON o_cor.pregunta_id = p.id AND o_cor.es_correcta = true
		WHERE p.examen_id = $1
		ORDER BY p.orden
	`, examenID, userID)
	if err != nil {
		log.Printf("[ERROR] InstructorGetRespuestasUsuario: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()

	type RespuestaDetalle struct {
		PreguntaID        string  `json:"pregunta_id"`
		Texto             string  `json:"texto"`
		Tipo              string  `json:"tipo"`
		Valor             float64 `json:"valor"`
		RespuestaDada     string  `json:"respuesta_dada"`
		EsCorrecta        bool    `json:"es_correcta"`
		RespuestaCorrecta string  `json:"respuesta_correcta"`
	}
	var result []RespuestaDetalle
	for rows.Next() {
		var r RespuestaDetalle
		rows.Scan(&r.PreguntaID, &r.Texto, &r.Tipo, &r.Valor, &r.RespuestaDada, &r.EsCorrecta, &r.RespuestaCorrecta)
		result = append(result, r)
	}
	if result == nil {
		result = []RespuestaDetalle{}
	}
	c.JSON(http.StatusOK, result)
}

func SubmitExamen(c *gin.Context) {
	examenID := c.Param("id")
	userID, _ := c.Get("user_id")

	var existCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM respuestas WHERE user_id=$1 AND examen_id=$2`, userID, examenID).Scan(&existCount)
	if existCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Ya has respondido este examen. Solicita a tu instructor que lo reasigne."})
		return
	}

	var respuestas []models.Respuesta
	if err := c.ShouldBindJSON(&respuestas); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	defer tx.Rollback()

	var puntaje, puntajeMax float64

	for _, r := range respuestas {
		var valor float64
		var tipo string
		err := tx.QueryRow(
			`SELECT valor, COALESCE(tipo,'multiple_choice') FROM preguntas WHERE id=$1 AND examen_id=$2`, r.PreguntaID, examenID,
		).Scan(&valor, &tipo)
		if err != nil {
			continue
		}
		puntajeMax += valor

		if tipo == "open_text" {
			tx.Exec(`
				INSERT INTO respuestas(user_id, examen_id, pregunta_id, respuesta_texto)
				VALUES($1,$2,$3,$4)
				ON CONFLICT(user_id, examen_id, pregunta_id) DO UPDATE SET respuesta_texto=$4, respondido_at=NOW()
			`, userID, examenID, r.PreguntaID, r.RespuestaTexto)
		} else {
			var esCorrecta bool
			if r.OpcionID != "" {
				tx.QueryRow(
					`SELECT es_correcta FROM opciones WHERE id=$1 AND pregunta_id=$2`, r.OpcionID, r.PreguntaID,
				).Scan(&esCorrecta)
				if esCorrecta {
					puntaje += valor
				}
				tx.Exec(`
					INSERT INTO respuestas(user_id, examen_id, pregunta_id, opcion_id)
					VALUES($1,$2,$3,$4)
					ON CONFLICT(user_id, examen_id, pregunta_id) DO UPDATE SET opcion_id=$4, respondido_at=NOW()
				`, userID, examenID, r.PreguntaID, r.OpcionID)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
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
