package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"lessonHttp/internal/httpx"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func (s *FakeUserService) Register(ctx context.Context, input RegisterInput) (User, error) {
	if len(s.Users) >= 1 {
		return s.Users[0], s.Err
	}

	return User{}, s.Err
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

func TestRegisterHandler(t *testing.T) {
	var tests = []struct {
		name           string
		input          RegisterInput
		wantStatusCode int
		wantErr        error
	}{
		{
			name: "correct register",
			input: RegisterInput{
				Name:     "Alice",
				Age:      27,
				Email:    "tre@ghj.com",
				Password: "12345",
			},
			wantStatusCode: 201,
		},
		{
			name: "bad request",
			input: RegisterInput{
				Name:     "",
				Age:      27,
				Email:    "tre@ghj.com",
				Password: "12345",
			},
			wantStatusCode: 400,
			wantErr:        &ValidationError{Field: "name"},
		},
		{
			name: "internal server error",
			input: RegisterInput{
				Name:     "",
				Age:      27,
				Email:    "tre@ghj.com",
				Password: "12345",
			},
			wantStatusCode: 500,
			wantErr:        fmt.Errorf("internal server error"),
		},
		{
			name: "Email Already Exists",
			input: RegisterInput{
				Name:     "",
				Age:      27,
				Email:    "tre@ghj.com",
				Password: "12345",
			},
			wantStatusCode: 409,
			wantErr:        ErrEmailAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rUser := User{
				ID:   1,
				Name: tt.input.Name,
				Age:  tt.input.Age,
			}
			service := FakeUserService{
				Users: []User{
					rUser,
				},
				Err: tt.wantErr,
			}
			handler := NewHandler(&service)

			mux := http.NewServeMux()
			mux.HandleFunc("/auth/register", handler.RegisterHandler)

			jsonInput, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("marshal input %v", err)
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonInput))

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Fatalf("request status code %d; want %d", rec.Code, tt.wantStatusCode)
			}

			if tt.wantErr != nil {
				var response httpx.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				if err != nil {
					t.Fatalf("decode recponse %v", err)
				}
				if response.Error != tt.wantErr.Error() {
					t.Errorf("got %q; expected %q", response.Error, tt.wantErr.Error())
				}
				return
			}

			var response User
			err = json.NewDecoder(rec.Body).Decode(&response)
			if err != nil {
				t.Fatalf("decode recponse %v", err)
			}

			if response.Name != rUser.Name {
				t.Errorf("response user name %q; want %q", response.Name, rUser.Name)
			}
			if response.Age != rUser.Age {
				t.Errorf("response user age %d; want %d", response.Age, rUser.Age)
			}
		})
	}
}
