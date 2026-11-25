// Package gitlab is the provider for GitLab's AI Gateway proxy, which serves
// vendor models (Anthropic, for now) through a GitLab personal access token.
//
// The gateway forwards request and response bodies unchanged, so the chat
// mapping is a copy of the anthropic provider's. What differs is around it:
// authentication goes through a short-lived token (see token.go), responses
// are streamed from the gateway to outlast a timeout in front of it (see
// stream.go), and the proxy has no models endpoint, so the served models are
// hard-coded (see models.go).
package gitlab

import (
	"context"
	"fmt"
	"sync"

	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/chat"
)

// Provider talks to the GitLab AI Gateway proxy.
type Provider struct {
	pat string

	// mu guards session, the cached gateway token and the client that sends it.
	mu      sync.Mutex
	session *session
}

// The use cases this provider serves.
var (
	_ chat.Provider = (*Provider)(nil)
)

// New creates a new GitLab provider with the given personal access token. It
// needs the api or ai_features scope.
func New(pat string) *Provider {
	return &Provider{pat: pat}
}

func (p *Provider) Name() string {
	return "gitlab"
}

func (p *Provider) Type() router.ProviderType {
	return router.Cloud
}

// Verify checks the whole path: minting a gateway token proves the personal
// access token and its Duo access, and counting tokens proves the token, its
// headers and the model reach Anthropic. Counting generates nothing and is
// not billed.
func (p *Provider) Verify(ctx context.Context) error {
	client, err := p.client(ctx)
	if err != nil {
		return fmt.Errorf("gitlab verification failed: %w", err)
	}
	req := countTokensRequest{
		Model:    models[0].id,
		Messages: []anthropicMessage{{Role: "user", Content: []interface{}{anthropicTextBlock{Type: "text", Text: "ping"}}}},
	}
	if err := client.Post(ctx, "/messages/count_tokens?beta=true", req, nil); err != nil {
		return fmt.Errorf("gitlab verification failed: %w", err)
	}
	return nil
}

// countTokensRequest is the body of POST /messages/count_tokens.
type countTokensRequest struct {
	Model    string             `json:"model"`
	Messages []anthropicMessage `json:"messages"`
}
