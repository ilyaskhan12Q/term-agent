package model

import (
	"fmt"
	"sync"
)

// ProviderRegistry manages registered LLM providers.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]ModelProvider
}

// NewProviderRegistry constructs a new ProviderRegistry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]ModelProvider),
	}
}

// Register registers a provider.
func (r *ProviderRegistry) Register(p ModelProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// Get retrieves a provider by name.
func (r *ProviderRegistry) Get(name string) (ModelProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return p, nil
}
