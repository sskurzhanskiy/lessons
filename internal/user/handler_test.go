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

	IsRegisterCall bool
	RegistredUser  User
	SavedInput     RegisterInput

	IsAuthCall bool
	UserID     int
	AuthErr    error
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
	s.IsRegisterCall = true
	s.SavedInput = input

	return s.RegistredUser, s.Err
}

func (s *FakeUserService) Login(ctx context.Context, email string, password string) (int, error) {
	s.IsAuthCall = true
	return s.UserID, s.AuthErr
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
		input          string
		IsCall         bool
		wantStatusCode int
		wantErr        error
	}{
		{
			name: "parameters not correct",
			input: `{"name":"Alice",
					"age":27,
					"email":4567,
					"passw":"3erfghbhj9o"}`,
			wantStatusCode: 400,
			wantErr:        &ValidationError{Field: "parameters not correct"},
		},
		{
			name: "input data not correct",
			input: `{"name":"",
					"age":27,
					"email":"test@test.vom",
					"password":"3erfghbhj9o"}`,
			wantStatusCode: 400,
			IsCall:         true,
			wantErr:        &ValidationError{Field: "name"},
		},
		{
			name: "Email Already Exists",
			input: `{"name":"",
					"age":27,
					"email":"test@test.vom",
					"password":"3erfghbhj9o"}`,
			wantStatusCode: 409,
			IsCall:         true,
			wantErr:        ErrEmailAlreadyExists,
		},
		{
			name: "correct register",
			input: `{"name":"Alice",
					"age":27,
					"email":"test@test.vom",
					"password":"3erfghbhj9o"}`,
			IsCall:         true,
			wantStatusCode: 201,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rUser := User{
				ID:   1,
				Name: "Alice",
				Age:  27,
			}
			service := FakeUserService{
				Err:            tt.wantErr,
				RegistredUser:  rUser,
				IsRegisterCall: tt.IsCall,
			}
			handler := NewHandler(&service)

			mux := http.NewServeMux()
			mux.HandleFunc("/auth/register", handler.RegisterHandler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte(tt.input)))

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Fatalf("request status code %d; want %d", rec.Code, tt.wantStatusCode)
			}

			if tt.wantErr != nil {
				var response httpx.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				if err != nil {
					t.Fatalf("decode response %v", err)
				}
				if response.Error != tt.wantErr.Error() {
					t.Errorf("got %q; expected %q", response.Error, tt.wantErr.Error())
				}

				return
			}

			if service.IsRegisterCall != tt.IsCall {
				t.Errorf("register is call; expected register not call")
			}

			var response User
			err := json.NewDecoder(rec.Body).Decode(&response)
			if err != nil {
				t.Fatalf("decode recponse %v", err)
			}

			if response.ID != rUser.ID {
				t.Errorf("response user ID %d; want %d", response.ID, rUser.ID)
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

func TestLoginHandler(t *testing.T) {
	var tests = []struct {
		name            string
		body            string
		IsServiceCall   bool
		AuthErr         error
		wantStatusCode  int
		wantResponseErr string
	}{
		// {
		// 	name: "parameters not correct",
		// 	body: `{"emaidsad":"email@email.com",
		// 			"passwor":"password"}`,
		// 	IsServiceCall:   false,
		// 	AuthErr:         &ValidationError{Field: "parameters not correct"},
		// 	wantStatusCode:  400,
		// 	wantResponseErr: "parameters not correct",
		// },
		{
			name: "service validation error",
			body: `{"email":"",
					"password":"12345"}`,
			IsServiceCall:   true,
			AuthErr:         ErrInvalidCredentials,
			wantStatusCode:  400,
			wantResponseErr: ErrInvalidCredentials.Error(),
		},
		{
			name: "service email not found",
			body: `{"email":"",
					"password":"12345"}`,
			IsServiceCall:   true,
			AuthErr:         ErrNotFound,
			wantStatusCode:  404,
			wantResponseErr: ErrNotFound.Error(),
		},
		{
			name: "service internal server error",
			body: `{"email":"",
					"password":"12345"}`,
			IsServiceCall:   true,
			AuthErr:         fmt.Errorf("some error"),
			wantStatusCode:  500,
			wantResponseErr: "internal server error",
		},
		{
			name: "request correct",
			body: `{"email":"email@email.com",
					"password":"12345"}`,
			IsServiceCall:  true,
			wantStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credentials := Credentials{UserID: 1}
			service := FakeUserService{
				AuthErr: tt.AuthErr,
				UserID:  1,
			}
			handler := NewHandler(&service)

			mux := http.NewServeMux()
			mux.HandleFunc("/auth/login", handler.LoginHandler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer([]byte(tt.body)))
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Fatalf("status code %d; want %d", rec.Code, tt.wantStatusCode)
			}

			if service.IsAuthCall != tt.IsServiceCall {
				t.Errorf("service call authorization %t;want %t", service.IsAuthCall, tt.IsServiceCall)
			}
			if tt.wantResponseErr != "" {
				var responseErr *httpx.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&responseErr)
				if err != nil {
					t.Fatalf("decoded error response %v", err)
				}
				if responseErr.Error != tt.wantResponseErr {
					t.Errorf("response error %v; want %v", responseErr.Error, tt.wantResponseErr)
				}
				return
			}

			var response Credentials
			err := json.NewDecoder(rec.Body).Decode(&response)
			if err != nil {
				t.Fatalf("decoded response %v", err)
			}

			if response.UserID != credentials.UserID {
				t.Fatalf("response user ID %d;want %d", response.UserID, credentials.UserID)
			}
		})
	}
}
