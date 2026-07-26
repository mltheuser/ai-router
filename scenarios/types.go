package scenarios

import (
	"context"
	"time"

	"github.com/mltheuser/ai-router/api"
)

// DefaultTimeout is the per-scenario execution budget applied unless a scenario
// implements TimeoutOverrider to request more.
const DefaultTimeout = 60 * time.Second

// Scenario defines a test capability scenario.
type Scenario interface {
	// Name returns the unique identifier for this scenario.
	Name() string

	// Description explains what this scenario tests.
	Description() string

	// RequiredCapabilities returns the list of capabilities a model must have
	// to run this scenario.
	RequiredCapabilities() []api.Capability

	// Run executes the test against the running server and returns a result
	// containing one or more checks. Each check reports a pass or fail for
	// a specific sub-test within the scenario. The Name and Model fields
	// do not need to be set by the scenario; the caller fills them in.
	Run(ctx context.Context, baseURL string, modelID string) *api.ScenarioResult
}

// TimeoutOverrider is an optional interface a scenario may implement to request
// a longer execution budget than DefaultTimeout (e.g. reasoning models that
// stream long traces at high effort).
type TimeoutOverrider interface {
	Timeout() time.Duration
}

// TimeoutFor returns the execution budget for a scenario: its override if it
// implements TimeoutOverrider, otherwise DefaultTimeout.
func TimeoutFor(s Scenario) time.Duration {
	if o, ok := s.(TimeoutOverrider); ok {
		return o.Timeout()
	}
	return DefaultTimeout
}

var registry = make(map[string]Scenario)

// Register adds a scenario to the global registry.
func Register(s Scenario) {
	registry[s.Name()] = s
}

// Get returns a scenario by name.
func Get(name string) Scenario {
	return registry[name]
}

// List returns all registered scenarios.
func List() []Scenario {
	var list []Scenario
	for _, s := range registry {
		list = append(list, s)
	}
	return list
}
