package apperror

import "fmt"

// AppError is a typed, HTTP-aware application error.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

// New creates a custom AppError.
func New(code string, httpStatus int, msg string) *AppError {
	return &AppError{Code: code, HTTPStatus: httpStatus, Message: msg}
}

// Newf creates a custom AppError with format string.
func Newf(code string, httpStatus int, format string, args ...any) *AppError {
	return &AppError{Code: code, HTTPStatus: httpStatus, Message: fmt.Sprintf(format, args...)}
}

// Sentinel errors.
var (
	ErrNotFound      = &AppError{Code: "NOT_FOUND", HTTPStatus: 404, Message: "resource not found"}
	ErrUnauthorized  = &AppError{Code: "UNAUTHORIZED", HTTPStatus: 401, Message: "authentication required"}
	ErrForbidden     = &AppError{Code: "FORBIDDEN", HTTPStatus: 403, Message: "access denied"}
	ErrConflict      = &AppError{Code: "CONFLICT", HTTPStatus: 409, Message: "resource already exists"}
	ErrInternal      = &AppError{Code: "INTERNAL", HTTPStatus: 500, Message: "internal server error"}
	ErrBadTransition = &AppError{Code: "BAD_TRANSITION", HTTPStatus: 422, Message: "status transition not allowed"}
)

// Validation builds a validation error with a field-specific message.
func Validation(msg string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", HTTPStatus: 422, Message: msg}
}

// NotFound builds a not-found error for a specific resource.
func NotFound(resource string) *AppError {
	return &AppError{Code: "NOT_FOUND", HTTPStatus: 404, Message: resource + " not found"}
}

// Conflict builds a conflict error.
func Conflict(msg string) *AppError {
	return &AppError{Code: "CONFLICT", HTTPStatus: 409, Message: msg}
}
