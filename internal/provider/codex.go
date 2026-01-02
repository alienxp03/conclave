package provider

import (
	"context"
	"os"
	"time"

	"github.com/alienxp03/conclave/internal/config"
	"github.com/alienxp03/conclave/internal/core"
)

// CodexResponseSchema is the JSON schema for Codex structured output
const CodexResponseSchema = `{
  "type": "object",
  "properties": {
    "response": { "type": "string" }
  },
  "required": ["response"],
  "additionalProperties": false
}`

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
	args := []string{"exec"}

	// Use structured JSON output with schema
	if p.useJSON {
		schemaPath, err := p.createSchemaFile()
		if err != nil {
			return "", err
		}
		defer os.Remove(schemaPath)
		args = append(args, "--output-schema", schemaPath)
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
	args := []string{"exec"}

	schemaPath, err := p.createSchemaFile()
	if err != nil {
		return nil, err
	}
	defer os.Remove(schemaPath)

	args = append(args, "--output-schema", schemaPath)

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

// createSchemaFile writes the response schema to a temporary file
func (p *CodexProvider) createSchemaFile() (string, error) {
	tmpDir := os.TempDir()

	f, err := os.CreateTemp(tmpDir, "codex-schema-")
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.WriteString(CodexResponseSchema)
	if err != nil {
		os.Remove(f.Name())
		return "", err
	}

	return f.Name(), nil
}
