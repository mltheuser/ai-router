package debug

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/providers/httpclient"
)

// Provider wraps a real provider and logs the full request lifecycle.
// It captures 4 data points per call:
//  1. Incoming shared API request
//  2. Outgoing provider-specific HTTP request (via context collector)
//  3. Incoming provider-specific HTTP response (via context collector)
//  4. Outgoing shared API response
type Provider struct {
	inner provider.Provider
	mu    *sync.Mutex // shared across all debug providers for atomic output
	w     io.Writer
}

// WrapProvider decorates a provider with debug logging.
// The mutex should be shared across all debug-wrapped providers so that
// output blocks from concurrent requests never interleave.
func WrapProvider(p provider.Provider, mu *sync.Mutex, w io.Writer) provider.Provider {
	return &Provider{inner: p, mu: mu, w: w}
}

// --- Delegated methods (no debug needed) ---

func (d *Provider) Name() string                                          { return d.inner.Name() }
func (d *Provider) Type() api.ProviderType                                { return d.inner.Type() }
func (d *Provider) Verify(ctx context.Context) error                      { return d.inner.Verify(ctx) }
func (d *Provider) ListModels(ctx context.Context) ([]api.ModelInfo, error) { return d.inner.ListModels(ctx) }

// --- Debug-instrumented methods ---

func (d *Provider) Chat(ctx context.Context, req *api.ChatRequest) (*api.ChatResponse, error) {
	reqJSON := marshalPretty(req)

	dc := &httpclient.DebugCollector{}
	ctx = httpclient.NewDebugContext(ctx, dc)

	start := time.Now()
	resp, err := d.inner.Chat(ctx, req)
	elapsed := time.Since(start)

	respJSON := marshalPretty(resp)
	d.printBlock("Chat", dc, reqJSON, respJSON, elapsed, err)
	return resp, err
}

func (d *Provider) Embed(ctx context.Context, req *api.EmbedRequest) (*api.EmbedResponse, error) {
	reqJSON := marshalPretty(req)

	dc := &httpclient.DebugCollector{}
	ctx = httpclient.NewDebugContext(ctx, dc)

	start := time.Now()
	resp, err := d.inner.Embed(ctx, req)
	elapsed := time.Since(start)

	respJSON := marshalPretty(resp)
	d.printBlock("Embed", dc, reqJSON, respJSON, elapsed, err)
	return resp, err
}

// --- Output formatting ---

func (d *Provider) printBlock(op string, dc *httpclient.DebugCollector, inReq, outResp []byte, elapsed time.Duration, callErr error) {
	reqID := shortID()
	provName := d.inner.Name()
	ts := time.Now().Format(time.RFC3339)

	var buf bytes.Buffer
	line := "══════════════════════════════════════════════════════════════════════"
	thin := "──────────────────────────────────────────────────────────────────────"

	fmt.Fprintf(&buf, "\n╔%s\n", line)
	fmt.Fprintf(&buf, "║ DEBUG [%s] %s → %s\n", reqID, op, provName)
	fmt.Fprintf(&buf, "║ %s\n", ts)
	fmt.Fprintf(&buf, "╠%s\n", line)

	// 1. Incoming shared API request
	fmt.Fprintf(&buf, "║ ► INCOMING REQUEST (shared API)\n")
	writeIndented(&buf, inReq)
	fmt.Fprintf(&buf, "╠%s\n", thin)

	// 2. Outgoing provider-specific request
	fmt.Fprintf(&buf, "║ ► OUTGOING PROVIDER REQUEST (%s)\n", provName)
	if dc.RequestMethod != "" {
		fmt.Fprintf(&buf, "║ %s %s\n", dc.RequestMethod, dc.RequestURL)
	}
	writeIndented(&buf, prettyJSON(dc.RequestBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	// 3. Incoming provider-specific response
	fmt.Fprintf(&buf, "║ ► INCOMING PROVIDER RESPONSE (%s)\n", provName)
	writeIndented(&buf, prettyJSON(dc.ResponseBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	// 4. Outgoing shared API response
	fmt.Fprintf(&buf, "║ ► OUTGOING RESPONSE (shared API)\n")
	if callErr != nil {
		fmt.Fprintf(&buf, "║ ERROR: %s\n", callErr)
	}
	writeIndented(&buf, outResp)
	fmt.Fprintf(&buf, "╠%s\n", thin)

	fmt.Fprintf(&buf, "║ Duration: %s\n", elapsed.Round(time.Millisecond))
	fmt.Fprintf(&buf, "╚%s\n", line)

	// Atomic write: lock so concurrent requests don't interleave
	d.mu.Lock()
	d.w.Write(buf.Bytes())
	d.mu.Unlock()
}

// --- Helpers ---

func marshalPretty(v interface{}) []byte {
	if v == nil {
		return []byte("<nil>")
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return []byte(fmt.Sprintf("<marshal error: %s>", err))
	}
	return b
}

func prettyJSON(raw []byte) []byte {
	if len(raw) == 0 {
		return []byte("<empty>")
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return raw // already not valid JSON, return as-is
	}
	return buf.Bytes()
}

func writeIndented(buf *bytes.Buffer, data []byte) {
	for _, line := range bytes.Split(data, []byte("\n")) {
		fmt.Fprintf(buf, "║ %s\n", line)
	}
}

func shortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
