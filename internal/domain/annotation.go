package domain

import "time"

type Annotation struct {
	ID          string    `json:"id" db:"id"`
	VersionID   string    `json:"version_id" db:"version_id"`
	Type        string    `json:"type" db:"type"`
	Data        string    `json:"data" db:"data"`
	UserID      string    `json:"user_id" db:"user_id"`
	SessionID   string    `json:"session_id" db:"session_id"`
	Color       string    `json:"color" db:"color"`
	StrokeW     float64   `json:"stroke_w" db:"stroke_w"`
	StrokeStyle string    `json:"stroke_style" db:"stroke_style"`
	X           float64   `json:"x" db:"x"`
	Y           float64   `json:"y" db:"y"`
	Title       string    `json:"title" db:"title"`
	Text        string    `json:"text" db:"text"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateAnnotationRequest struct {
	Type        string  `json:"type" validate:"required"`
	Data        string  `json:"data"`
	SessionID   string  `json:"sessionId"`
	Color       string  `json:"color"`
	StrokeW     float64 `json:"strokeW"`
	StrokeStyle string  `json:"strokeStyle"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Title       string  `json:"title"`
	Text        string  `json:"text"`
}

type UpdateAnnotationRequest struct {
	Data        string  `json:"data"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Title       string  `json:"title"`
	Text        string  `json:"text"`
	Color       string  `json:"color"`
}

type AnnotationResponse struct {
	ID          string    `json:"id"`
	VersionID   string    `json:"versionId"`
	Type        string    `json:"type"`
	Data        string    `json:"data"`
	UserID      string    `json:"userId"`
	SessionID   string    `json:"sessionId"`
	Color       string    `json:"color"`
	StrokeW     float64   `json:"strokeW"`
	StrokeStyle string    `json:"strokeStyle"`
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	Title       string    `json:"title"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
