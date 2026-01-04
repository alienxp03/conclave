package provider

import (
	"context"
	"log/slog"
	"time"

	"github.com/alienxp03/conclave/internal/config"
	"github.com/alienxp03/conclave/internal/core"
)

// OpencodeProvider implements the Provider interface for Opencode CLI.
type OpencodeProvider struct {
	BaseProvider
	useJSON bool
}

// NewOpencodeProvider creates a new Opencode provider with defaults.
func NewOpencodeProvider() *OpencodeProvider {
	return NewOpencodeProviderWithConfig(config.ProviderConfig{
		Command:      core.DefaultCommandForProvider["opencode"],
		Args:         core.DefaultArgsForProvider["opencode"],
		DefaultModel: core.DefaultModelForProvider["opencode"],
		Models:       core.DefaultModelsForProvider["opencode"],
		Timeout:      0, // Uses default
		Enabled:      true,
	})
}

// NewOpencodeProviderWithConfig creates an Opencode provider from config.
func NewOpencodeProviderWithConfig(cfg config.ProviderConfig) *OpencodeProvider {
	return &OpencodeProvider{
		BaseProvider: NewBaseProvider("opencode", "Opencode", cfg),
		useJSON:      true,
	}
}

// Generate sends a prompt to Opencode CLI and returns the response.
func (p *OpencodeProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *OpencodeProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	return p.GenerateWithDir(ctx, prompt, model, "")
}

// GenerateWithDir sends a prompt with a specific model and working directory.
func (p *OpencodeProvider) GenerateWithDir(ctx context.Context, prompt, model, dir string) (string, error) {
	args := []string{"run"}

	if p.useJSON {
		args = append(args, "--format", "json")
	}

	// Add model flag if specified
	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)

	slog.Info("Opencode generating response", "model", model, "prompt_len", len(prompt), "dir", dir)
	slog.Debug("Opencode prompt", "content", prompt)

	rawOutput, err := p.ExecuteWithDir(ctx, dir, args...)
	if err != nil {
		slog.Error("Opencode generation failed", "error", err)
		return "", err
	}

	slog.Debug("Opencode raw output", "output", rawOutput)

	if p.useJSON {
		resp, parseErr := ParseOpencodeJSON(rawOutput)
		if parseErr != nil {
			return rawOutput, nil
		}
		return resp.Content, nil
	}

	return rawOutput, nil
}

// GenerateWithResponse sends a prompt and returns a structured response with metadata.
func (p *OpencodeProvider) GenerateWithResponse(ctx context.Context, prompt, model string) (*Response, error) {
	return p.GenerateWithResponseDir(ctx, prompt, model, "")
}

// GenerateWithResponseDir sends a prompt with working directory and returns structured response.
func (p *OpencodeProvider) GenerateWithResponseDir(ctx context.Context, prompt, model, dir string) (*Response, error) {
	args := []string{"run"}
	args = append(args, "--format", "json")

	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)

	slog.Info("Opencode generating response (structured)", "model", model, "prompt_len", len(prompt), "dir", dir)
	slog.Debug("Opencode prompt", "content", prompt)

	rawOutput, err := p.ExecuteWithDir(ctx, dir, args...)
	if err != nil {
		slog.Error("Opencode generation failed", "error", err)
		return nil, err
	}

	slog.Debug("Opencode raw output", "output", rawOutput)

	resp, parseErr := ParseOpencodeJSON(rawOutput)
	if parseErr != nil {
		return &Response{
			Content:  rawOutput,
			Provider: p.name,
		}, nil
	}

	resp.Provider = p.name
	if model != "" {
		resp.Model = model
	}
	return resp, nil
}
