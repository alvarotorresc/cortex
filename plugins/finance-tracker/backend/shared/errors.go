package shared

import "fmt"

// AppError represents a typed application error with HTTP status code.
// All domain errors should use this type instead of raw strings, enabling
// consistent error responses across the plugin.
type AppError struct {
	Code       string
	Message    string
	StatusCode int
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates an AppError with explicit code, message, and HTTP status.
func NewAppError(code string, message string, statusCode int) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: statusCode}
}

// NewValidationError creates a 400 validation error.
func NewValidationError(message string) *AppError {
	return NewAppError("VALIDATION_ERROR", message, 400)
}

// NewNotFoundError creates a 404 not-found error for a specific resource.
func NewNotFoundError(resource string, id string) *AppError {
	return NewAppError("NOT_FOUND", fmt.Sprintf("%s %s not found", resource, id), 404)
}

// NewConflictError creates a 409 conflict error.
func NewConflictError(message string) *AppError {
	return NewAppError("CONFLICT", message, 409)
}
