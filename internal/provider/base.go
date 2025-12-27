package provider

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/alienxp03/dbate/internal/config"
)

// BaseProvider provides common functionality for CLI-based providers.
type BaseProvider struct {
	name         string
	displayName  string
	command      string
	args         []string
	defaultModel string
	models       []string
	timeout      time.Duration
}

// NewBaseProvider creates a new base provider from config.
func NewBaseProvider(name, displayName string, cfg config.ProviderConfig) BaseProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	return BaseProvider{
		name:         name,
		displayName:  displayName,
		command:      cfg.Command,
		args:         cfg.Args,
		defaultModel: cfg.DefaultModel,
		models:       cfg.Models,
		timeout:      timeout,
	}
}

// Name returns the provider identifier.
func (p *BaseProvider) Name() string { return p.name }

// DisplayName returns the human-friendly name.
func (p *BaseProvider) DisplayName() string { return p.displayName }

// Models returns available models.
func (p *BaseProvider) Models() []string { return p.models }

// DefaultModel returns the default model.
func (p *BaseProvider) DefaultModel() string { return p.defaultModel }

// Timeout returns the configured timeout.
func (p *BaseProvider) Timeout() time.Duration { return p.timeout }

// Available checks if the CLI is installed.
func (p *BaseProvider) Available() bool {
	_, err := exec.LookPath(p.command)
	return err == nil
}

// Execute runs the CLI command with the given arguments.
func (p *BaseProvider) Execute(ctx context.Context, extraArgs ...string) (string, error) {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	allArgs := append(p.args, extraArgs...)
	cmd := exec.CommandContext(ctx, p.command, allArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", &CLIError{
				Provider: p.name,
				Message:  "command timed out",
				Err:      ctx.Err(),
			}
		}
		if stderr.Len() > 0 {
			return "", &CLIError{
				Provider: p.name,
				Message:  stderr.String(),
				Err:      err,
			}
		}
		return "", &CLIError{
			Provider: p.name,
			Message:  "command failed",
			Err:      err,
		}
	}

	return strings.TrimSpace(stdout.String()), nil
}
