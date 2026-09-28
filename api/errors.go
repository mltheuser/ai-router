package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Error is an error with the HTTP status it is served with. Handlers return
// it, or wrap it, to control the response; any other error is served as 500.
type Error struct {
	StatusCode int    `json:"-"`
	Type       string `json:"type"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

func (e *Error) Error() string { return e.Message }

// ErrorResponse is the top-level error response wrapper.
type ErrorResponse struct {
	Error Error `json:"error"`
}

// NewError builds an *Error, deriving the error type string from the status code.
func NewError(statusCode int, message string) *Error {
	return &Error{StatusCode: statusCode, Type: typeForStatus(statusCode), Message: message}
}

// NewUpstreamError converts a non-2xx response from a provider's backend API
// into an *Error. Client errors (4xx) pass through: the backend rejected the
// request itself, so retrying it unchanged will not help. 401/403 are the
// exception: they mean the router's own API key was refused, which the client
// cannot fix. Those and every other status (5xx) mean the provider failed and
// become 502 Bad Gateway. The upstream status and body stay in the message.
func NewUpstreamError(upstreamStatus int, body string) *Error {
	status := http.StatusBadGateway
	if upstreamStatus >= 400 && upstreamStatus < 500 &&
		upstreamStatus != http.StatusUnauthorized && upstreamStatus != http.StatusForbidden {
		status = upstreamStatus
	}
	return NewError(status, fmt.Sprintf("API error (status %d): %s", upstreamStatus, body))
}

func typeForStatus(statusCode int) string {
	switch statusCode {
	case http.StatusNotFound:
		return "not_found_error"
	case http.StatusTooManyRequests:
		return "rate_limit_error"
	case http.StatusInternalServerError:
		return "server_error"
	case http.StatusBadGateway:
		return "upstream_error"
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
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(out.StatusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: out})
}
