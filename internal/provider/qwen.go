package provider

import (
	"context"

	"github.com/alienxp03/dbate/internal/config"
)

// QwenProvider implements the Provider interface for Qwen CLI.
type QwenProvider struct {
	BaseProvider
}

// NewQwenProvider creates a new Qwen provider with defaults.
func NewQwenProvider() *QwenProvider {
	return NewQwenProviderWithConfig(config.ProviderConfig{
		Command:      "qwen",
		Args:         []string{},
		DefaultModel: "",
		Models:       []string{"qwen-turbo", "qwen-plus", "qwen-max"},
		Timeout:      0,
		Enabled:      true,
	})
}

// NewQwenProviderWithConfig creates a Qwen provider from config.
func NewQwenProviderWithConfig(cfg config.ProviderConfig) *QwenProvider {
	return &QwenProvider{
		BaseProvider: NewBaseProvider("qwen", "Alibaba Qwen", cfg),
	}
}

// Generate sends a prompt to Qwen CLI and returns the response.
func (p *QwenProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *QwenProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	args := []string{}

	// Add model flag if specified
	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)
	return p.Execute(ctx, args...)
}
