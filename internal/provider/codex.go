package provider

import (
	"context"

	"github.com/alienxp03/conclave/internal/config"
	"github.com/alienxp03/conclave/internal/core"
)

// CodexProvider implements the Provider interface for OpenAI Codex CLI.
type CodexProvider struct {
	BaseProvider
	useJSON bool
}

// NewCodexProvider creates a new Codex provider with defaults.
func NewCodexProvider() *CodexProvider {
	return NewCodexProviderWithConfig(config.ProviderConfig{
		Command:      core.DefaultCommandForProvider["codex"],
		Args:         core.DefaultArgsForProvider["codex"],
		DefaultModel: core.DefaultModelForProvider["codex"],
		Models:       core.DefaultModelsForProvider["codex"],
		Timeout:      0,
		Enabled:      true,
	})
}

// NewCodexProviderWithConfig creates a Codex provider from config.
func NewCodexProviderWithConfig(cfg config.ProviderConfig) *CodexProvider {
	return &CodexProvider{
		BaseProvider: NewBaseProvider("codex", "OpenAI Codex", cfg),
		useJSON:      true,
	}
}

// Generate sends a prompt to Codex CLI and returns the response.
func (p *CodexProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *CodexProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	return p.GenerateWithDir(ctx, prompt, model, "")
}

// GenerateWithDir sends a prompt with a specific model and working directory.
func (p *CodexProvider) GenerateWithDir(ctx context.Context, prompt, model, dir string) (string, error) {
	args := []string{"exec"}

	// Use JSON streaming output for metadata capture
	if p.useJSON {
		args = append(args, "--json")
	}

	// Add model flag if specified
	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)

	rawOutput, err := p.ExecuteWithDir(ctx, dir, args...)
	if err != nil {
		return "", err
	}

	// Parse JSON response if using JSON mode
	if p.useJSON {
		resp, parseErr := ParseCodexJSON(rawOutput)
		if parseErr != nil {
			return rawOutput, nil
		}
		resp.Provider = p.name
		return resp.Content, nil
	}

	return rawOutput, nil
}

// GenerateWithResponse sends a prompt and returns a structured response with metadata.
func (p *CodexProvider) GenerateWithResponse(ctx context.Context, prompt, model string) (*Response, error) {
	return p.GenerateWithResponseDir(ctx, prompt, model, "")
}

// GenerateWithResponseDir sends a prompt with working directory and returns structured response.
func (p *CodexProvider) GenerateWithResponseDir(ctx context.Context, prompt, model, dir string) (*Response, error) {
	args := []string{"exec", "--json"}

	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)

	rawOutput, err := p.ExecuteWithDir(ctx, dir, args...)
	if err != nil {
		return nil, err
	}

	resp, parseErr := ParseCodexJSON(rawOutput)
	if parseErr != nil {
		return &Response{
			Content:  rawOutput,
			Provider: p.name,
		}, nil
	}

	resp.Provider = p.name
	return resp, nil
}
