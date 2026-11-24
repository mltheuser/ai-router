package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mltheuser/ai-router/httpx"
)

// The gateway does not accept the personal access token itself. GitLab mints
// a gateway token from it, valid for an hour, together with headers the
// gateway requires on every call. Minting takes about a second and writes an
// audit event, so the token is cached here and reused until shortly before it
// expires. This file is the only state in the provider.

const (
	gitlabURL = "https://gitlab.com/api/v4"
	proxyURL  = "https://cloud.gitlab.com/ai/v1/proxy/anthropic/v1"

	// anthropicVersion is the API version sent in the anthropic-version
	// header, which the proxy requires and forwards to Anthropic.
	anthropicVersion = "2023-06-01"

	// userAgent is sent explicitly: Cloudflare in front of the gateway blocks
	// some default user agents.
	userAgent = "ai-router"

	// renewBefore is how long before expiry a token is replaced. The gateway
	// checks the token when a request arrives, so this only has to cover
	// clock skew and the time a request takes to get there.
	renewBefore = 5 * time.Minute
)

// session is one gateway token and the client that sends it.
type session struct {
	client    *httpx.Client
	expiresAt time.Time
}

// directAccessResponse is the response from POST
// /ai/third_party_agents/direct_access.
type directAccessResponse struct {
	Token     string            `json:"token"`
	ExpiresAt int64             `json:"expires_at"`
	Headers   map[string]string `json:"headers"`
}

// client returns the client for the proxy, minting a new gateway token when
// there is none or the current one is about to expire. The lock is held while
// minting, so concurrent requests wait for one token instead of each minting
// their own.
func (p *Provider) client(ctx context.Context) (*httpx.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.session != nil && time.Until(p.session.expiresAt) > renewBefore {
		return p.session.client, nil
	}
	s, err := p.mint(ctx)
	if err != nil {
		return nil, err
	}
	p.session = s
	return s.client, nil
}

// mint requests a new gateway token and builds the client that sends it.
func (p *Provider) mint(ctx context.Context) (*session, error) {
	gitlab := httpx.NewClient(gitlabURL,
		httpx.WithHeader("Authorization", "Bearer "+p.pat),
		httpx.WithHeader("User-Agent", userAgent),
	)
	var resp directAccessResponse
	if err := gitlab.Post(ctx, "/ai/third_party_agents/direct_access", struct{}{}, &resp); err != nil {
		return nil, fmt.Errorf("minting a gateway token: %w", err)
	}
	if resp.Token == "" || resp.ExpiresAt == 0 {
		return nil, httpx.NewError(http.StatusBadGateway, "minting a gateway token: the response has no token or expiry")
	}

	// Every returned header is sent verbatim: the gateway enforces several of
	// them and rejects the request without them.
	opts := make([]httpx.Option, 0, len(resp.Headers)+3)
	for k, v := range resp.Headers {
		opts = append(opts, httpx.WithHeader(k, v))
	}
	opts = append(opts,
		httpx.WithHeader("Authorization", "Bearer "+resp.Token),
		httpx.WithHeader("anthropic-version", anthropicVersion),
		httpx.WithHeader("User-Agent", userAgent),
	)
	return &session{
		client:    httpx.NewClient(proxyURL, opts...),
		expiresAt: time.Unix(resp.ExpiresAt, 0),
	}, nil
}
