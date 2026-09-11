package middleware

import (
	"encoding/json"
	"lessonHttp/internal/httpx"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRecovery(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	RequestID(Recovery(logger, h)).ServeHTTP(rec, req)

	var response httpx.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("response status code %d; want code = %d", rec.Code, http.StatusInternalServerError)
	}

	if response.Error != "internal server error" {
		t.Errorf("error: %q; want %q", response.Error, "internal server error")
	}
}
