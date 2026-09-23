package user

import (
	"encoding/json"
	"errors"
	httpx "lessonHttp/internal/httpx"
	"net/http"
	"strconv"
)

type CreateRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	pagination, err := parsePagination(r)
	if err != nil {
		writeError(w, err)
		return
	}

	users, err := h.service.List(r.Context(), pagination.Limit, pagination.Offset)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		writeError(w, &ValidationError{Field: "parameters not correct"})
		return
	}

	user, err := h.service.Register(r.Context(), input)
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

func writeError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		httpx.WriteJSON(w,
			http.StatusBadRequest,
			httpx.ErrorResponse{Error: validationErr.Error()},
		)
	case errors.Is(err, ErrNotFound):
		httpx.WriteJSON(w,
			http.StatusNotFound,
			httpx.ErrorResponse{Error: ErrNotFound.Error()},
		)
	case errors.Is(err, ErrInvalidParameter):
		httpx.WriteJSON(w,
			http.StatusBadRequest,
			httpx.ErrorResponse{Error: ErrInvalidParameter.Error()},
		)
	case errors.Is(err, ErrEmailAlreadyExists):
		httpx.WriteJSON(w,
			http.StatusConflict,
			httpx.ErrorResponse{Error: ErrEmailAlreadyExists.Error()})

	default:
		httpx.WriteJSON(w,
			http.StatusInternalServerError,
			httpx.ErrorResponse{Error: "internal server error"},
		)
	}
}
