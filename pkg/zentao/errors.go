package zentao

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorKind classifies ZenTao API errors so callers can react structurally
// (exit codes, retry decisions) instead of parsing error strings.
type ErrorKind string

const (
	// KindAuth means the session expired or the request was not authenticated.
	KindAuth ErrorKind = "auth"
	// KindPermission means the request was authenticated but not allowed.
	KindPermission ErrorKind = "permission"
	// KindAPI covers every other ZenTao-side failure (form validation,
	// server errors, malformed envelopes).
	KindAPI ErrorKind = "api"
)

// ErrValidation marks client-side validation failures such as missing
// required arguments. Wrap it: fmt.Errorf("%w: task create: --id is required", ErrValidation).
var ErrValidation = errors.New("invalid arguments")

// Error is a structured ZenTao API error. HTTPStatus is set (non-zero) when
// the server responded with an HTTP error code.
type Error struct {
	Kind       ErrorKind
	HTTPStatus int
	Message    string
}

// Error implements the error interface. Message format is stable and matches
// the pre-structured-error output.
func (e *Error) Error() string {
	if e.HTTPStatus > 0 {
		return fmt.Sprintf("http %d: %s", e.HTTPStatus, e.Message)
	}
	return fmt.Sprintf("zentao: %s", e.Message)
}

// authError builds a KindAuth error.
func authError(msg string) *Error {
	return &Error{Kind: KindAuth, Message: msg}
}

// permissionError builds a KindPermission error.
func permissionError(msg string) *Error {
	return &Error{Kind: KindPermission, Message: msg}
}

// apiError builds a KindAPI error.
func apiError(msg string) *Error {
	return &Error{Kind: KindAPI, Message: msg}
}

// IsAuthError reports whether err is an authentication failure (session
// expired, login required, 401) that a re-login may fix.
func IsAuthError(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		if e.Kind == KindAuth {
			return true
		}
		return e.HTTPStatus == http.StatusUnauthorized
	}
	return false
}

// IsPermissionError reports whether err is an authenticated-but-forbidden
// failure (or 403) that re-login cannot fix.
func IsPermissionError(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		if e.Kind == KindPermission {
			return true
		}
		return e.HTTPStatus == http.StatusForbidden
	}
	return false
}
