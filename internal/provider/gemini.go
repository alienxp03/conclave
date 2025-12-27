package provider

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// GeminiProvider implements the Provider interface for Gemini CLI.
type GeminiProvider struct {
	command string
}

// NewGeminiProvider creates a new Gemini provider.
func NewGeminiProvider() *GeminiProvider {
	return &GeminiProvider{
		command: "gemini",
	}
}

// Name returns the provider identifier.
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// DisplayName returns a human-friendly name.
func (p *GeminiProvider) DisplayName() string {
	return "Google Gemini"
}

// Generate sends a prompt to Gemini CLI and returns the response.
func (p *GeminiProvider) Generate(ctx context.Context, prompt string) (string, error) {
	// Use gemini CLI - adjust flags as needed based on actual CLI interface
	cmd := exec.CommandContext(ctx, p.command, prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", &CLIError{
				Provider: p.Name(),
				Message:  stderr.String(),
				Err:      err,
			}
		}
		return "", &CLIError{
			Provider: p.Name(),
			Message:  "failed to execute gemini command",
			Err:      err,
		}
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Available checks if the Gemini CLI is installed.
func (p *GeminiProvider) Available() bool {
	_, err := exec.LookPath(p.command)
	return err == nil
}
