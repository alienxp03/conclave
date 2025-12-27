package provider

import (
	"context"

	"github.com/alienxp03/dbate/internal/config"
)

// ClaudeProvider implements the Provider interface for Claude CLI.
type ClaudeProvider struct {
	BaseProvider
}

// NewClaudeProvider creates a new Claude provider with defaults.
func NewClaudeProvider() *ClaudeProvider {
	return NewClaudeProviderWithConfig(config.ProviderConfig{
		Command:      "claude",
		Args:         []string{"--print"},
		DefaultModel: "",
		Models:       []string{"opus", "sonnet", "haiku"},
		Timeout:      0, // Uses default
		Enabled:      true,
	})
}

// NewClaudeProviderWithConfig creates a Claude provider from config.
func NewClaudeProviderWithConfig(cfg config.ProviderConfig) *ClaudeProvider {
	return &ClaudeProvider{
		BaseProvider: NewBaseProvider("claude", "Claude", cfg),
	}
}

// Generate sends a prompt to Claude CLI and returns the response.
func (p *ClaudeProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *ClaudeProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	args := []string{}

	// Add model flag if specified
	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)
	return p.Execute(ctx, args...)
}
