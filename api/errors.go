package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Sentinel errors for provider implementations.
var (
	ErrNotSupported       = fmt.Errorf("capability not supported by this provider")
	ErrModelNotFound      = fmt.Errorf("model not found")
	ErrProviderUnavailable = fmt.Errorf("provider unavailable")
)

// APIError represents an OpenAI-compatible error response.
type APIError struct {
	StatusCode int    `json:"-"`
	Type       string `json:"type"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// ErrorResponse is the top-level error response wrapper.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// WriteError writes an OpenAI-compatible error response to the http.ResponseWriter.
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	errType := "invalid_request_error"
	switch {
	case statusCode == http.StatusNotFound:
		errType = "not_found_error"
	case statusCode == http.StatusInternalServerError:
		errType = "server_error"
	case statusCode == http.StatusServiceUnavailable:
		errType = "service_unavailable"
	}

	resp := ErrorResponse{
		Error: APIError{
			Type:    errType,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
