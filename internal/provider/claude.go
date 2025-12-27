package provider

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// ClaudeProvider implements the Provider interface for Claude CLI.
type ClaudeProvider struct {
	command string
}

// NewClaudeProvider creates a new Claude provider.
func NewClaudeProvider() *ClaudeProvider {
	return &ClaudeProvider{
		command: "claude",
	}
}

// Name returns the provider identifier.
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// DisplayName returns a human-friendly name.
func (p *ClaudeProvider) DisplayName() string {
	return "Claude"
}

// Generate sends a prompt to Claude CLI and returns the response.
func (p *ClaudeProvider) Generate(ctx context.Context, prompt string) (string, error) {
	// Use claude CLI with --print flag for non-interactive mode
	cmd := exec.CommandContext(ctx, p.command, "--print", prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if stderr has useful error info
		if stderr.Len() > 0 {
			return "", &CLIError{
				Provider: p.Name(),
				Message:  stderr.String(),
				Err:      err,
			}
		}
		return "", &CLIError{
			Provider: p.Name(),
			Message:  "failed to execute claude command",
			Err:      err,
		}
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Available checks if the Claude CLI is installed.
func (p *ClaudeProvider) Available() bool {
	_, err := exec.LookPath(p.command)
	return err == nil
}
