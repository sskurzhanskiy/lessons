package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type ErrorResponse struct {
	Err string `json:"error"`
}

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
		writeError(w, http.StatusBadRequest, "parameters not correct")
		return
	}

	user, err := h.service.Create(r.Context(), createUser.Name, createUser.Age)
	if err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			writeError(w, http.StatusBadRequest, validationErr.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) UserByIDHandler(w http.ResponseWriter, r *http.Request) {
	rawUserID := r.PathValue("id")
	id, err := strconv.Atoi(rawUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.ByID(r.Context(), id)
	if err != nil {
		var validationErr *ValidationError
		switch {
		case errors.As(err, &validationErr):
			writeError(w, http.StatusBadRequest, validationErr.Error())
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, "user not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		// log
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, ErrorResponse{
		Err: message,
	})
}
