package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultTimeout is a scenario's execution budget unless it sets its own.
const DefaultTimeout = 60 * time.Second

// Scenario is one end-to-end check of a use case.
type Scenario[M Model] struct {
	// Name identifies the scenario in test reports.
	Name string

	// Applies reports whether the scenario can run against m, e.g. because
	// it needs a feature only some models have. Nil means every model.
	Applies func(m M) bool

	// Timeout overrides DefaultTimeout when set.
	Timeout time.Duration

	// Run exercises the endpoint at url with the fully-qualified model string
	// and records its findings in res.
	Run func(ctx context.Context, url, model string, res *Result)
}

// TestRequest is the body of POST /v1/test.
type TestRequest struct {
	Provider string `json:"provider"`
	// UseCase, if set, limits the test to one use case.
	UseCase string `json:"use_case,omitempty"`
	// Model, if set, pins the model every scenario runs against by its bare
	// ID; the provider comes from Provider.
	Model string `json:"model,omitempty"`
	// URL is where the server under test serves.
	URL string `json:"-"`
}

// Status is the outcome of a single check.
type Status string

// Check statuses.
const (
	Pass    Status = "pass"
	Fail    Status = "fail"
	Skipped Status = "skipped"
)

// Check is one verified fact. A pass or skip may omit the error; a fail
// always explains why.
type Check struct {
	Status      Status `json:"status"`
	Description string `json:"description,omitempty"`
	Error       string `json:"error,omitempty"`
}

// Result collects the checks of one scenario run.
type Result struct {
	Name   string  `json:"name"`
	Model  string  `json:"model,omitempty"`
	Checks []Check `json:"checks"`
}

// Pass records a passing check.
func (r *Result) Pass(description string) {
	r.Checks = append(r.Checks, Check{Status: Pass, Description: description})
}

// Fail records a failing check.
func (r *Result) Fail(description, err string) {
	r.Checks = append(r.Checks, Check{Status: Fail, Description: description, Error: err})
}

func (r *Result) skip(reason string) {
	r.Checks = append(r.Checks, Check{Status: Skipped, Error: reason})
}

// Report is the outcome of testing one provider's implementation of one use
// case: whether it serves any models, then every scenario.
type Report struct {
	Models    Check    `json:"models"`
	Scenarios []Result `json:"scenarios"`
}

func (b *Base[M, P]) Test(ctx context.Context, req TestRequest) (Report, bool) {
	if _, ok := b.providers[req.Provider]; !ok {
		return Report{}, false
	}

	// Test exactly what the router serves: the models listed at startup.
	report := Report{Scenarios: []Result{}}
	models := b.models[req.Provider]
	if len(models) == 0 {
		report.Models = Check{Status: Fail, Error: "no models served: the listing failed at startup (see the log) or the provider offers none"}
		return report, true
	}
	report.Models = Check{Status: Pass, Description: fmt.Sprintf("%d models served", len(models))}

	for _, sc := range b.spec.Scenarios {
		res := Result{Name: sc.Name}
		if m, reason := pickModel(sc, models, req.Model, b.spec.Prefer); reason != "" {
			res.skip(reason)
		} else {
			res.Model = m.Ref().ID
			timeout := sc.Timeout
			if timeout == 0 {
				timeout = DefaultTimeout
			}
			scCtx, cancel := context.WithTimeout(ctx, timeout)
			sc.Run(scCtx, req.URL+"/v1/"+b.spec.Name, m.Ref().Model, &res)
			cancel()
		}
		report.Scenarios = append(report.Scenarios, res)
	}
	return report, true
}

// pickModel chooses the model a scenario runs against: the pinned one if
// pinned, otherwise the preferred model the scenario applies to. A non-empty
// reason means the scenario cannot run.
func pickModel[M Model](sc Scenario[M], models []M, pinned string, prefer func(a, b M) bool) (best M, reason string) {
	applies := func(m M) bool { return sc.Applies == nil || sc.Applies(m) }

	if pinned != "" {
		for _, m := range models {
			if m.Ref().ID != pinned {
				continue
			}
			if !applies(m) {
				return best, fmt.Sprintf("model '%s' does not support this scenario", pinned)
			}
			return m, ""
		}
		return best, fmt.Sprintf("model '%s' is not among the provider's served models", pinned)
	}

	found := false
	for _, m := range models {
		if applies(m) && (!found || prefer(m, best)) {
			best, found = m, true
		}
	}
	if !found {
		return best, "no served model supports this scenario"
	}
	return best, ""
}

// PostJSON is the HTTP call scenarios make: it sends body as JSON to url and
// decodes a 200 response into a Res. Any other status becomes an error
// carrying the response body, which explains what went wrong.
func PostJSON[Res any](ctx context.Context, url string, body any) (*Res, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, bytes.TrimSpace(respBody))
	}

	var res Res
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &res, nil
}

// LessKnown orders two optional quantities, such as prices or sizes, for a
// Spec.Prefer: it reports whether a is smaller than b, where an unknown (nil)
// value ranks after every known one.
func LessKnown[T int64 | float64](a, b *T) bool {
	switch {
	case a == nil:
		return false
	case b == nil:
		return true
	default:
		return *a < *b
	}
}
