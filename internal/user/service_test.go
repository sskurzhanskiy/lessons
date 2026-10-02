package user

import (
	"context"
	"errors"
	"lessonHttp/internal/password"
	"testing"
)

type FakeUserRepository struct {
	CreateUser User
	CreateErr  error
	ByIDUser   User
	ByIDErr    error

	CreateArg CreateUserParams
	ByIDArg   int

	CreateCalled bool
	ByIDCalled   bool

	Credentials Credentials
	AuthErr     error
	AuthCalled  bool
}

func (r *FakeUserRepository) Create(ctx context.Context, params CreateUserParams) (User, error) {
	r.CreateCalled = true
	r.CreateArg = params
	return r.CreateUser, r.CreateErr
}

func (r *FakeUserRepository) ByID(ctx context.Context, id int) (User, error) {
	r.ByIDCalled = true
	r.ByIDArg = id
	return r.ByIDUser, r.ByIDErr
}

func (r *FakeUserRepository) List(ctx context.Context, limit int, offset int) ([]User, error) {
	return []User{}, nil
}

func (r *FakeUserRepository) ByEmail(ctx context.Context, email string) (Credentials, error) {
	r.AuthCalled = true

	return r.Credentials, r.AuthErr
}

type FakeTokenGenerator struct {
	accessToken string
	generateErr error

	isCallGenerator bool
	userID          int
}

func (g *FakeTokenGenerator) Generate(userID int) (string, error) {
	g.isCallGenerator = true
	g.userID = userID

	if g.generateErr != nil {
		return "", g.generateErr
	}

	return g.accessToken, nil
}

func TestServiceCreate(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{
		CreateUser: User{
			ID:   1,
			Name: "Alice",
			Age:  22,
		},
	}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)

	user, err := service.Create(ctx, "Alice", 22)

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !repo.CreateCalled {
		t.Fatal("repository Create was not called")
	}

	if repo.CreateArg.Name != "Alice" {
		t.Errorf("argument name %q; want Alice", repo.CreateArg.Name)
	}
	if repo.CreateArg.Age != 22 {
		t.Errorf("argument age %d; want 22", repo.CreateArg.Age)
	}

	if user.ID != 1 {
		t.Errorf("user id = %d; want id = 1", user.ID)
	}
	if user.Name != "Alice" {
		t.Errorf("user name = %q; want name = Alice", user.Name)
	}
	if user.Age != 22 {
		t.Errorf("user age = %d; want age = 22", user.Age)
	}
}

func TestServiceCreateInvalidName(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)

	_, err := service.Create(ctx, "", 22)
	if repo.CreateCalled {
		t.Errorf("repository Create was called for invalid input")
	}

	var validationErr *ValidationError
	if err == nil {
		t.Fatal("expected error")
	}

	if errors.As(err, &validationErr) {
		if validationErr.Field != "name" {
			t.Errorf("validation error field = %q; want field = name", validationErr.Field)
		}
	} else {
		t.Fatalf("error %T; want *ValidationError", err)
	}
}

func TestServiceCreateInvalidAge(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)

	_, err := service.Create(ctx, "Alice", 0)
	if repo.CreateCalled {
		t.Errorf("repository Create was called for invalid input")
	}

	var validationErr *ValidationError
	if err == nil {
		t.Fatal("expected error")
	}

	if errors.As(err, &validationErr) {
		if validationErr.Field != "age" {
			t.Errorf("validation error field = %q; want field = age", validationErr.Field)
		}
	} else {
		t.Fatalf("error %T; want *ValidationError", err)
	}
}

func TestServiceCreatePropagationError(t *testing.T) {
	var errRepo = errors.New("somewhere done error")
	ctx := context.Background()
	repo := &FakeUserRepository{
		CreateErr: errRepo,
	}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)
	_, err := service.Create(ctx, "Alice", 31)
	if err == nil {
		t.Fatal("expected error")
	}

	if !repo.CreateCalled {
		t.Fatal("repository Create was not called")
	}

	if !errors.Is(err, errRepo) {
		t.Errorf("unexpected error %v; want %v", err, errRepo)
	}
}

func TestServiceByIDInvalidID(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)

	_, err := service.ByID(ctx, 1, -1)
	if repo.ByIDCalled {
		t.Error("repository ByID was called for invalid input")
	}

	var validationErr *ValidationError
	if err == nil {
		t.Fatal("expected error")
	}

	if errors.As(err, &validationErr) {
		if validationErr.Field != "id" {
			t.Errorf("validation error field = %q; want = id", validationErr.Field)
		}
	} else {
		t.Fatalf("error %T; want *ValidationError", err)
	}
}
func TestServiceByIDNotFound(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{
		ByIDErr: ErrNotFound,
	}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)

	_, err := service.ByID(ctx, 999, 999)

	if err == nil {
		t.Fatal("expected error")
	}
	if !repo.ByIDCalled {
		t.Fatal("repository ByID was not called")
	}

	if repo.ByIDArg != 999 {
		t.Errorf("argument ByID = %d; want 999", repo.ByIDArg)
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error is %v, want %v", err, ErrNotFound)
	}
}

func TestServiceByIDHappyPath(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{
		ByIDUser: User{
			ID:   1,
			Name: "Bob",
			Age:  35,
		},
	}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)
	user, err := service.ByID(ctx, 1, 1)

	if err != nil {
		t.Fatal("unexpected error")
	}
	if !repo.ByIDCalled {
		t.Fatal("repository ByID was not called")
	}
	if repo.ByIDArg != 1 {
		t.Errorf("argument id %d; want = 1", repo.ByIDArg)
	}
	if user.ID != 1 {
		t.Errorf("user id = %d; want id = 1", user.ID)
	}
	if user.Name != "Bob" {
		t.Errorf("user name = %q; want name = Bob", user.Name)
	}
	if user.Age != 35 {
		t.Errorf("user age = %d; want age = 35", user.Age)
	}
}

func TestServiceByIDPropagationError(t *testing.T) {
	errRepo := errors.New("repository error")
	ctx := context.Background()
	repo := &FakeUserRepository{
		ByIDErr: errRepo,
	}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)

	_, err := service.ByID(ctx, 10, 10)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errRepo) {
		t.Errorf("error = %v; want wrapped %v", err, errRepo)
	}
}

func TestServiceByIDRequesterIncorrect(t *testing.T) {
	ctx := context.Background()
	repo := &FakeUserRepository{}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)

	requesterID := 5
	userID := 10
	_, err := service.ByID(ctx, requesterID, userID)
	if err == nil {
		t.Fatalf("expected error")
	}

	if repo.ByIDCalled {
		t.Error("repo must not call method ByID")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v; want %v", err, ErrNotFound)
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name    string
		input   RegisterInput
		wantErr error
	}{
		{
			name: "correct input",
			input: RegisterInput{
				Name:     "Alice",
				Age:      27,
				Email:    "email@email.com",
				Password: "12345",
			},
		},
		{
			name: "incorrect name",
			input: RegisterInput{
				Name:     "",
				Age:      27,
				Email:    "email@email.com",
				Password: "12345",
			},
			wantErr: &ValidationError{Field: "name"},
		},
		{
			name: "incorrect age",
			input: RegisterInput{
				Name:     "Alice",
				Age:      0,
				Email:    "email@email.com",
				Password: "12345",
			},
			wantErr: &ValidationError{Field: "age"},
		},
		{
			name: "incorrect email",
			input: RegisterInput{
				Name:     "Alice",
				Age:      27,
				Email:    "",
				Password: "12345",
			},
			wantErr: &ValidationError{Field: "email"},
		},
		{
			name: "incorrect password",
			input: RegisterInput{
				Name:     "Alice",
				Age:      27,
				Email:    "email@email.com",
				Password: "",
			},
			wantErr: &ValidationError{Field: "password"},
		},
	}

	wantUser := User{
		ID:   1,
		Name: "Alice",
		Age:  12,
	}
	ctx := context.Background()
	repo := &FakeUserRepository{
		CreateUser: wantUser,
	}
	tGenerator := &FakeTokenGenerator{}
	service := NewService(repo, tGenerator)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Register(ctx, tt.input)
			if tt.wantErr != nil {
				var validationErr *ValidationError
				if !errors.As(err, &validationErr) {
					t.Errorf("register error %v; want %v", err, validationErr)
				}
				var wantErr *ValidationError
				if !errors.As(err, &wantErr) {
					t.Errorf("register error %v; want %v", err, wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if !repo.CreateCalled {
				t.Fatal("repo::Create don't call")
			}

			if repo.CreateArg.Name != tt.input.Name {
				t.Fatalf("repo::Create got name %q; expected %q", repo.CreateArg.Name, tt.input.Name)
			}
			if repo.CreateArg.Age != tt.input.Age {
				t.Fatalf("repo::Create got age %d; expected %d", repo.CreateArg.Age, tt.input.Age)
			}
			if repo.CreateArg.Email != tt.input.Email {
				t.Fatalf("repo::Create got email %q; expected %q", repo.CreateArg.Email, tt.input.Email)
			}

			isVerify, err := password.Verify(tt.input.Password, repo.CreateArg.PasswordHash)
			if err != nil {
				t.Fatalf("create hash error: %v", err)
			}

			if !isVerify {
				t.Fatal("repo::Create password hash not verify")
			}

			if user.ID != wantUser.ID {
				t.Errorf("create user ID %d want %d", user.ID, wantUser.ID)
			}
			if user.Name != wantUser.Name {
				t.Errorf("create user name %q; want %q", user.Name, wantUser.Name)
			}
			if user.Age != wantUser.Age {
				t.Errorf("create user age %d; want %d", user.Age, wantUser.Age)
			}
		})
	}
}

var ErrFakeRepoError = errors.New("some repo error")
var ErrTokenGenerateError = errors.New("some token generate error")

func TestLoginService(t *testing.T) {
	var tests = []struct {
		name            string
		email           string
		password        string
		isRepoCalled    bool
		hashMatch       bool
		RepoErr         error
		wantAccessToken string
		generaterErr    error
		userID          int
		isGenerateCall  bool
		wantErr         error
	}{
		{
			name:         "empty email",
			email:        "",
			password:     "123",
			isRepoCalled: false,
			hashMatch:    true,
			wantErr:      ErrInvalidCredentials,
		},
		{
			name:         "empty password",
			email:        "email",
			password:     "",
			isRepoCalled: false,
			hashMatch:    true,
			wantErr:      ErrInvalidCredentials,
		},
		{
			name:         "email not found",
			email:        "unknow_email",
			password:     "123",
			isRepoCalled: true,
			hashMatch:    true,
			RepoErr:      ErrNotFound,
			wantErr:      ErrInvalidCredentials,
		},
		{
			name:         "password hash not match",
			email:        "email@test.com",
			password:     "123456",
			isRepoCalled: true,
			hashMatch:    false,
			wantErr:      ErrInvalidCredentials,
		},
		{
			name:         "repository error",
			email:        "email@test.com",
			password:     "123456",
			isRepoCalled: true,
			hashMatch:    true,
			RepoErr:      ErrFakeRepoError,
			wantErr:      ErrFakeRepoError,
		},
		{
			name:            "token generation error",
			email:           "email@test.com",
			password:        "123456",
			isRepoCalled:    true,
			hashMatch:       true,
			wantAccessToken: "access_token",
			isGenerateCall:  true,
			userID:          1,
			generaterErr:    ErrTokenGenerateError,
			wantErr:         ErrTokenGenerateError,
		},
		{
			name:            "success",
			email:           "email@test.com",
			password:        "123456",
			isRepoCalled:    true,
			hashMatch:       true,
			wantAccessToken: "access_token",
			isGenerateCall:  true,
			userID:          1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passwordHash, err := password.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error: %v", err)
			}

			pass := tt.password
			if !tt.hashMatch {
				pass = pass + "1"
			}
			repo := FakeUserRepository{
				Credentials: Credentials{
					UserID:       tt.userID,
					PasswordHash: passwordHash,
				},
				AuthErr: tt.RepoErr,
			}
			tGenerator := &FakeTokenGenerator{
				accessToken: tt.wantAccessToken,
				userID:      tt.userID,
				generateErr: tt.generaterErr,
			}
			service := NewService(&repo, tGenerator)

			ctx := context.Background()
			token, err := service.Login(ctx, tt.email, pass)

			if repo.AuthCalled != tt.isRepoCalled {
				t.Errorf("repo method ByEmail is called %t; want %t", repo.AuthCalled, tt.isRepoCalled)
			}
			if tGenerator.isCallGenerator != tt.isGenerateCall {
				t.Errorf("token generate is called %t; want %t", tGenerator.isCallGenerator, tt.isGenerateCall)
			}

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error %v;expected %v", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if token != tt.wantAccessToken {
				t.Errorf("got access token %q; want %q", token, tt.wantAccessToken)
			}
		})
	}
}
