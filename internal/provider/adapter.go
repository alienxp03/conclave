// Package provider provides backward-compatible adapters for the new provider package.
// This package wraps github.com/alienxp03/conclave/provider to maintain compatibility
// with existing code while using the new reusable provider implementation.
package provider

import (
	"context"
	"time"

	"github.com/alienxp03/conclave/provider"
)

// Provider defines the backward-compatible interface for AI providers.
// This wraps the new provider.Provider interface with legacy methods.
type Provider interface {
	// Core interface (same as new provider.Provider)
	Name() string
	Available() bool
	Execute(ctx context.Context, req *provider.Request) (*provider.Response, error)

	// Legacy methods for backward compatibility
	DisplayName() string
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateWithModel(ctx context.Context, prompt, model string) (string, error)
	GenerateWithDir(ctx context.Context, prompt, model, dir string) (string, error)
	GenerateWithResponseDir(ctx context.Context, prompt, model, dir string) (*provider.Response, error)
	Models() []string
	DefaultModel() string
	Timeout() time.Duration
}

// Adapter wraps a provider.Provider to provide backward-compatible methods.
type Adapter struct {
	provider.Provider
}

// NewAdapter creates a new adapter from a provider.Provider.
func NewAdapter(p provider.Provider) *Adapter {
	return &Adapter{Provider: p}
}

// DisplayName returns a human-friendly name.
func (a *Adapter) DisplayName() string {
	// The new provider interface doesn't have DisplayName, use Name
	return a.Name()
}

// Generate sends a prompt and returns the response.
func (a *Adapter) Generate(ctx context.Context, prompt string) (string, error) {
	return a.GenerateWithModel(ctx, prompt, "")
}

// GenerateWithModel sends a prompt with a specific model.
func (a *Adapter) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	return a.GenerateWithDir(ctx, prompt, model, "")
}

// GenerateWithDir sends a prompt with a specific model and working directory.
func (a *Adapter) GenerateWithDir(ctx context.Context, prompt, model, dir string) (string, error) {
	resp, err := a.Execute(ctx, &provider.Request{
		Prompt:     prompt,
		Model:      model,
		WorkingDir: dir,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GenerateWithResponseDir sends a prompt and returns a structured response with metadata.
func (a *Adapter) GenerateWithResponseDir(ctx context.Context, prompt, model, dir string) (*provider.Response, error) {
	return a.Execute(ctx, &provider.Request{
		Prompt:     prompt,
		Model:      model,
		WorkingDir: dir,
	})
}

// Models returns available models.
// Note: This is a helper that's not in the core interface but useful for backward compatibility.
func (a *Adapter) Models() []string {
	// The new provider interface doesn't expose Models directly
	// We can use type assertion to get it from BaseProvider if needed
	if bp, ok := a.Provider.(interface{ Models() []string }); ok {
		return bp.Models()
	}
	return nil
}

// DefaultModel returns the default model.
func (a *Adapter) DefaultModel() string {
	if bp, ok := a.Provider.(interface{ DefaultModel() string }); ok {
		return bp.DefaultModel()
	}
	return ""
}

// Timeout returns the configured timeout.
func (a *Adapter) Timeout() time.Duration {
	if bp, ok := a.Provider.(interface{ Timeout() time.Duration }); ok {
		return bp.Timeout()
	}
	return provider.DefaultTimeout
}

// Registry wraps provider.Registry with backward-compatible adapters.
type Registry struct {
	*provider.Registry
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		Registry: provider.NewRegistry(),
	}
}

// Register adds a provider to the registry.
// Automatically wraps it with an adapter.
func (r *Registry) Register(p provider.Provider) {
	r.Registry.Register(p)
}

// Get retrieves a provider by name and wraps it with an adapter.
func (r *Registry) Get(name string) (Provider, error) {
	p, err := r.Registry.Get(name)
	if err != nil {
		return nil, err
	}
	return NewAdapter(p), nil
}

// List returns all registered providers wrapped with adapters.
func (r *Registry) List() []Provider {
	providers := r.Registry.List()
	adapted := make([]Provider, len(providers))
	for i, p := range providers {
		adapted[i] = NewAdapter(p)
	}
	return adapted
}

// Available returns all available providers wrapped with adapters.
func (r *Registry) Available() []Provider {
	providers := r.Registry.Available()
	adapted := make([]Provider, len(providers))
	for i, p := range providers {
		adapted[i] = NewAdapter(p)
	}
	return adapted
}

// Response is an alias for provider.Response for backward compatibility.
type Response = provider.Response

// CLIError is an alias for provider.CLIError for backward compatibility.
type CLIError = provider.CLIError
