package user

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type CreateUserParams struct {
	ID           int
	Name         string
	Age          int
	Email        string
	PasswordHash string
}

type RegisterInput struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type Credentials struct {
	UserID       int `json:"id"`
	PasswordHash string
}
