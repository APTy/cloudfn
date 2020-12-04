package fnerrors

import (
	"fmt"
	"net/http"
)

// Common errors.
var (
	MethodNotAllowed   = New(http.StatusMethodNotAllowed, "method not allowed")
	NotFound           = New(http.StatusNotFound, "not found")
	Unauthorized       = New(http.StatusUnauthorized, "unauthorized")
	ServiceUnavailable = New(http.StatusServiceUnavailable, "service unavailable")
	BadRequest         = New(http.StatusBadRequest, "bad request")
)

// Error is an error type.
type Error struct {
	status int
	msg    string
	detail string
}

// Option is used to customize the error.
type Option func(*Error)

// Detail appends detail to the error.
func Detail(detail string) Option {
	return func(e *Error) {
		e.detail = detail
	}
}

// New returns a new error with the provided HTTP status.
func New(status int, msg string, opts ...Option) error {
	e := &Error{
		status: status,
		msg:    msg,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Error) Error() string {
	return e.msg
}

// HTTPStatus returns the HTTP status code for the error.
func (e *Error) HTTPStatus() int {
	return e.status
}

// Detail returns the detail for the error.
func (e *Error) Detail() string {
	return e.detail
}

// Wrap wraps the error.
func Wrap(msg string, err error) error {
	if err, ok := err.(*Error); ok {
		return New(err.status, fmt.Sprintf("%s: %v", msg, err))
	}
	return err
}

// GetDetail returns the detail of the error.
func GetDetail(err error) string {
	if err, ok := err.(*Error); ok {
		return err.Detail()
	}
	return ""
}
