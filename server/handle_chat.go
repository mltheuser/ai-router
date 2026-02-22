package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mltheuser/ai-router/api"
)

// handleChatCompletions handles POST /v1/chat/completions requests.
func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req api.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Resolve the provider based on the model string
	// Chat completions is the "chat" capability
	resolution, err := s.router.Resolve(req.Model, api.CapabilityChat)
	if err != nil {
		if errors.Is(err, api.ErrModelNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if errors.Is(err, api.ErrNotSupported) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Replace model with the cleaned ID (without :tag)
	req.Model = resolution.ModelID

	// Forward the request to the resolved provider
	resp, err := resolution.Provider.Chat(r.Context(), &req)
	if err != nil {
		// Determine suitable status code based on error type?
		// For now, internal server error is safe.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Too late to write header error
		return
	}
}
