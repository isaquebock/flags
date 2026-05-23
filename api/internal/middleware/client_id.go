package middleware

import (
	"context"
	"net/http"

	"github.com/isaquebock/flags-api/internal/handlers"
	"github.com/isaquebock/flags-api/internal/validation"
)

const clientIDKey = "client_id"

func GetClientID(r *http.Request) string {
	return r.Header.Get("X-Client-Id")
}

func WithClientID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, clientIDKey, id)
}

func ClientIDFromContext(ctx context.Context) string {
	if id := ctx.Value(clientIDKey); id != nil {
		return id.(string)
	}
	return ""
}

func ClientIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := GetClientID(r)

		if clientID == "" {
			handlers.WriteError(w, http.StatusBadRequest, "missing_client_id", "X-Client-Id header is required")
			return
		}

		if err := validation.ValidateClientID(clientID); err != nil {
			handlers.WriteError(w, http.StatusBadRequest, "invalid_client_id", err.Error())
			return
		}

		ctx := WithClientID(r.Context(), clientID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
