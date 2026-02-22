package api

type TestRequest struct {
	Provider string `json:"provider"`
	// Model is the raw model identifier (e.g. "llama3.2") without routing
	// suffixes like ":cloud", ":local", or "@provider". The routing context
	// is inferred from the Provider field.
	Model string `json:"model,omitempty"`
}

type TestStatus string

const (
	StatusPass    TestStatus = "pass"
	StatusFail    TestStatus = "fail"
	StatusSkipped TestStatus = "skipped"
)

// Check is a reusable status + description pair used across the test report.
// A pass or skip may omit the error; a fail should always explain why.
type Check struct {
	Status      TestStatus `json:"status"`
	Description string     `json:"description,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type ProviderHealth struct {
	Verify     Check `json:"verify"`
	ListModels Check `json:"list_models"`
}

type TestReport struct {
	Provider  string           `json:"provider"`
	Checks    ProviderHealth   `json:"checks"`
	Scenarios []ScenarioResult `json:"scenarios"`
}

type ScenarioResult struct {
	Name   string  `json:"name"`
	Model  string  `json:"model,omitempty"`
	Checks []Check `json:"checks"`
}

// NewResult creates an empty ScenarioResult to accumulate checks into.
func NewResult() *ScenarioResult {
	return &ScenarioResult{}
}

// Pass appends a passing check with a description of what was tested.
func (r *ScenarioResult) Pass(description string) {
	r.Checks = append(r.Checks, Check{Status: StatusPass, Description: description})
}

// Fail appends a failing check with a description and error message.
func (r *ScenarioResult) Fail(description string, err string) {
	r.Checks = append(r.Checks, Check{Status: StatusFail, Description: description, Error: err})
}
