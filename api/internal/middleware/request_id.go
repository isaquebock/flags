package middleware

import (
	"context"
	"net/http"
)

const requestIDKey = "request_id"

func GetRequestID(r *http.Request) string {
	return r.Header.Get("X-Request-Id")
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) string {
	if id := ctx.Value(requestIDKey); id != nil {
		return id.(string)
	}
	return ""
}
