// Package debug logs the full lifecycle of API requests for troubleshooting:
// the shared API request and response, and the provider-specific exchange in
// between.
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
	"regexp"
	"sync"
	"time"
	"unicode/utf8"
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
	writeIndented(&buf, renderBody(reqBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	fmt.Fprintf(&buf, "║ ► OUTGOING PROVIDER REQUEST\n")
	if ex.Method != "" {
		fmt.Fprintf(&buf, "║ %s %s\n", ex.Method, ex.URL)
	}
	writeIndented(&buf, renderBody(ex.RequestBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	fmt.Fprintf(&buf, "║ ► INCOMING PROVIDER RESPONSE\n")
	writeIndented(&buf, renderBody(ex.ResponseBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	fmt.Fprintf(&buf, "║ ► OUTGOING RESPONSE (shared API)\n")
	writeIndented(&buf, renderBody(respBody))
	fmt.Fprintf(&buf, "╠%s\n", thin)

	fmt.Fprintf(&buf, "║ Duration: %s\n", elapsed.Round(time.Millisecond))
	fmt.Fprintf(&buf, "╚%s\n", line)
	return buf.Bytes()
}

// renderBody renders a body for the log: JSON indented, with long base64
// strings (inline audio or images) elided; other text as is; binary data
// (e.g. audio) as its size.
func renderBody(raw []byte) []byte {
	if len(raw) == 0 {
		return []byte("<empty>")
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, bytes.TrimSpace(raw), "", "  "); err != nil {
		if isBinary(raw) {
			return fmt.Appendf(nil, "<%d bytes of binary data>", len(raw))
		}
		return raw
	}
	return elideBase64(buf.Bytes())
}

// base64String matches a JSON string holding a long base64 payload, optionally
// as a data URI. In indented JSON, a quote, a run of base64 characters and a
// quote can only be one whole string. Prose contains spaces and never matches.
var base64String = regexp.MustCompile(`"(data:[^;"]+;base64,)?([A-Za-z0-9+/=_-]{1000,})"`)

func elideBase64(js []byte) []byte {
	return base64String.ReplaceAllFunc(js, func(m []byte) []byte {
		sub := base64String.FindSubmatch(m)
		return fmt.Appendf(nil, `"%s<base64, %d chars>"`, sub[1], len(sub[2]))
	})
}

// isBinary reports whether data is not text: invalid UTF-8, or control
// characters other than tab, newline and carriage return.
func isBinary(data []byte) bool {
	if !utf8.Valid(data) {
		return true
	}
	for _, b := range data {
		if b < 0x20 && b != '\t' && b != '\n' && b != '\r' {
			return true
		}
	}
	return false
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
