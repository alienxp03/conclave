package provider

import (
	"context"

	"github.com/alienxp03/dbate/internal/config"
)

// GeminiProvider implements the Provider interface for Gemini CLI.
type GeminiProvider struct {
	BaseProvider
}

// NewGeminiProvider creates a new Gemini provider with defaults.
func NewGeminiProvider() *GeminiProvider {
	return NewGeminiProviderWithConfig(config.ProviderConfig{
		Command:      "gemini",
		Args:         []string{},
		DefaultModel: "",
		Models:       []string{"pro", "flash", "ultra"},
		Timeout:      0,
		Enabled:      true,
	})
}

// NewGeminiProviderWithConfig creates a Gemini provider from config.
func NewGeminiProviderWithConfig(cfg config.ProviderConfig) *GeminiProvider {
	return &GeminiProvider{
		BaseProvider: NewBaseProvider("gemini", "Google Gemini", cfg),
	}
}

// Generate sends a prompt to Gemini CLI and returns the response.
func (p *GeminiProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *GeminiProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	args := []string{}

	// Add model flag if specified
	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)
	return p.Execute(ctx, args...)
}
