package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Sentinel errors for provider implementations.
var (
	ErrNotSupported        = fmt.Errorf("capability not supported by this provider")
	ErrModelNotFound       = fmt.Errorf("model not found")
	ErrProviderUnavailable = fmt.Errorf("provider unavailable")
)

// Error represents an OpenAI-compatible error response.
type Error struct {
	StatusCode int    `json:"-"`
	Type       string `json:"type"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// ErrorResponse is the top-level error response wrapper.
type ErrorResponse struct {
	Error Error `json:"error"`
}

// WriteError writes an OpenAI-compatible error response to the http.ResponseWriter.
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	var errType string
	switch statusCode {
	case http.StatusNotFound:
		errType = "not_found_error"
	case http.StatusInternalServerError:
		errType = "server_error"
	case http.StatusServiceUnavailable:
		errType = "service_unavailable"
	default:
		errType = "invalid_request_error"
	}

	resp := ErrorResponse{
		Error: Error{
			Type:    errType,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}
