// Package debug logs the full lifecycle of API requests for troubleshooting.
// Each logged request shows four bodies: the shared API request and response
// (captured by Middleware at the HTTP boundary) and the provider-specific
// request and response in between (recorded by the provider's HTTP client
// into the Exchange that Middleware attaches to the request context).
package debug

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Exchange is the provider-specific HTTP exchange of one API request. When a
// provider makes several calls, the last one wins.
type Exchange struct {
	Method       string
	URL          string
	RequestBody  []byte
	ResponseBody []byte
}

type exchangeKey struct{}

// ExchangeFrom returns the Exchange to record into, or nil when the request is
// not being debug-logged.
func ExchangeFrom(ctx context.Context) *Exchange {
	ex, _ := ctx.Value(exchangeKey{}).(*Exchange)
	return ex
}

// Middleware returns an HTTP middleware that writes one debug block per
// request to w. Blocks are written atomically, so concurrent requests never
// interleave.
func Middleware(w io.Writer) func(http.Handler) http.Handler {
	var mu sync.Mutex
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			reqBody, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(reqBody))

			ex := &Exchange{}
			r = r.WithContext(context.WithValue(r.Context(), exchangeKey{}, ex))
			rec := &recorder{ResponseWriter: rw}

			start := time.Now()
			next.ServeHTTP(rec, r)
			block := format(r, ex, reqBody, rec.body.Bytes(), time.Since(start))

			mu.Lock()
			_, _ = w.Write(block)
			mu.Unlock()
		})
	}
}

// recorder tees the response body so it can be logged after it was sent.
type recorder struct {
	http.ResponseWriter
	body bytes.Buffer
}

func (r *recorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Unwrap exposes the wrapped writer to http.ResponseController.
func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

func format(r *http.Request, ex *Exchange, reqBody, respBody []byte, elapsed time.Duration) []byte {
	const (
		line = "══════════════════════════════════════════════════════════════════════"
		thin = "──────────────────────────────────────────────────────────────────────"
	)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "\n╔%s\n", line)
	fmt.Fprintf(&buf, "║ DEBUG [%s] %s %s\n", shortID(), r.Method, r.URL.Path)
	fmt.Fprintf(&buf, "║ %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&buf, "╠%s\n", line)

	fmt.Fprintf(&buf, "║ ► INCOMING REQUEST (shared API)\n")
	writeIndented(&buf, prettyJSON(reqBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	fmt.Fprintf(&buf, "║ ► OUTGOING PROVIDER REQUEST\n")
	if ex.Method != "" {
		fmt.Fprintf(&buf, "║ %s %s\n", ex.Method, ex.URL)
	}
	writeIndented(&buf, prettyJSON(ex.RequestBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	fmt.Fprintf(&buf, "║ ► INCOMING PROVIDER RESPONSE\n")
	writeIndented(&buf, prettyJSON(ex.ResponseBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	fmt.Fprintf(&buf, "║ ► OUTGOING RESPONSE (shared API)\n")
	writeIndented(&buf, prettyJSON(respBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	fmt.Fprintf(&buf, "║ Duration: %s\n", elapsed.Round(time.Millisecond))
	fmt.Fprintf(&buf, "╚%s\n", line)
	return buf.Bytes()
}

func prettyJSON(raw []byte) []byte {
	if len(raw) == 0 {
		return []byte("<empty>")
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, bytes.TrimSpace(raw), "", "  "); err != nil {
		return raw // not JSON; show as-is
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
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
