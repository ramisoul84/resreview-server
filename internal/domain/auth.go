package domain

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=5,max=100"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refreshToken"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Color        string `json:"color"`
	IsAdmin      bool   `json:"is_admin"`
}

type UpdateRoleRequest struct {
	IsAdmin bool `json:"isAdmin"`
}
