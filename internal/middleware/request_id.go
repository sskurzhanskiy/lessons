package middleware

import (
	"context"
	"net/http"
	"sync/atomic"
)

type contextKey string

const requestIDKey contextKey = "context_request_id_key"

var requestCounter atomic.Uint64

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestCounter.Add(1)
		ctx := context.WithValue(r.Context(), requestIDKey, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDFromContext(ctx context.Context) uint64 {
	value := ctx.Value(requestIDKey)

	raw, ok := value.(uint64)

	if !ok {
		return 0
	}

	return uint64(raw)
}
