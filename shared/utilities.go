package shared

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrEmptyJWTId = errors.New("erro ao extrair id do usuario")
var ErrInvalidJWT = errors.New("JWT inválido")

func ExtractIDFromJWT(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", ErrInvalidJWT
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("erro on decode jwt. err: %v", err)
	}
	var claims map[string]any
	json.Unmarshal(payload, &claims)
	id, ok := claims["id"]
	if ok {
		return fmt.Sprintf("%.0f", id.(float64)), nil
	}
	return "", ErrEmptyJWTId
}

func AllowOrigin(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
}
