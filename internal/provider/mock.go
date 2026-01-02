package provider

import (
	"context"
	"fmt"
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
	// Simulate processing time
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}

	// Simple simulated response
	return fmt.Sprintf("Mock response to: %s... [Simulated content]", truncate(prompt, 50)), nil
}

// GenerateWithModel generates a simulated response with a specific model.
func (p *MockProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	return p.Generate(ctx, prompt)
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
