package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// bindError convierte errores de ShouldBindJSON en respuestas HTTP legibles.
// Para errores de validación retorna {"errors": {"campo": "mensaje en español"}}.
// Para otros errores (JSON malformado, etc.) retorna {"error": "datos inválidos"}.
func bindError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make(map[string]string, len(ve))
		for _, e := range ve {
			out[jsonFieldName(e)] = validationMessage(e)
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": out})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos"})
}

// jsonFieldName convierte el nombre de campo del validador al formato snake_case.
func jsonFieldName(e validator.FieldError) string {
	return camelToSnake(e.Field())
}

// validationMessage retorna el mensaje de error en español para cada regla de validación.
func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "Este campo es obligatorio"
	case "email":
		return "Ingresa un correo electrónico válido"
	case "min":
		return fmt.Sprintf("Debe tener al menos %s caracteres", e.Param())
	case "max":
		return fmt.Sprintf("No puede superar %s caracteres", e.Param())
	case "len":
		return fmt.Sprintf("Debe tener exactamente %s caracteres", e.Param())
	case "uuid", "uuid4":
		return "Debe ser un UUID válido"
	case "oneof":
		return fmt.Sprintf("Debe ser uno de: %s", strings.ReplaceAll(e.Param(), " ", ", "))
	case "url":
		return "Debe ser una URL válida"
	case "numeric":
		return "Debe ser un valor numérico"
	case "alphanum":
		return "Solo se permiten letras y números"
	case "gte":
		return fmt.Sprintf("El valor mínimo permitido es %s", e.Param())
	case "lte":
		return fmt.Sprintf("El valor máximo permitido es %s", e.Param())
	default:
		return fmt.Sprintf("Valor inválido (regla: %s)", e.Tag())
	}
}

// camelToSnake convierte CamelCase → snake_case para mapear nombres de struct a JSON.
func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + 32)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
