package provider

import (
	"context"
	"time"

	"github.com/alienxp03/conclave/internal/config"
	"github.com/alienxp03/conclave/internal/core"
)

// ClaudeProvider implements the Provider interface for Claude CLI.
type ClaudeProvider struct {
	BaseProvider
	useJSON bool
}

// NewClaudeProvider creates a new Claude provider with defaults.
func NewClaudeProvider() *ClaudeProvider {
	return NewClaudeProviderWithConfig(config.ProviderConfig{
		Command:      core.DefaultCommandForProvider["claude"],
		Args:         core.DefaultArgsForProvider["claude"],
		DefaultModel: core.DefaultModelForProvider["claude"],
		Models:       core.DefaultModelsForProvider["claude"],
		Timeout:      0, // Uses default
		Enabled:      true,
	})
}

// NewClaudeProviderWithConfig creates a Claude provider from config.
func NewClaudeProviderWithConfig(cfg config.ProviderConfig) *ClaudeProvider {
	return &ClaudeProvider{
		BaseProvider: NewBaseProvider("claude", "Claude", cfg),
		useJSON:      true, // Enable JSON output by default
	}
}

// Generate sends a prompt to Claude CLI and returns the response.
func (p *ClaudeProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *ClaudeProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	return p.GenerateWithDir(ctx, prompt, model, "")
}

// GenerateWithDir sends a prompt with a specific model and working directory.
func (p *ClaudeProvider) GenerateWithDir(ctx context.Context, prompt, model, dir string) (string, error) {
	args := []string{}

	// Use JSON output format for structured responses
	if p.useJSON {
		args = append(args, "--output-format", "json")
	}

	// Add model flag if specified
	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)

	start := time.Now()
	rawOutput, err := p.ExecuteWithDir(ctx, dir, args...)
	if err != nil {
		return "", err
	}

	// Parse JSON response if using JSON mode
	if p.useJSON {
		resp, parseErr := ParseClaudeJSON(rawOutput)
		if parseErr != nil {
			// Fall back to raw output if parsing fails
			return rawOutput, nil
		}
		resp.Duration = time.Since(start)
		resp.Provider = p.name
		return resp.Content, nil
	}

	return rawOutput, nil
}

// GenerateWithResponse sends a prompt and returns a structured response with metadata.
func (p *ClaudeProvider) GenerateWithResponse(ctx context.Context, prompt, model string) (*Response, error) {
	args := []string{}

	// Always use JSON output for structured responses
	args = append(args, "--output-format", "json")

	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)

	start := time.Now()
	rawOutput, err := p.Execute(ctx, args...)
	if err != nil {
		return nil, err
	}

	resp, parseErr := ParseClaudeJSON(rawOutput)
	if parseErr != nil {
		return &Response{
			Content:  rawOutput,
			Provider: p.name,
			Duration: time.Since(start),
		}, nil
	}

	resp.Duration = time.Since(start)
	resp.Provider = p.name
	return resp, nil
}
