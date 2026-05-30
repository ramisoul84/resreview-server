package domain

import "time"

type Version struct {
	ID        string    `json:"id" db:"id"`
	Label     string    `json:"label" db:"label"`
	Name      string    `json:"name" db:"name"`
	ProductID string    `json:"product_id" db:"product_id"`
	URL       string    `json:"url" db:"url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateVersionRequest struct {
	Label string `json:"label" validate:"required,min=1,max=50"`
	Name  string `json:"name" validate:"required,min=1,max=100"`
}

type UpdateVersionRequest struct {
	Label string `json:"label" validate:"required,min=1,max=50"`
	Name  string `json:"name" validate:"required,min=1,max=100"`
}

type VersionResponse struct {
	ID        string    `json:"id" db:"id"`
	Label     string    `json:"label" db:"label"`
	Name      string    `json:"name" db:"name"`
	ProductID string    `json:"productId" db:"product_id"`
	URL       string    `json:"url" db:"url"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
