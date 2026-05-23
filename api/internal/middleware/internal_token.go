package middleware

import (
	"net/http"

	"github.com/isaquebock/flags-api/internal/handlers"
)

func InternalTokenMiddleware(expectedToken string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Internal-Token")

			if token == "" {
				handlers.WriteError(w, http.StatusUnauthorized, "missing_internal_token", "X-Internal-Token header is required")
				return
			}

			if token != expectedToken {
				handlers.WriteError(w, http.StatusUnauthorized, "invalid_internal_token", "Invalid X-Internal-Token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
