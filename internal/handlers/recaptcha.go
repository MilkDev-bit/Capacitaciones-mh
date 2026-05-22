package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

type recaptchaResponse struct {
	Success    bool     `json:"success"`
	Score      float64  `json:"score"`
	Action     string   `json:"action"`
	ErrorCodes []string `json:"error-codes"`
}

// verifyRecaptcha verifica un token de reCAPTCHA v3 contra la API de Google.
// Retorna true si el score es aceptable (>= 0.5).
// Si RECAPTCHA_SECRET_KEY no está configurado, permite todo (útil en desarrollo local).
func verifyRecaptcha(token string) bool {
	secret := os.Getenv("RECAPTCHA_SECRET_KEY")
	if secret == "" {
		return true // sin clave configurada → skip (dev)
	}
	if token == "" {
		// Token vacío: VITE_RECAPTCHA_SITE_KEY probablemente no estaba en el build.
		// Registrar advertencia y permitir para no bloquear usuarios legítimos.
		log.Printf("[RECAPTCHA] advertencia: token vacío recibido (¿VITE_RECAPTCHA_SITE_KEY no configurada en el build?)")
		return true
	}
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {secret},
		"response": {token},
	})
	if err != nil {
		log.Printf("[RECAPTCHA] error al contactar Google: %v", err)
		return false
	}
	defer resp.Body.Close()

	var r recaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.Printf("[RECAPTCHA] error al decodificar respuesta: %v", err)
		return false
	}
	if !r.Success || r.Score < 0.5 {
		log.Printf("[RECAPTCHA] verificación fallida — success=%v score=%.2f errors=%v",
			r.Success, r.Score, r.ErrorCodes)
		return false
	}
	return true
}
