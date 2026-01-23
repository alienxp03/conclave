// Package generic provides a generic CLI provider implementation.
// This package can be used for any custom CLI tool that doesn't have
// a specialized provider implementation.
package generic

import (
	"context"
	"time"

	"github.com/alienxp03/conclave/provider"
)

// Provider is a generic configurable provider for custom CLI tools.
type Provider struct {
	provider.BaseProvider
}

// New creates a generic provider from configuration.
func New(cfg provider.Config) *Provider {
	return &Provider{
		BaseProvider: provider.NewBaseProvider(cfg),
	}
}

// Execute sends a request and returns the raw response.
// Generic providers don't attempt to parse structured metadata.
func (p *Provider) Execute(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	// Build arguments
	args := []string{}

	// Add model flag if specified
	model := req.Model
	if model == "" {
		model = p.DefaultModel()
	}
	if model != "" {
		args = append(args, "--model", model)
	}

	// Add the prompt
	args = append(args, req.Prompt)

	// Add any custom args
	if len(req.Args) > 0 {
		args = append(args, req.Args...)
	}

	// Execute command
	execReq := &provider.Request{
		Prompt:     req.Prompt,
		Model:      model,
		WorkingDir: req.WorkingDir,
		Args:       args,
	}

	start := time.Now()
	content, err := p.ExecuteCommand(ctx, execReq)
	if err != nil {
		return nil, err
	}
	duration := time.Since(start)

	return &provider.Response{
		Content:  content,
		Model:    model,
		Provider: p.Name(),
		Metadata: &provider.Metadata{
			Duration: duration,
		},
		Raw: content,
	}, nil
}

// HealthCheck performs a quick health check using the provider execution path.
func (p *Provider) HealthCheck(ctx context.Context) provider.HealthStatus {
	return provider.HealthCheckWithExecute(ctx, p.DefaultModel(), p.Execute)
}
