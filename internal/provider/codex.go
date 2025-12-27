package provider

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// CodexProvider implements the Provider interface for OpenAI Codex CLI.
type CodexProvider struct {
	command string
}

// NewCodexProvider creates a new Codex provider.
func NewCodexProvider() *CodexProvider {
	return &CodexProvider{
		command: "codex",
	}
}

// Name returns the provider identifier.
func (p *CodexProvider) Name() string {
	return "codex"
}

// DisplayName returns a human-friendly name.
func (p *CodexProvider) DisplayName() string {
	return "OpenAI Codex"
}

// Generate sends a prompt to Codex CLI and returns the response.
func (p *CodexProvider) Generate(ctx context.Context, prompt string) (string, error) {
	// Use codex CLI - adjust flags as needed based on actual CLI interface
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
			Message:  "failed to execute codex command",
			Err:      err,
		}
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Available checks if the Codex CLI is installed.
func (p *CodexProvider) Available() bool {
	_, err := exec.LookPath(p.command)
	return err == nil
}
