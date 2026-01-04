package provider

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/alienxp03/conclave/internal/config"
)

// MockProvider is a provider that generates simulated responses for testing.
type MockProvider struct {
	BaseProvider
}

// NewMockProvider creates a new mock provider.
func NewMockProvider(cfg config.ProviderConfig) *MockProvider {
	return &MockProvider{
		BaseProvider: NewBaseProvider("mock", "Mock (Simulated)", cfg),
	}
}

// Generate generates a simulated response.
func (p *MockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	slog.Info("Mock provider generating response", "prompt_len", len(prompt))
	slog.Debug("Mock prompt", "content", prompt)

	// Simulate processing time
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}

	// Simple simulated response
	response := fmt.Sprintf("Mock response to: %s... [Simulated content]", truncate(prompt, 50))
	slog.Debug("Mock response", "content", response)
	return response, nil
}

// GenerateWithModel generates a simulated response with a specific model.
func (p *MockProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	return p.GenerateWithDir(ctx, prompt, model, "")
}

// GenerateWithDir generates a simulated response with a specific model and directory.
func (p *MockProvider) GenerateWithDir(ctx context.Context, prompt, model, dir string) (string, error) {
	slog.Info("Mock provider generating response", "prompt_len", len(prompt), "dir", dir)
	return p.Generate(ctx, prompt)
}

// GenerateWithResponseDir generates a simulated response with metadata.
func (p *MockProvider) GenerateWithResponseDir(ctx context.Context, prompt, model, dir string) (*Response, error) {
	content, err := p.GenerateWithDir(ctx, prompt, model, dir)
	if err != nil {
		return nil, err
	}

	return &Response{
		Content:  content,
		Model:    model,
		Provider: p.name,
		Metadata: &ResponseMeta{
			InputTokens:  len(prompt) / 4, // Simulated token count
			OutputTokens: len(content) / 4,
			TotalTokens:  (len(prompt) + len(content)) / 4,
			DurationMs:   500, // Simulated duration
			StopReason:   "end_turn",
		},
	}, nil
}

// Available always returns true for mock provider.
func (p *MockProvider) Available() bool {
	return true
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

// Helper to register mock in registry from config
func NewMockProviderWithConfig(cfg config.ProviderConfig) *MockProvider {
	if cfg.Command == "" {
		cfg.Command = "mock"
	}
	if len(cfg.Models) == 0 {
		cfg.Models = []string{"mock-v1", "mock-v2"}
	}
	return NewMockProvider(cfg)
}

func init() {
	// Seed random generator
	rand.Seed(time.Now().UnixNano())
}
