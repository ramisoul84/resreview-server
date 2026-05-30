package domain

import "time"

type Session struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateSessionRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type UpdateSessionRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type SessionResponse struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	UserID    string    `json:"userId" db:"user_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
