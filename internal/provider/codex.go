package provider

import (
	"context"

	"github.com/alienxp03/dbate/internal/config"
)

// CodexProvider implements the Provider interface for OpenAI Codex CLI.
type CodexProvider struct {
	BaseProvider
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
	}
}

// Generate sends a prompt to Codex CLI and returns the response.
func (p *CodexProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *CodexProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	args := []string{}

	// Add model flag if specified
	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)
	return p.Execute(ctx, args...)
}
