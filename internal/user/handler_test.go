package user

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

/*
func TestGetUser(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantUser   *User
		wantError  string
	}{
		{
			name:       "user found",
			path:       "/users/1",
			wantStatus: http.StatusOK,
			wantUser: &User{
				ID:   1,
				Name: "Alice",
				Age:  30,
			},
		},
		{
			name:       "user not found",
			path:       "/users/999",
			wantStatus: http.StatusNotFound,
			wantError:  "user not found",
		},
		{
			name:       "invalid user id",
			path:       "/users/abc",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid user id",
		},
	}

	repo := NewMemoryRepository()
	service := NewService(repo)
	handler := NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", handler.UserByIDHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				tt.path,
				nil,
			)
			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status code = %d; want = %d", recorder.Code, tt.wantStatus)
			}

			contentType := recorder.Header().Get("Content-Type")
			wantContentType := "application/json"
			if contentType != wantContentType {
				t.Errorf("Content-Type = %q; want = %q", contentType, wantContentType)
			}

			if tt.wantUser != nil {
				var response User
				err := json.NewDecoder(recorder.Body).Decode(&response)
				if err != nil {
					t.Fatalf("decoder response %v", err)
				}
				if response.ID != tt.wantUser.ID {
					t.Errorf("user ID = %d; want = %d", response.ID, tt.wantUser.ID)
				}
				if response.Name != tt.wantUser.Name {
					t.Errorf("user Name = %q; want = %q", response.Name, tt.wantUser.Name)
				}
				if response.Age != tt.wantUser.Age {
					t.Errorf("user Age = %d; want = %d", response.Age, tt.wantUser.Age)
				}
			}

			if tt.wantError != "" {
				var response ErrorResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				if err != nil {
					t.Fatalf("decoder response %v", err)
				}
				if response.Err != tt.wantError {
					t.Errorf("response error = %q; want = %q", response.Err, tt.wantError)
				}
			}
		})
	}
}
*/

func TestParseParameter(t *testing.T) {
	var tests = []struct {
		name         string
		param        string
		defaultValue int
		minValue     int
		maxValue     int
		wantValue    int
		wantErr      error
	}{
		{
			name:         "limit: default values",
			param:        "",
			defaultValue: defaultLimit,
			minValue:     minLimit,
			maxValue:     maxLimit,
			wantValue:    defaultLimit,
		},
		{
			name:         "limit: correct values",
			param:        "10",
			defaultValue: defaultLimit,
			minValue:     minLimit,
			maxValue:     maxLimit,
			wantValue:    10,
		},
		{
			name:         "limit: not correct",
			param:        "fghjk",
			defaultValue: defaultLimit,
			minValue:     minLimit,
			maxValue:     maxLimit,
			wantValue:    0,
			wantErr:      ErrInvalidParameter,
		},
		{
			name:         "limit: low zero values",
			param:        "-5",
			defaultValue: defaultLimit,
			minValue:     minLimit,
			maxValue:     maxLimit,
			wantValue:    0,
			wantErr:      ErrInvalidParameter,
		},
		{
			name:         "limit: high limit value",
			param:        "500",
			defaultValue: defaultLimit,
			minValue:     minLimit,
			maxValue:     maxLimit,
			wantValue:    0,
			wantErr:      ErrInvalidParameter,
		},
		{
			name:         "limit: low limit value",
			param:        "0",
			defaultValue: defaultLimit,
			minValue:     minLimit,
			maxValue:     maxLimit,
			wantValue:    0,
			wantErr:      ErrInvalidParameter,
		},
		{
			name:         "limit: low limit border",
			param:        "1",
			defaultValue: defaultLimit,
			minValue:     minLimit,
			maxValue:     maxLimit,
			wantValue:    1,
		},
		{
			name:         "limit: high limit border",
			param:        "100",
			defaultValue: defaultLimit,
			minValue:     minLimit,
			maxValue:     maxLimit,
			wantValue:    100,
		},
		{
			name:         "limit: upper limit",
			param:        "101",
			defaultValue: defaultLimit,
			minValue:     minLimit,
			maxValue:     maxLimit,
			wantValue:    0,
			wantErr:      ErrInvalidParameter,
		},
		{
			name:         "offset: default values",
			param:        "",
			defaultValue: defaultOffset,
			minValue:     minOffset,
			maxValue:     maxOffset,
			wantValue:    defaultOffset,
		},
		{
			name:         "offset: correct values",
			param:        "1",
			defaultValue: defaultOffset,
			minValue:     minOffset,
			maxValue:     maxOffset,
			wantValue:    1,
		},
		{
			name:         "offset: not correct values",
			param:        "rfgbn",
			defaultValue: defaultOffset,
			minValue:     minOffset,
			maxValue:     maxOffset,
			wantValue:    0,
			wantErr:      ErrInvalidParameter,
		},
		{
			name:         "offset: low limit value",
			param:        "-1",
			defaultValue: defaultOffset,
			minValue:     minOffset,
			maxValue:     maxOffset,
			wantValue:    0,
			wantErr:      ErrInvalidParameter,
		},
		{
			name:         "offset: low limit border",
			param:        "0",
			defaultValue: defaultOffset,
			minValue:     minOffset,
			maxValue:     maxOffset,
			wantValue:    0,
		},
		{
			name:         "offset: no upper limit",
			param:        "500",
			defaultValue: defaultOffset,
			minValue:     minOffset,
			maxValue:     maxOffset,
			wantValue:    500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := parseParameter(tt.param, tt.defaultValue, tt.minValue, tt.maxValue)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error %v; want = %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if value != tt.wantValue {
				t.Errorf("value = %d; want = %d", value, tt.wantValue)
			}
		})
	}
}

func TestParsePagination(t *testing.T) {
	var tests = []struct {
		name           string
		target         string
		wantPagination Pagination
		wantErr        error
	}{
		{
			name:   "default",
			target: "/users",
			wantPagination: Pagination{
				Limit:  defaultLimit,
				Offset: defaultOffset,
			},
		},
		{
			name:   "correct",
			target: "/users?limit=10&offset=20",
			wantPagination: Pagination{
				Limit:  10,
				Offset: 20,
			},
		},
		{
			name:   "offset default",
			target: "/users?limit=30",
			wantPagination: Pagination{
				Limit:  30,
				Offset: 0,
			},
		},
		{
			name:   "limit default",
			target: "/users?offset=5",
			wantPagination: Pagination{
				Limit:  defaultLimit,
				Offset: 5,
			},
		},
		{
			name:   "limit invalid",
			target: "/users?limit=rty&offset=0",
			wantPagination: Pagination{
				Limit:  defaultLimit,
				Offset: defaultOffset,
			},
			wantErr: ErrInvalidParameter,
		},
		{
			name:   "offset invalid",
			target: "/users?limit=15&offset=1rty",
			wantPagination: Pagination{
				Limit:  defaultLimit,
				Offset: defaultOffset,
			},
			wantErr: ErrInvalidParameter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.target, nil)
			p, err := parsePagination(r)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error %v; want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if p.Limit != tt.wantPagination.Limit ||
				p.Offset != tt.wantPagination.Offset {
				t.Errorf("limit=%d offset=%d; want limit=%d offset=%d",
					p.Limit,
					p.Offset,
					tt.wantPagination.Limit,
					tt.wantPagination.Offset,
				)
			}
		})
	}

}
