// Package provider provides a reusable abstraction for AI CLI tools.
//
// This package wraps command-line AI tools (Claude, Gemini, OpenAI, etc.)
// with a unified interface, making it easy to integrate multiple AI providers
// into your applications.
package provider

import (
	"context"
	"time"
)

// Provider defines the interface for AI CLI providers.
type Provider interface {
	// Name returns the provider's unique identifier (e.g., "claude", "gemini").
	Name() string

	// Available checks if the provider's CLI tool is installed and accessible.
	Available() bool

	// Execute sends a request to the provider and returns a structured response.
	Execute(ctx context.Context, req *Request) (*Response, error)
}

// Request represents a generation request to an AI provider.
type Request struct {
	// Prompt is the input text to send to the AI.
	Prompt string

	// Model is the specific model to use (e.g., "sonnet", "gpt-4").
	// If empty, the provider's default model will be used.
	Model string

	// WorkingDir is the directory to execute the CLI command in.
	// Some providers support context-aware operations based on the working directory.
	WorkingDir string

	// Args are additional command-line arguments to pass to the provider.
	Args []string
}

// Response represents a provider's response with metadata.
type Response struct {
	// Content is the AI-generated text response.
	Content string `json:"content"`

	// Model is the model that was used for this response.
	Model string `json:"model,omitempty"`

	// Provider is the name of the provider that generated this response.
	Provider string `json:"provider,omitempty"`

	// Metadata contains usage statistics and additional information.
	Metadata *Metadata `json:"metadata,omitempty"`

	// Raw is the unprocessed output from the CLI tool (for debugging).
	Raw string `json:"-"`
}

// Metadata contains usage statistics and additional response information.
type Metadata struct {
	// InputTokens is the number of tokens in the input/prompt.
	InputTokens int `json:"input_tokens,omitempty"`

	// OutputTokens is the number of tokens in the generated response.
	OutputTokens int `json:"output_tokens,omitempty"`

	// TotalTokens is the total number of tokens (input + output).
	TotalTokens int `json:"total_tokens,omitempty"`

	// Duration is the time taken to generate the response.
	Duration time.Duration `json:"duration,omitempty"`

	// StopReason indicates why the generation stopped (e.g., "end_turn", "max_tokens").
	StopReason string `json:"stop_reason,omitempty"`

	// SessionID is a unique identifier for this session (if supported by the provider).
	SessionID string `json:"session_id,omitempty"`
}

// Config holds configuration for creating a provider.
type Config struct {
	// Name is the unique identifier for this provider (e.g., "claude").
	Name string

	// DisplayName is a human-friendly name for this provider.
	// If empty, Name will be used.
	DisplayName string

	// Command is the CLI executable name (e.g., "claude", "gemini").
	Command string

	// Args are default arguments to pass to the CLI command.
	Args []string

	// DefaultModel is the model to use when Request.Model is empty.
	DefaultModel string

	// Models is a list of available models for this provider.
	Models []string

	// Timeout is the maximum duration for a request.
	// Default: 5 minutes.
	Timeout time.Duration
}
