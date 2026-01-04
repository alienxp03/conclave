// Package opencode provides an Opencode CLI provider implementation.
package opencode

import (
	"context"
	"time"

	"github.com/alienxp03/conclave/provider"
)

// Provider implements the provider.Provider interface for Opencode CLI.
type Provider struct {
	provider.BaseProvider
}

// New creates a new Opencode provider with the given configuration.
func New(cfg provider.Config) *Provider {
	return &Provider{
		BaseProvider: provider.NewBaseProvider(cfg),
	}
}

// Execute sends a request to Opencode CLI and returns a structured response.
func (p *Provider) Execute(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	// Build arguments
	args := []string{"--json"}

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
	rawOutput, err := p.ExecuteCommand(ctx, execReq)
	if err != nil {
		return nil, err
	}
	duration := time.Since(start)

	// Parse JSON response
	resp, parseErr := ParseJSON(rawOutput, duration)
	if parseErr != nil {
		// Fall back to raw output if parsing fails
		return &provider.Response{
			Content:  rawOutput,
			Provider: p.Name(),
			Model:    model,
		}, nil
	}

	resp.Provider = p.Name()
	if model != "" && resp.Model == "" {
		resp.Model = model
	}

	return resp, nil
}
