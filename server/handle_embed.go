package server

import (
	"encoding/json"
	"net/http"

	"github.com/mltheuser/ai-router/api"
)

// handleEmbed handles POST /v1/embeddings requests.
func (s *Server) handleEmbed(w http.ResponseWriter, r *http.Request) {
	var req api.EmbedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if typeErr, ok := err.(*json.UnmarshalTypeError); ok && typeErr.Field == "input" {
			api.WriteBadRequest(w, "Invalid request body: 'input' must be an array of strings (batch mode)")
			return
		}
		api.WriteBadRequest(w, "Invalid request body: "+err.Error())
		return
	}

	if len(req.Input) == 0 {
		api.WriteBadRequest(w, "input is required")
		return
	}

	// Resolve the model to a provider
	result, err := s.router.Resolve(req.Model, api.CapabilityEmbed)
	if err != nil {
		api.WriteError(w, err)
		return
	}

	// Replace model with the cleaned ID (without :tag)
	req.Model = result.ModelID

	// Call the provider
	resp, err := result.Provider.Embed(r.Context(), &req)
	if err != nil {
		s.logger.Error("Embed failed", "provider", result.Provider.Name(), "model", result.ModelID, "error", err)
		api.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
