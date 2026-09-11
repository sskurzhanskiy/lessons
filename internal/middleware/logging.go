package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(p)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.status = statusCode
	w.wroteHeader = true

	w.ResponseWriter.WriteHeader(statusCode)
}

func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		started := time.Now()
		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(rw, r)

		requestID := RequestIDFromContext(r.Context())

		logger.Info(
			"request completed",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(started),
		)
	})
}
