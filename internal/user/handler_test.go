package user

import (
	"context"
	"encoding/json"
	"errors"
	"lessonHttp/internal/httpx"
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

type FakeUserService struct {
	Users []User
	Err   error

	ReceivedLimit  int
	ReceivedOffset int
}

func (s *FakeUserService) Create(ctx context.Context, name string, age int) (User, error) {
	return User{}, nil
}
func (s *FakeUserService) ByID(ctx context.Context, id int) (User, error) {
	return User{}, nil
}

func (s *FakeUserService) List(ctx context.Context, limit int, offset int) ([]User, error) {
	s.ReceivedLimit = limit
	s.ReceivedOffset = offset

	return s.Users, s.Err
}

func TestListHandlerSuccess(t *testing.T) {
	service := FakeUserService{
		Users: []User{
			{
				ID:   1,
				Name: "Alice",
				Age:  27,
			},
			{
				ID:   2,
				Name: "Bob",
				Age:  15,
			},
		},
	}
	handler := NewHandler(&service)

	mux := http.NewServeMux()
	mux.HandleFunc("/users", handler.ListHandler)

	req := httptest.NewRequest(http.MethodGet, "/users?limit=2", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code %d; want %d", rec.Code, http.StatusOK)
	}

	var users []User
	if err := json.NewDecoder(rec.Body).Decode(&users); err != nil {
		t.Fatalf("decoder failed %v", err)
	}

	if len(users) != len(service.Users) {
		t.Errorf("response have %d users; want = %d", len(users), len(service.Users))
	}

	for i := 0; i < len(service.Users); i++ {
		if users[i].ID != service.Users[i].ID {
			t.Errorf("response user.ID = %d; want = %d", users[i].ID, service.Users[i].ID)
		}
		if users[i].Name != service.Users[i].Name {
			t.Errorf("response user.Name = %q; want = %q", users[i].Name, service.Users[i].Name)
		}
		if users[i].Age != service.Users[i].Age {
			t.Errorf("response user.Age = %d; want = %d", users[i].Age, service.Users[i].Age)
		}
	}
}

func TestListHandlerParseParam(t *testing.T) {
	service := FakeUserService{}
	handler := NewHandler(&service)
	mux := http.NewServeMux()
	mux.HandleFunc("/users", handler.ListHandler)

	req := httptest.NewRequest(http.MethodGet, "/users?limit=2&offset=5", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if service.ReceivedLimit != 2 {
		t.Errorf("pagination pasre limit = %d;want = 2", service.ReceivedLimit)
	}
	if service.ReceivedOffset != 5 {
		t.Errorf("pagination pasre offset = %d;want = 5", service.ReceivedOffset)
	}
}

func TestListHandlerInvalidParams(t *testing.T) {
	service := FakeUserService{}
	handler := NewHandler(&service)
	mux := http.NewServeMux()
	mux.HandleFunc("/users", handler.ListHandler)

	req := httptest.NewRequest(http.MethodGet, "/users?limit=0", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code %d; want = %d", rec.Code, http.StatusBadRequest)
	}

	var err httpx.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&err); err != nil {
		t.Fatalf("decoder failed %v", err)
	}

	if err.Error != ErrInvalidParameter.Error() {
		t.Errorf("response error %v; want %v", err.Error, ErrInvalidParameter.Error())
	}
}

func TestListHandlerServiceError(t *testing.T) {
	service := FakeUserService{
		Err: errors.New("fake service error"),
	}
	handler := NewHandler(&service)
	mux := http.NewServeMux()
	mux.HandleFunc("/users", handler.ListHandler)

	req := httptest.NewRequest(http.MethodGet, "/users?limit=10", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status code %d; want = %d", rec.Code, http.StatusInternalServerError)
	}

	var response httpx.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decoder failed %v", err)
	}

	if response.Error != "internal server error" {
		t.Errorf("response error %v; want %v", response.Error, "internal server error")
	}
}
