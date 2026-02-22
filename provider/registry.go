package provider

import (
	"fmt"
	"sort"
)

// CloudFactory constructs a cloud provider given an API key.
type CloudFactory func(apiKey string) Provider

// LocalFactory constructs a local provider.
type LocalFactory func() Provider

// Registry is the central mapping of provider names to their factories.
// It eliminates the need for switch statements across the codebase — add a
// provider once here and every CLI handler and server startup path sees it.
type Registry struct {
	clouds map[string]CloudFactory
	locals map[string]LocalFactory
}

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		clouds: make(map[string]CloudFactory),
		locals: make(map[string]LocalFactory),
	}
}

// RegisterCloud registers a cloud provider factory by name.
func (r *Registry) RegisterCloud(name string, f CloudFactory) {
	r.clouds[name] = f
}

// RegisterLocal registers a local provider factory by name.
func (r *Registry) RegisterLocal(name string, f LocalFactory) {
	r.locals[name] = f
}

// NewCloud constructs a cloud provider by name with the given API key.
func (r *Registry) NewCloud(name, apiKey string) (Provider, error) {
	f, ok := r.clouds[name]
	if !ok {
		return nil, fmt.Errorf("unknown cloud provider: %s", name)
	}
	return f(apiKey), nil
}

// NewLocal constructs a local provider by name.
func (r *Registry) NewLocal(name string) (Provider, error) {
	f, ok := r.locals[name]
	if !ok {
		return nil, fmt.Errorf("unknown local runner: %s", name)
	}
	return f(), nil
}

// CloudNames returns all registered cloud provider names, sorted.
func (r *Registry) CloudNames() []string {
	names := make([]string, 0, len(r.clouds))
	for name := range r.clouds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// LocalNames returns all registered local runner names, sorted.
func (r *Registry) LocalNames() []string {
	names := make([]string, 0, len(r.locals))
	for name := range r.locals {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsCloud returns true if the name is a registered cloud provider.
func (r *Registry) IsCloud(name string) bool {
	_, ok := r.clouds[name]
	return ok
}

// IsLocal returns true if the name is a registered local runner.
func (r *Registry) IsLocal(name string) bool {
	_, ok := r.locals[name]
	return ok
}

// IsKnown returns true if the name is any registered provider or runner.
func (r *Registry) IsKnown(name string) bool {
	return r.IsCloud(name) || r.IsLocal(name)
}
