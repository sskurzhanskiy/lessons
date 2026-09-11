package middleware

import (
	"lessonHttp/internal/httpx"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()

				logger.Error(
					"panic recovered",
					"panic", rec,
					"request_id", RequestIDFromContext(r.Context()),
					"stack", string(stack),
				)

				httpx.WriteJSON(w,
					http.StatusInternalServerError,
					httpx.ErrorResponse{Error: "internal server error"},
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
