package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mltheuser/ai-router/api"
)

// handleListModels handles GET /v1/models.
// Query params: ?type=cloud|local, ?capability=embed|chat, ?search=<text>
func (s *Server) handleListModels(w http.ResponseWriter, r *http.Request) {
	var provType *api.ProviderType
	if t := r.URL.Query().Get("type"); t != "" {
		pt := api.ProviderType(t)
		if pt != api.ProviderTypeCloud && pt != api.ProviderTypeLocal {
			api.WriteError(w, http.StatusBadRequest, "type must be 'cloud' or 'local'")
			return
		}
		provType = &pt
	}

	var capability *api.Capability
	if c := r.URL.Query().Get("capability"); c != "" {
		cp := api.Capability(c)
		if cp != api.CapabilityChat && cp != api.CapabilityEmbed {
			api.WriteError(w, http.StatusBadRequest, "capability must be 'chat' or 'embed'")
			return
		}
		capability = &cp
	}

	search := r.URL.Query().Get("search")

	models := s.catalog.AllModels(provType, capability)

	// Apply search filter
	if search != "" {
		search = strings.ToLower(search)
		var filtered []api.ModelInfo
		for _, m := range models {
			if strings.Contains(strings.ToLower(m.ID), search) {
				filtered = append(filtered, m)
			}
		}
		models = filtered
	}

	resp := api.ModelList{
		Object: "list",
		Data:   models,
	}
	if resp.Data == nil {
		resp.Data = []api.ModelInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// handleRefreshModels handles POST /v1/models/refresh.
func (s *Server) handleRefreshModels(w http.ResponseWriter, r *http.Request) {
	s.catalog.RefreshAll(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"refreshed"}`))
}
