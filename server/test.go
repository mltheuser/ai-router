package server

import (
	"fmt"
	"net/http"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
)

// testReport is the response of POST /v1/test: a report per use case the
// provider serves.
type testReport struct {
	Provider string                   `json:"provider"`
	UseCases map[string]router.Report `json:"use_cases"`
}

// handleTest serves POST /v1/test: it verifies one provider end to end by
// running every use case's scenarios against it through this server, on the
// models the server serves.
func (s *Server) handleTest(w http.ResponseWriter, r *http.Request) error {
	var req router.TestRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}
	if req.Provider == "" {
		return httpx.NewError(http.StatusBadRequest, "provider is required")
	}
	req.URL = "http://" + s.httpServer.Addr

	report := testReport{Provider: req.Provider, UseCases: map[string]router.Report{}}
	knownUseCase := req.UseCase == ""
	for _, uc := range s.cfg.UseCases {
		if req.UseCase != "" && req.UseCase != uc.Name() {
			continue
		}
		knownUseCase = true
		if ucReport, ok := uc.Test(r.Context(), req); ok {
			report.UseCases[uc.Name()] = ucReport
		}
	}

	switch {
	case !knownUseCase:
		return httpx.NewError(http.StatusBadRequest, fmt.Sprintf("unknown use case '%s'", req.UseCase))
	case len(report.UseCases) == 0 && req.UseCase != "":
		return httpx.NewError(http.StatusNotFound, fmt.Sprintf("provider '%s' does not serve %s: it is not loaded or does not implement it", req.Provider, req.UseCase))
	case len(report.UseCases) == 0:
		return httpx.NewError(http.StatusNotFound, fmt.Sprintf("provider '%s' is not loaded: it is not configured or failed verification at startup", req.Provider))
	}
	httpx.WriteJSON(w, report)
	return nil
}
