package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Error is both the JSON error wire shape and a classifiable domain error: it
// carries the HTTP status and error type alongside the message, so an error's
// full mapping lives in one place.
type Error struct {
	StatusCode int    `json:"-"`
	Type       string `json:"type"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

func (e *Error) Error() string { return e.Message }

// Sentinel domain errors. Each carries its own HTTP status and error type.
var (
	ErrNotSupported        = NewError(http.StatusBadRequest, "capability not supported by this provider")
	ErrModelNotFound       = NewError(http.StatusNotFound, "model not found")
	ErrProviderUnavailable = NewError(http.StatusServiceUnavailable, "provider unavailable")
	ErrInvalidModel        = NewError(http.StatusBadRequest, "invalid model string")
)

// ErrorResponse is the top-level error response wrapper.
type ErrorResponse struct {
	Error Error `json:"error"`
}

// NewError builds an *Error, deriving the error type string from the status code.
func NewError(statusCode int, message string) *Error {
	return &Error{StatusCode: statusCode, Type: typeForStatus(statusCode), Message: message}
}

func typeForStatus(statusCode int) string {
	switch statusCode {
	case http.StatusNotFound:
		return "not_found_error"
	case http.StatusInternalServerError:
		return "server_error"
	case http.StatusServiceUnavailable:
		return "service_unavailable"
	default:
		return "invalid_request_error"
	}
}

// WriteError writes err as a JSON error response. If err is, or wraps, an
// *Error, that error's status and type are used; otherwise err is treated as an
// internal server error. The full (possibly wrapped) message is preserved.
func WriteError(w http.ResponseWriter, err error) {
	out := Error{StatusCode: http.StatusInternalServerError, Type: "server_error", Message: err.Error()}
	var domain *Error
	if errors.As(err, &domain) {
		out.StatusCode = domain.StatusCode
		out.Type = domain.Type
		out.Message = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(out.StatusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: out})
}

// WriteBadRequest writes an error response with StatusBadRequest.
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, NewError(http.StatusBadRequest, message))
}
