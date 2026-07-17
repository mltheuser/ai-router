package api

// ProviderType distinguishes cloud from local providers.
type ProviderType string

// Provider types.
const (
	ProviderTypeCloud ProviderType = "cloud"
	ProviderTypeLocal ProviderType = "local"
)

// Capability represents a model capability.
type Capability string

// Model capabilities.
const (
	CapabilityChat             Capability = "chat"
	CapabilityEmbed            Capability = "embed"
	CapabilityStructuredOutput Capability = "structured_output"
	CapabilityReasoning        Capability = "reasoning"
	CapabilityTools            Capability = "tools"
	CapabilityVision           Capability = "vision"
)

// ModelInfo describes a model available through a specific provider.
type ModelInfo struct {
	ID string `json:"id"`
	// Model is the fully-qualified string ("id:provider_type@provider")
	// to pass verbatim as `model` in requests to address this entry.
	Model        string       `json:"model"`
	Provider     string       `json:"provider"`
	ProviderType ProviderType `json:"provider_type"`
	Capabilities []Capability `json:"capabilities"`

	// Cloud-specific metadata (nil implies unknown, 0 implies free)
	ContextWindow  int      `json:"context_window,omitempty"`
	CostPerMInput  *float64 `json:"cost_per_m_input,omitempty"`
	CostPerMOutput *float64 `json:"cost_per_m_output,omitempty"`

	// Local-specific metadata (zero values for cloud models)
	SizeBytes *int64 `json:"size_bytes,omitempty"`
}

// HasCapability checks if the model supports a given capability.
func (m ModelInfo) HasCapability(capability Capability) bool {
	for _, c := range m.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// ModelList is the response format for listing models.
type ModelList struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}
