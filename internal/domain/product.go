package domain

import "time"

type Product struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateProductRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type UpdateProductRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type ProductResponse struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	UserID    string    `json:"userId" db:"user_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
