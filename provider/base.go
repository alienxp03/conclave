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
)

const (
	// MaxOutputSize is the maximum size of CLI output (10MB).
	MaxOutputSize = 10 * 1024 * 1024

	// DefaultTimeout is the default timeout for CLI commands.
	DefaultTimeout = 5 * time.Minute
)

// BaseProvider provides common functionality for CLI-based providers.
// Specific providers can embed this to inherit standard CLI execution logic.
type BaseProvider struct {
	name         string
	displayName  string
	command      string
	args         []string
	defaultModel string
	models       []string
	timeout      time.Duration
}

// NewBaseProvider creates a new base provider from configuration.
func NewBaseProvider(cfg Config) BaseProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	displayName := cfg.DisplayName
	if displayName == "" {
		displayName = cfg.Name
	}

	return BaseProvider{
		name:         cfg.Name,
		displayName:  displayName,
		command:      cfg.Command,
		args:         cfg.Args,
		defaultModel: cfg.DefaultModel,
		models:       cfg.Models,
		timeout:      timeout,
	}
}

// Name returns the provider identifier.
func (p *BaseProvider) Name() string {
	return p.name
}

// DisplayName returns the human-friendly name.
func (p *BaseProvider) DisplayName() string {
	return p.displayName
}

// Models returns available models.
func (p *BaseProvider) Models() []string {
	return p.models
}

// DefaultModel returns the default model.
func (p *BaseProvider) DefaultModel() string {
	return p.defaultModel
}

// Timeout returns the configured timeout.
func (p *BaseProvider) Timeout() time.Duration {
	return p.timeout
}

// Available checks if the CLI tool is installed and accessible.
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

// ExecuteCommand runs the CLI command with the given arguments.
func (p *BaseProvider) ExecuteCommand(ctx context.Context, req *Request) (string, error) {
	// Validate executable before running
	if err := p.ValidateExecutable(); err != nil {
		return "", err
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Build command arguments
	allArgs := append([]string{}, p.args...)
	if len(req.Args) > 0 {
		allArgs = append(allArgs, req.Args...)
	}

	slog.Debug("Executing CLI command",
		"provider", p.name,
		"command", p.command,
		"args", allArgs,
		"dir", req.WorkingDir,
	)

	cmd := exec.CommandContext(ctx, p.command, allArgs...)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
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
