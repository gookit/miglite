package common

import (
	"fmt"
	"log"
)

// AppError represents an application-specific error
type AppError struct {
	Code    string
	Message string
	Err     error
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewError creates a new AppError
func NewError(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, code, message string) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Log default Logger instance
var Log = &DefaultLogger{}

// DefaultLogger is a simple logger implementation
type DefaultLogger struct{}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, args ...any) {
	log.Printf("[DEBUG] "+msg, args...)
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string, args ...any) {
	log.Printf("[INFO] "+msg, args...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, args ...any) {
	log.Printf("[WARN] "+msg, args...)
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, args ...any) {
	log.Printf("[ERROR] "+msg, args...)
}
