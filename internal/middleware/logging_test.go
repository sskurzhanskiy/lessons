package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLogging(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	RequestID(Logging(logger, h)).ServeHTTP(rec, req)
}

func TestResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()

	rw := &responseWriter{
		ResponseWriter: rec,
		status:         http.StatusOK,
	}

	if rw.status != rec.Code || rw.status != http.StatusOK {
		t.Errorf("rw status %d; want = 200", rw.status)
	}

	rw.WriteHeader(http.StatusNotFound)
	if rw.status != rec.Code || rw.status != http.StatusNotFound {
		t.Errorf("rw status %d; want = 404", rw.status)
	}

	rw.WriteHeader(http.StatusInternalServerError)
	if rw.status != rec.Code || rw.status != http.StatusNotFound {
		t.Errorf("rw status %d; want = 404", rw.status)
	}
}

func TestResponseWriterWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rec,
	}

	rw.Write([]byte("hello"))
	if rw.status != rec.Code {
		t.Fatal("rw.status and rec.Code must be equal")
	}
	if rw.status != http.StatusOK {
		t.Errorf("rw status %d; want = 200", rw.status)
	}

	rw.WriteHeader(http.StatusInternalServerError)
	if rw.status != rec.Code {
		t.Fatal("rw.status and rec.Code must be equal")
	}
	if rw.status != http.StatusOK {
		t.Errorf("rw status %d; want = 200", rw.status)
	}
}
