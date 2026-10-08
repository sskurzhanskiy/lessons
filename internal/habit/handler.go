package habit

import (
	"encoding/json"
	"errors"
	"lessonHttp/internal/httpx"
	"lessonHttp/internal/middleware"
	"net/http"
	"strconv"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(sevice ServiceInterface) *Handler {
	return &Handler{
		service: sevice,
	}
}

func (h *Handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials)
		return
	}

	var input CreateInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		writeError(w, ErrInvalidParameter)
		return
	}

	habit, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, habit)
}

var defaultHabitLimit = 10
var maxHabitLimit = 50

func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials)
		return
	}

	pagination, err := httpx.ParsePagination(r, defaultHabitLimit, maxHabitLimit)
	if err != nil {
		writeError(w, ErrInvalidParameter)
		return
	}

	habits, err := h.service.List(r.Context(), userID, pagination.Limit, pagination.Offset)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, habits)
}

func (h *Handler) ByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials)
		return
	}

	habitID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalidParameter)
		return
	}

	habit, err := h.service.ByID(r.Context(), userID, habitID)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, habit)
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials)
		return
	}

	habitID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalidParameter)
		return
	}

	var input UpdateInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		writeError(w, ErrInvalidParameter)
		return
	}

	habit, err := h.service.Update(r.Context(), userID, habitID, input)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, habit)
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials)
		return
	}
	habitID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalidParameter)
		return
	}

	err = h.service.Delete(r.Context(), userID, habitID)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		httpx.WriteError(w, http.StatusBadRequest, validationErr.Error())
	case errors.Is(err, ErrInvalidParameter):
		httpx.WriteError(w, http.StatusBadRequest, ErrInvalidParameter.Error())
	case errors.Is(err, ErrInvalidCredentials):
		httpx.WriteError(w, http.StatusUnauthorized, ErrInvalidCredentials.Error())
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, ErrNotFound.Error())

	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
