package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/usecase"
)

// testReport is the response of POST /v1/test: the provider's verification,
// then a report per use case the provider serves.
type testReport struct {
	Provider string                    `json:"provider"`
	Verify   usecase.Check             `json:"verify"`
	UseCases map[string]usecase.Report `json:"use_cases"`
}

// handleTest serves POST /v1/test: it verifies one provider end to end, by
// running every use case's scenarios against it through this server.
func (s *Server) handleTest(w http.ResponseWriter, r *http.Request) error {
	var req usecase.TestRequest
	if err := api.DecodeJSON(r, &req); err != nil {
		return err
	}
	if req.Provider == "" {
		return api.NewError(http.StatusBadRequest, "provider is required")
	}
	req.URL = "http://" + s.httpServer.Addr

	var p provider.Provider
	for _, candidate := range s.cfg.Providers {
		if candidate.Name() == req.Provider {
			p = candidate
		}
	}
	if p == nil {
		return api.NewError(http.StatusNotFound, fmt.Sprintf("provider '%s' is not configured or failed verification at startup", req.Provider))
	}

	report := testReport{Provider: req.Provider, UseCases: map[string]usecase.Report{}}
	verifyCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if err := p.Verify(verifyCtx); err != nil {
		report.Verify = usecase.Check{Status: usecase.Fail, Error: err.Error()}
		api.WriteJSON(w, report)
		return nil
	}
	report.Verify = usecase.Check{Status: usecase.Pass}

	for _, uc := range s.cfg.UseCases {
		if req.UseCase != "" && req.UseCase != uc.Name() {
			continue
		}
		if ucReport, ok := uc.Test(r.Context(), req); ok {
			report.UseCases[uc.Name()] = ucReport
		}
	}
	api.WriteJSON(w, report)
	return nil
}
