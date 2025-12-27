package provider

import (
	"context"
	"time"

	"github.com/alienxp03/dbate/internal/config"
)

// GeminiProvider implements the Provider interface for Gemini CLI.
type GeminiProvider struct {
	BaseProvider
	useJSON bool
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
		useJSON:      true,
	}
}

// Generate sends a prompt to Gemini CLI and returns the response.
func (p *GeminiProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *GeminiProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
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
		resp, parseErr := ParseGeminiJSON(rawOutput)
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
func (p *GeminiProvider) GenerateWithResponse(ctx context.Context, prompt, model string) (*Response, error) {
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

	resp, parseErr := ParseGeminiJSON(rawOutput)
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
