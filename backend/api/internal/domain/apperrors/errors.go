// Package apperrors holds cross-cutting application errors shared across layers.
// Driving adapters may map these to HTTP status codes without importing persistence.
package apperrors

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrNotImplemented = errors.New("not implemented")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidToken   = errors.New("invalid token")
	ErrConflict       = errors.New("conflict")
	ErrValidation     = errors.New("validation failed")
)
