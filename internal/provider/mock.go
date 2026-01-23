package provider

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alienxp03/conclave/provider"
)

// MockProvider is a provider that generates simulated responses for testing.
type MockProvider struct {
	name    string
	models  []string
	timeout time.Duration
}

// NewMockProvider creates a new mock provider.
func NewMockProvider(cfg provider.Config) *MockProvider {
	if cfg.Name == "" {
		cfg.Name = "mock"
	}
	if len(cfg.Models) == 0 {
		cfg.Models = []string{"mock-v1", "mock-v2"}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 1 * time.Minute
	}

	return &MockProvider{
		name:    cfg.Name,
		models:  cfg.Models,
		timeout: cfg.Timeout,
	}
}

// Name returns the provider identifier.
func (p *MockProvider) Name() string {
	return p.name
}

// Available always returns true for mock provider.
func (p *MockProvider) Available() bool {
	return true
}

// Execute generates a simulated response.
func (p *MockProvider) Execute(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	slog.Info("Mock provider generating response", "prompt_len", len(req.Prompt), "dir", req.WorkingDir)
	slog.Debug("Mock prompt", "content", req.Prompt)

	// Simulate processing time
	start := time.Now()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}
	duration := time.Since(start)

	// Simple simulated response
	content := fmt.Sprintf("Mock response to: %s... [Simulated content]", truncate(req.Prompt, 50))
	slog.Debug("Mock response", "content", content)

	model := req.Model
	if model == "" && len(p.models) > 0 {
		model = p.models[0]
	}

	return &provider.Response{
		Content:  content,
		Model:    model,
		Provider: p.name,
		Metadata: &provider.Metadata{
			InputTokens:  len(req.Prompt) / 4, // Simulated token count
			OutputTokens: len(content) / 4,
			TotalTokens:  (len(req.Prompt) + len(content)) / 4,
			Duration:     duration,
			StopReason:   "end_turn",
		},
		Raw: content,
	}, nil
}

// HealthCheck always succeeds for mock provider.
func (p *MockProvider) HealthCheck(ctx context.Context) provider.HealthStatus {
	return provider.HealthStatus{
		Available:    true,
		ResponseTime: 0,
		CheckedAt:    time.Now(),
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}
