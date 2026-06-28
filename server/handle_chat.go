package server

import (
	"encoding/json"
	"net/http"

	"github.com/mltheuser/ai-router/api"
)

// handleChatCompletions handles POST /v1/chat/completions requests.
func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req api.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteBadRequest(w, "invalid request body")
		return
	}

	if !validateChatRequest(w, &req) {
		return
	}

	// Resolve the provider based on the model string
	// Chat completions is the "chat" capability
	resolution, err := s.router.Resolve(req.Model, api.CapabilityChat)
	if err != nil {
		api.WriteError(w, err)
		return
	}

	// Replace model with the cleaned ID (without :tag)
	req.Model = resolution.ModelID

	// Forward the request to the resolved provider
	resp, err := resolution.Provider.Chat(r.Context(), &req)
	if err != nil {
		api.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Too late to write header error
		return
	}
}

// validateChatRequest checks inbound enum fields and writes a 400 JSON error if
// any are invalid. It returns false when a response has already been written.
func validateChatRequest(w http.ResponseWriter, req *api.ChatRequest) bool {
	if len(req.Messages) == 0 {
		api.WriteBadRequest(w, "messages is required")
		return false
	}

	if req.ReasoningEffort != nil {
		switch *req.ReasoningEffort {
		case api.ReasoningEffortNone, api.ReasoningEffortLow, api.ReasoningEffortMedium, api.ReasoningEffortHigh:
		default:
			api.WriteBadRequest(w, "reasoning_effort must be one of: none, low, medium, high")
			return false
		}
	}

	for _, m := range req.Messages {
		switch m.Role {
		case api.RoleSystem, api.RoleUser, api.RoleAssistant, api.RoleTool:
		default:
			api.WriteBadRequest(w, "message role must be one of: system, user, assistant, tool")
			return false
		}
	}

	return true
}
