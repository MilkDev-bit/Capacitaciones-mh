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

func verifyRecaptcha(token string) bool {
	secret := os.Getenv("RECAPTCHA_SECRET_KEY")
	if secret == "" {
		return true
	}
	if token == "" {
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
