package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	requestID := performRequest()
	if requestID == 0 {
		t.Errorf("request ID %d; want id != 0", requestID)
	}

	firstRequestID := performRequest()
	secondRequestID := performRequest()

	if firstRequestID == secondRequestID {
		t.Error("requestID1 == requestID2; want !=")
	}
}

func performRequest() uint64 {
	var requestID uint64

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID = RequestIDFromContext(r.Context())
	})
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	RequestID(h).ServeHTTP(rec, req)

	return requestID
}
