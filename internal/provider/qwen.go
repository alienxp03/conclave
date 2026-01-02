package provider

import (
	"context"
	"time"

	"github.com/alienxp03/conclave/internal/config"
	"github.com/alienxp03/conclave/internal/core"
)

// QwenProvider implements the Provider interface for Qwen CLI.
type QwenProvider struct {
	BaseProvider
	useJSON bool
}

// NewQwenProvider creates a new Qwen provider with defaults.
func NewQwenProvider() *QwenProvider {
	return NewQwenProviderWithConfig(config.ProviderConfig{
		Command:      core.DefaultCommandForProvider["qwen"],
		Args:         core.DefaultArgsForProvider["qwen"],
		DefaultModel: core.DefaultModelForProvider["qwen"],
		Models:       core.DefaultModelsForProvider["qwen"],
		Timeout:      0,
		Enabled:      true,
	})
}

// NewQwenProviderWithConfig creates a Qwen provider from config.
func NewQwenProviderWithConfig(cfg config.ProviderConfig) *QwenProvider {
	return &QwenProvider{
		BaseProvider: NewBaseProvider("qwen", "Alibaba Qwen", cfg),
		useJSON:      true,
	}
}

// Generate sends a prompt to Qwen CLI and returns the response.
func (p *QwenProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *QwenProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
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
		resp, parseErr := ParseQwenJSON(rawOutput)
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
func (p *QwenProvider) GenerateWithResponse(ctx context.Context, prompt, model string) (*Response, error) {
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

	resp, parseErr := ParseQwenJSON(rawOutput)
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
