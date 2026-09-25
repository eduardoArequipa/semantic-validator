package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

func Middleware(expectedAPIKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		provided := ""
		if strings.HasPrefix(authorization, "Bearer ") {
			provided = strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		}
		if !matches(provided, expectedAPIKey) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "unauthorized",
					"message": "valid Semantic Validator API key required",
				},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func matches(provided, expected string) bool {
	providedHash := sha256.Sum256([]byte(provided))
	expectedHash := sha256.Sum256([]byte(expected))
	return expected != "" && subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) == 1
}
