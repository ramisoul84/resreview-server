package domain

import "time"

type User struct {
	ID           string     `db:"id"`
	Name         string     `db:"name"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Color        string     `db:"color"`
	IsAdmin      bool       `db:"is_admin"`
	CreatedAt    time.Time  `db:"created_at"`
	LastLoginAt  *time.Time `db:"last_login_at"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Color string `json:"color" validate:"required"`
}

type UserResponse struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Email       string     `json:"email" db:"email"`
	Color       string     `json:"color" db:"color"`
	IsAdmin     bool       `json:"isAdmin" db:"is_admin"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	LastLoginAt *time.Time `json:"lastLoginAt" db:"last_login_at"`
}
