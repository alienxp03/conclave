package provider

import (
	"context"
	"time"

	"github.com/alienxp03/dbate/internal/config"
)

// CodexProvider implements the Provider interface for OpenAI Codex CLI.
type CodexProvider struct {
	BaseProvider
	useJSON bool
}

// NewCodexProvider creates a new Codex provider with defaults.
func NewCodexProvider() *CodexProvider {
	return NewCodexProviderWithConfig(config.ProviderConfig{
		Command:      "codex",
		Args:         []string{},
		DefaultModel: "",
		Models:       []string{"gpt-4", "gpt-4o", "gpt-3.5-turbo"},
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
	rawOutput, err := p.Execute(ctx, args...)
	if err != nil {
		return "", err
	}

	// Parse JSON response if using JSON mode
	if p.useJSON {
		resp, parseErr := ParseCodexJSON(rawOutput)
		if parseErr != nil {
			return rawOutput, nil
		}
		resp.Duration = time.Since(start)
		resp.Provider = p.name
		return resp.Content, nil
	}

	return rawOutput, nil
}

// GenerateWithResponse sends a prompt and returns a structured response with metadata.
func (p *CodexProvider) GenerateWithResponse(ctx context.Context, prompt, model string) (*Response, error) {
	args := []string{}
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

	resp, parseErr := ParseCodexJSON(rawOutput)
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
