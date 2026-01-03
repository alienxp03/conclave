package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/alienxp03/conclave/internal/config"
)

const (
	// MaxOutputSize is the maximum size of CLI output (10MB).
	MaxOutputSize = 10 * 1024 * 1024
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

// ValidateExecutable checks if the CLI is available before execution.
// Returns an error if the executable is not found.
func (p *BaseProvider) ValidateExecutable() error {
	path, err := exec.LookPath(p.command)
	if err != nil {
		return &CLIError{
			Provider: p.name,
			Message:  fmt.Sprintf("executable '%s' not found in PATH", p.command),
			Err:      err,
		}
	}
	// Verify it's actually executable
	if path == "" {
		return &CLIError{
			Provider: p.name,
			Message:  fmt.Sprintf("executable '%s' found but path is empty", p.command),
		}
	}
	return nil
}

// limitedWriter wraps an io.Writer and limits total bytes written.
type limitedWriter struct {
	w       io.Writer
	n       int64
	limit   int64
	limited bool
}

func newLimitedWriter(w io.Writer, limit int64) *limitedWriter {
	return &limitedWriter{w: w, limit: limit}
}

func (l *limitedWriter) Write(p []byte) (n int, err error) {
	if l.n >= l.limit {
		l.limited = true
		return len(p), nil // Discard, but don't error
	}

	remaining := l.limit - l.n
	if int64(len(p)) > remaining {
		p = p[:remaining]
		l.limited = true
	}

	n, err = l.w.Write(p)
	l.n += int64(n)
	return n, err
}

// GenerateWithDir sends a prompt with a specific model and working directory.
func (p *BaseProvider) GenerateWithDir(ctx context.Context, prompt, model, dir string) (string, error) {
	// Base implementation doesn't know about specific flags, so it just calls GenerateWithModel
	// Subclasses should override this if they need to pass the directory differently
	// However, since BaseProvider doesn't implement GenerateWithModel (it's in the specific providers),
	// this is a bit tricky.
	// Actually, BaseProvider implements the common Execute method.
	// Specific providers call Execute.
	return "", fmt.Errorf("GenerateWithDir not implemented for %s", p.name)
}

// Execute runs the CLI command with the given arguments.
func (p *BaseProvider) Execute(ctx context.Context, extraArgs ...string) (string, error) {
	return p.ExecuteWithDir(ctx, "", extraArgs...)
}

// ExecuteWithDir runs the CLI command with the given arguments in a specific directory.
func (p *BaseProvider) ExecuteWithDir(ctx context.Context, dir string, extraArgs ...string) (string, error) {
	// Validate executable before running
	if err := p.ValidateExecutable(); err != nil {
		return "", err
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	allArgs := append(p.args, extraArgs...)
	slog.Debug("Executing CLI command",
		"provider", p.name,
		"command", p.command,
		"args", allArgs,
		"dir", dir,
	)
	cmd := exec.CommandContext(ctx, p.command, allArgs...)
	if dir != "" {
		cmd.Dir = dir
	}

	// Use size-limited writers to prevent memory issues
	var stdout, stderr bytes.Buffer
	stdoutLimited := newLimitedWriter(&stdout, MaxOutputSize)
	stderrLimited := newLimitedWriter(&stderr, MaxOutputSize)

	cmd.Stdout = stdoutLimited
	cmd.Stderr = stderrLimited

	if err := cmd.Run(); err != nil {
		slog.Error("CLI command failed",
			"provider", p.name,
			"error", err,
			"stderr", stderr.String(),
		)
		if ctx.Err() == context.DeadlineExceeded {
			return "", &CLIError{
				Provider: p.name,
				Message:  "command timed out",
				Err:      ctx.Err(),
			}
		}
		if stderr.Len() > 0 {
			errMsg := stderr.String()
			if stderrLimited.limited {
				errMsg = errMsg + "\n... (output truncated)"
			}
			return "", &CLIError{
				Provider: p.name,
				Message:  errMsg,
				Err:      err,
			}
		}
		return "", &CLIError{
			Provider: p.name,
			Message:  "command failed",
			Err:      err,
		}
	}

	result := strings.TrimSpace(stdout.String())
	slog.Debug("CLI command successful",
		"provider", p.name,
		"output_len", len(result),
	)
	if stdoutLimited.limited {
		result = result + "\n... (output truncated at 10MB)"
	}

	return result, nil
}
