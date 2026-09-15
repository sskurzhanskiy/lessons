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

type Pagination struct {
	Limit  int
	Offset int
}

const (
	defaultLimit  = 20
	minLimit      = 1
	maxLimit      = 100
	defaultOffset = 0
	minOffset     = 0
	maxOffset     = -1
)

var ErrInvalidParameter = errors.New("invalid parameter")

func parsePagination(r *http.Request) (Pagination, error) {
	limit, err := parseParameter(r.URL.Query().Get("limit"),
		defaultLimit,
		minLimit,
		maxLimit,
	)
	if err != nil {
		return Pagination{}, err
	}

	offset, err := parseParameter(r.URL.Query().Get("offset"),
		defaultOffset,
		minOffset,
		maxOffset,
	)
	if err != nil {
		return Pagination{}, err
	}

	return Pagination{
		Limit:  limit,
		Offset: offset,
	}, nil
}

func parseParameter(param string, defaultValue int, minValue int, maxValue int) (int, error) {
	if param == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, ErrInvalidParameter
	}

	if value < minValue {
		return 0, ErrInvalidParameter
	}

	if maxValue <= 0 {
		return value, nil
	}

	if value > maxValue {
		return 0, ErrInvalidParameter
	}

	return value, nil
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {

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
