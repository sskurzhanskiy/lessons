package user

import (
	"encoding/json"
	"errors"
	httpx "lessonHttp/internal/httpx"
	"lessonHttp/internal/middleware"
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

var defaultLimit = 10
var maxLimit = 100

func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	pagination, err := httpx.ParsePagination(r, defaultLimit, maxLimit)
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

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		writeError(w, &ValidationError{Field: "parameters not correct"})
		return
	}

	token, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, LoginResponse{
		AccessToken: token,
	})
}

func (h *Handler) UserByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, &ValidationError{Field: "id"})
		return
	}

	requesterID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials)
		return
	}

	user, err := h.service.ByID(r.Context(), requesterID, userID)
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
		httpx.WriteError(w, http.StatusBadRequest, validationErr.Error())
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, ErrNotFound.Error())
	case errors.Is(err, ErrInvalidParameter):
		httpx.WriteError(w, http.StatusBadRequest, ErrInvalidParameter.Error())
	case errors.Is(err, ErrInvalidCredentials):
		httpx.WriteError(w, http.StatusUnauthorized, ErrInvalidCredentials.Error())
	case errors.Is(err, ErrEmailAlreadyExists):
		httpx.WriteError(w, http.StatusConflict, ErrEmailAlreadyExists.Error())

	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
