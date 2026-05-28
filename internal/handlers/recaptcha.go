package handlers

import (
	"encoding/json"
	"log/slog"
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
		slog.Warn("reCAPTCHA: token vacío recibido")
		return true
	}
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {secret},
		"response": {token},
	})
	if err != nil {
		slog.Error("reCAPTCHA: error al contactar Google", "error", err)
		return false
	}
	defer resp.Body.Close()

	var r recaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		slog.Error("reCAPTCHA: error al decodificar respuesta", "error", err)
		return false
	}
	if !r.Success || r.Score < 0.5 {
		slog.Warn("reCAPTCHA: verificación fallida", "success", r.Success, "score", r.Score, "errors", r.ErrorCodes)
		return false
	}
	return true
}
