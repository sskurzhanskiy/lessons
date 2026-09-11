package user

import (
	"encoding/json"
	httpx "lessonHttp/internal/httpx"
	"net/http"
	"strconv"
)

type CreateRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var createUser CreateRequest
	err := json.NewDecoder(r.Body).Decode(&createUser)
	if err != nil {
		writeError(w, &ValidationError{Field: "parameters not correct"})
		return
	}

	user, err := h.service.Create(r.Context(), createUser.Name, createUser.Age)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, user)
}

func (h *Handler) UserByIDHandler(w http.ResponseWriter, r *http.Request) {
	rawUserID := r.PathValue("id")
	id, err := strconv.Atoi(rawUserID)
	if err != nil {
		writeError(w, &ValidationError{Field: "id"})
		return
	}

	user, err := h.service.ByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, user)
}
