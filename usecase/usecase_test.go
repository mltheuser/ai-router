package usecase

import (
	"context"
	"testing"

	"github.com/mltheuser/ai-router/provider"
)

// stubProvider is a provider serving a fixed model list.
type stubProvider struct {
	name   string
	typ    provider.Type
	models []stubModel
}

func (p *stubProvider) Name() string                   { return p.name }
func (p *stubProvider) Type() provider.Type            { return p.typ }
func (p *stubProvider) Verify(_ context.Context) error { return nil }

func (p *stubProvider) listStubModels(_ context.Context) ([]stubModel, error) {
	return p.models, nil
}

type stubModel struct {
	provider.ModelRef
	cost *float64
}

func stubBase(providers ...*stubProvider) *Base[stubModel, *stubProvider] {
	ps := make([]provider.Provider, len(providers))
	for i, p := range providers {
		ps[i] = p
	}
	return NewBase(context.Background(), Spec[stubModel, *stubProvider]{
		Name:   "stub",
		List:   (*stubProvider).listStubModels,
		Prefer: func(a, b stubModel) bool { return LessKnown(a.cost, b.cost) },
	}, ps)
}

func withModels(p *stubProvider, costs map[string]*float64) *stubProvider {
	for id, cost := range costs {
		p.models = append(p.models, stubModel{ModelRef: provider.NewModelRef(p, id), cost: cost})
	}
	return p
}

func price(v float64) *float64 { return &v }

func TestResolve(t *testing.T) {
	b := stubBase(
		withModels(&stubProvider{name: "cheap", typ: provider.Cloud}, map[string]*float64{"gpt": price(1)}),
		withModels(&stubProvider{name: "pricey", typ: provider.Cloud}, map[string]*float64{"gpt": price(2)}),
		withModels(&stubProvider{name: "unpriced", typ: provider.Cloud}, map[string]*float64{"gpt": nil}),
		withModels(&stubProvider{name: "ollama", typ: provider.Local}, map[string]*float64{"llama": price(0)}),
	)

	tests := []struct {
		model        string
		wantProvider string // empty means an error is expected
	}{
		{"gpt:cloud", "cheap"},             // the preferred provider wins
		{"gpt:cloud@pricey", "pricey"},     // a pin overrides the preference
		{"gpt:cloud@unpriced", "unpriced"}, // a pin reaches an unpriced entry
		{"llama:local", "ollama"},          // local models resolve by tag
		{"llama:local@ollama", "ollama"},   // and by pin
		{"gpt", ""},                        // the tag is required
		{"gpt:remote", ""},                 // the tag must be cloud or local
		{":cloud", ""},                     // the ID is required
		{"gpt:local", ""},                  // the tag must match the provider type
		{"gpt:cloud@ollama", ""},           // the pinned provider must list the model
		{"unknown:cloud", ""},              // the model must be listed
		{"llama:local@nonexistent", ""},    // the pinned provider must exist
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			m, p, err := b.Resolve(tt.model)
			if tt.wantProvider == "" {
				if err == nil {
					t.Fatalf("expected an error, resolved to %s", m.Model)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Name() != tt.wantProvider || m.Provider != tt.wantProvider {
				t.Errorf("resolved to provider %s (model %s), want %s", p.Name(), m.Model, tt.wantProvider)
			}
		})
	}
}

func TestNewModelRefQualifiesTheID(t *testing.T) {
	ref := provider.NewModelRef(&stubProvider{name: "ollama", typ: provider.Local}, "qwen3:8b")
	if got, want := ref.Model, "qwen3:8b:local@ollama"; got != want {
		t.Errorf("model string %q, want %q", got, want)
	}
}

func TestLessKnown(t *testing.T) {
	tests := []struct {
		a, b *float64
		want bool
	}{
		{price(1), price(2), true},
		{price(2), price(1), false},
		{price(1), price(1), false},
		{price(5), nil, true}, // known ranks before unknown
		{nil, price(5), false},
		{nil, nil, false},
	}
	for _, tt := range tests {
		if got := LessKnown(tt.a, tt.b); got != tt.want {
			t.Errorf("LessKnown(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
