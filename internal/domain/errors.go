package domain

import "errors"

type contextKey string

const RequestIDKey contextKey = "request_id"

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrTokenInvalid    = errors.New("token is invalid")
	ErrTokenExpired    = errors.New("token has expired")
	ErrTokenNotFound   = errors.New("refresh token not found")
	ErrTokenRevoked    = errors.New("token has been revoked")
	ErrSessionNotFound = errors.New("session not found")
	ErrProductNotFound = errors.New("product not found")
	ErrVersionNotFound = errors.New("version not found")

	ErrPhotoNotFound     = errors.New("photo not found")
	ErrInvalidMimeType   = errors.New("invalid mime type")
	ErrFileTooLarge      = errors.New("file too large")
	ErrAnnotationNotFound = errors.New("annotation not found")
)
