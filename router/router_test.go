package router

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/provider"
)

// StubProvider implements provider.Provider for testing.
type StubProvider struct {
	name   string
	pType  api.ProviderType
	models []api.ModelInfo
}

func (p *StubProvider) Name() string {
	return p.name
}

func (p *StubProvider) Type() api.ProviderType {
	return p.pType
}

func (p *StubProvider) Verify(_ context.Context) error {
	return nil
}

func (p *StubProvider) ListModels(_ context.Context) ([]api.ModelInfo, error) {
	return p.models, nil
}

func (p *StubProvider) Embed(_ context.Context, _ *api.EmbedRequest) (*api.EmbedResponse, error) {
	return nil, fmt.Errorf("not supported")
}

func (p *StubProvider) Chat(_ context.Context, _ *api.ChatRequest) (*api.ChatResponse, error) {
	return nil, fmt.Errorf("not supported")
}

func TestSelectBestCandidate(t *testing.T) {
	c1 := 0.01
	c2 := 0.02

	cloudModels := []api.ModelInfo{
		{ID: "m1", Provider: "p1", ProviderType: api.ProviderTypeCloud, CostPerMInput: &c2},
		{ID: "m1", Provider: "p2", ProviderType: api.ProviderTypeCloud, CostPerMInput: &c1}, // Cheapest
	}

	best, err := SelectBestCandidate(cloudModels, api.ProviderTypeCloud)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if best.Provider != "p2" {
		t.Errorf("expected p2 (cheapest), got %s", best.Provider)
	}

	s1 := int64(100)
	s2 := int64(200)

	localModels := []api.ModelInfo{
		{ID: "m2", Provider: "p3", ProviderType: api.ProviderTypeLocal, SizeBytes: &s2},
		{ID: "m2", Provider: "p4", ProviderType: api.ProviderTypeLocal, SizeBytes: &s1}, // Smallest
	}

	bestLocal, err := SelectBestCandidate(localModels, api.ProviderTypeLocal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bestLocal.Provider != "p4" {
		t.Errorf("expected p4 (smallest), got %s", bestLocal.Provider)
	}

	// Test error case
	_, err = SelectBestCandidate(localModels, "invalid_tag")
	if err == nil {
		t.Error("expected error for invalid tag, got nil")
	}
}

func TestResolve(t *testing.T) {
	// Setup catalog with stubs
	cloudP := &StubProvider{
		name:  "openrouter",
		pType: api.ProviderTypeCloud,
		models: []api.ModelInfo{
			{ID: "gpt-4", Provider: "openrouter", ProviderType: api.ProviderTypeCloud, Capabilities: []api.Capability{api.CapabilityChat}, CostPerMInput: new(float64)},
		},
	}
	// Give it a cost so it can be selected
	*cloudP.models[0].CostPerMInput = 1.0

	localP := &StubProvider{
		name:  "ollama",
		pType: api.ProviderTypeLocal,
		models: []api.ModelInfo{
			{ID: "llama2", Provider: "ollama", ProviderType: api.ProviderTypeLocal, Capabilities: []api.Capability{api.CapabilityChat}, SizeBytes: new(int64)},
		},
	}
	*localP.models[0].SizeBytes = 1000

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	catalog := NewModelCatalog(logger, []provider.Provider{cloudP, localP})

	// Manually populate catalog to avoid async Initialize in unit test if possible,
	// but Initialize is robust so let's use it.
	if err := catalog.Initialize(context.Background()); err != nil {
		t.Fatalf("failed to initialize catalog: %v", err)
	}

	r := NewRouter(catalog)

	tests := []struct {
		name      string
		modelStr  string
		wantProv  string
		wantModel string
		expectErr bool
	}{
		{
			name:      "Resolve Cloud",
			modelStr:  "gpt-4:cloud",
			wantProv:  "openrouter",
			wantModel: "gpt-4",
			expectErr: false,
		},
		{
			name:      "Resolve Local",
			modelStr:  "llama2:local",
			wantProv:  "ollama",
			wantModel: "llama2",
			expectErr: false,
		},
		{
			name:      "Resolve Pinned",
			modelStr:  "gpt-4:cloud@openrouter",
			wantProv:  "openrouter",
			wantModel: "gpt-4",
			expectErr: false,
		},
		{
			name:      "Resolve Missing Tag",
			modelStr:  "gpt-4",
			expectErr: true,
		},
		{
			name:      "Resolve Wrong Tag",
			modelStr:  "gpt-4:local", // It's cloud only in our stub
			expectErr: true,
		},
		{
			name:      "Resolve Wrong Provider",
			modelStr:  "gpt-4:cloud@ollama",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := r.Resolve(tt.modelStr, api.CapabilityChat)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.Provider.Name() != tt.wantProv {
				t.Errorf("expected provider %s, got %s", tt.wantProv, res.Provider.Name())
			}
			if res.ModelID != tt.wantModel {
				t.Errorf("expected model %s, got %s", tt.wantModel, res.ModelID)
			}
		})
	}
}
