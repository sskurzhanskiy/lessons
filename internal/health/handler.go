package health

import (
	"lessonHttp/internal/httpx"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (*Handler) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, HealthResponse{
		Status: "ok",
	})
}
