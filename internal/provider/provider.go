// Package provider contains AI provider abstractions and implementations.
package provider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alienxp03/dbate/internal/config"
)

// Provider defines the interface for AI providers.
type Provider interface {
	// Name returns the provider's identifier.
	Name() string

	// DisplayName returns a human-friendly name.
	DisplayName() string

	// Generate sends a prompt and returns the response.
	Generate(ctx context.Context, prompt string) (string, error)

	// GenerateWithModel sends a prompt with a specific model.
	GenerateWithModel(ctx context.Context, prompt, model string) (string, error)

	// Available checks if the provider's CLI is installed and accessible.
	Available() bool

	// Models returns available models for this provider.
	Models() []string

	// DefaultModel returns the default model.
	DefaultModel() string

	// Timeout returns the configured timeout.
	Timeout() time.Duration
}

// Registry manages available AI providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return p, nil
}

// List returns all registered providers.
func (r *Registry) List() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}

// Available returns all providers that are currently available.
func (r *Registry) Available() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var available []Provider
	for _, p := range r.providers {
		if p.Available() {
			available = append(available, p)
		}
	}
	return available
}

// DefaultRegistry creates a registry with default provider configurations.
func DefaultRegistry() *Registry {
	cfg := config.Default()
	return RegistryFromConfig(cfg)
}

// RegistryFromConfig creates a registry from configuration.
func RegistryFromConfig(cfg *config.Config) *Registry {
	r := NewRegistry()

	for name, provCfg := range cfg.Providers {
		if !provCfg.Enabled {
			continue
		}

		var p Provider
		switch name {
		case "claude":
			p = NewClaudeProviderWithConfig(provCfg)
		case "codex":
			p = NewCodexProviderWithConfig(provCfg)
		case "gemini":
			p = NewGeminiProviderWithConfig(provCfg)
		case "qwen":
			p = NewQwenProviderWithConfig(provCfg)
		default:
			// Generic CLI provider for custom providers
			p = NewGenericProvider(name, provCfg)
		}

		r.Register(p)
	}

	return r
}
