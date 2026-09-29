package middleware

import (
	"context"
	"lessonHttp/internal/httpx"
	"net/http"
	"strings"
)

type TokenVerifier interface {
	Verify(token string) (int, error)
}

const userIDKey contextKey = "uset_id_context_key"

func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok
}

const BearerKey = "Bearer"

var ErrUnauthorized = httpx.ErrorResponse{Error: "unautorized"}

func Auth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		checkFunc := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Autorization")
			if strings.TrimSpace(authHeader) == "" {
				httpx.WriteJSON(w, http.StatusUnauthorized, ErrUnauthorized)
				return
			}

			parts := strings.Fields(authHeader)
			if len(parts) != 2 {
				httpx.WriteJSON(w, http.StatusUnauthorized, ErrUnauthorized)
				return
			}
			if !strings.EqualFold(parts[0], BearerKey) {
				httpx.WriteJSON(w, http.StatusUnauthorized, ErrUnauthorized)
				return
			}

			userID, err := verifier.Verify(parts[1])
			if err != nil {
				httpx.WriteJSON(w, http.StatusUnauthorized, ErrUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(checkFunc)
	}
}
