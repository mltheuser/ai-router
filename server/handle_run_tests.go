package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/scenarios"
)

func (s *Server) handleTest(w http.ResponseWriter, r *http.Request) {
	var req api.TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	report := api.TestReport{
		Provider: req.Provider,
	}

	// 1. Resolve Provider
	p := s.catalog.GetProvider(req.Provider)
	if p == nil {
		report.Checks.Verify = api.Check{Status: api.StatusFail, Error: fmt.Sprintf("Provider '%s' not found", req.Provider)}
		writeJSON(w, report)
		return
	}

	// 2. Common Checks (Always run)
	ctx := r.Context()
	verifyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := p.Verify(verifyCtx); err != nil {
		report.Checks.Verify = api.Check{Status: api.StatusFail, Error: fmt.Sprintf("Verify failed: %v", err)}
		writeJSON(w, report)
		return
	}
	report.Checks.Verify = api.Check{Status: api.StatusPass}

	models, err := p.ListModels(ctx)
	if err != nil {
		report.Checks.ListModels = api.Check{Status: api.StatusFail, Error: fmt.Sprintf("ListModels failed: %v", err)}
		writeJSON(w, report)
		return
	}

	if len(models) == 0 {
		report.Checks.ListModels = api.Check{Status: api.StatusFail, Error: "ListModels returned 0 models"}
		writeJSON(w, report)
		return
	}
	report.Checks.ListModels = api.Check{Status: api.StatusPass}

	// 3. Run Scenarios
	// Use localhost address for E2E tests
	baseURL := fmt.Sprintf("http://%s", s.httpServer.Addr)

	for _, sc := range scenarios.List() {
		// Find target model
		var targetModel api.ModelInfo
		var found bool

		// Filter capabilities
		required := sc.RequiredCapabilities()

		if req.Model != "" {
			// specific model requested
			for _, m := range models {
				if m.ID == req.Model {
					targetModel = m
					found = true
					break
				}
			}
			if !found {
				report.Scenarios = append(report.Scenarios, api.ScenarioResult{
					Name:   sc.Name(),
					Checks: []api.Check{{Status: api.StatusSkipped, Error: fmt.Sprintf("Model '%s' not found in provider list", req.Model)}},
				})
				continue
			}
			// Check capabilities
			for _, cap := range required {
				if !targetModel.HasCapability(cap) {
					found = false
					report.Scenarios = append(report.Scenarios, api.ScenarioResult{
						Name:   sc.Name(),
						Checks: []api.Check{{Status: api.StatusSkipped, Error: fmt.Sprintf("Model '%s' missing capability '%s'", req.Model, cap)}},
					})
					break
				}
			}
			if !found {
				continue
			}

		} else {
			// Auto-select best model
			var candidates []api.ModelInfo
			for _, m := range models {
				hasAll := true
				for _, cap := range required {
					if !m.HasCapability(cap) {
						hasAll = false
						break
					}
				}
				if hasAll {
					candidates = append(candidates, m)
				}
			}

			if len(candidates) == 0 {
				report.Scenarios = append(report.Scenarios, api.ScenarioResult{
					Name:   sc.Name(),
					Checks: []api.Check{{Status: api.StatusSkipped, Error: "No models found with required capabilities"}},
				})
				continue
			}

			var err error
			targetModel, err = router.SelectBestCandidate(candidates, candidates[0].ProviderType)
			if err != nil {
				report.Scenarios = append(report.Scenarios, api.ScenarioResult{
					Name:   sc.Name(),
					Checks: []api.Check{{Status: api.StatusFail, Error: fmt.Sprintf("Failed to select best candidate: %v", err)}},
				})
				continue
			}
		}

		// Execute Scenario
		testCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Construct routable ID (model:tag@provider)
		routableID := fmt.Sprintf("%s:%s@%s", targetModel.ID, targetModel.ProviderType, p.Name())

		scenarioResult := sc.Run(testCtx, baseURL, routableID)
		scenarioResult.Name = sc.Name()
		scenarioResult.Model = targetModel.ID

		report.Scenarios = append(report.Scenarios, *scenarioResult)
	}

	writeJSON(w, report)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
