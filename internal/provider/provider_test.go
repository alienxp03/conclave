package provider

import (
	"context"
	"testing"
	"time"
)

// MockProvider is a test provider
type MockProvider struct {
	name        string
	displayName string
	available   bool
	response    string
	err         error
}

func (m *MockProvider) Name() string        { return m.name }
func (m *MockProvider) DisplayName() string { return m.displayName }
func (m *MockProvider) Available() bool     { return m.available }
func (m *MockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}
func (m *MockProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	return m.Generate(ctx, prompt)
}
func (m *MockProvider) Models() []string       { return []string{"test-model"} }
func (m *MockProvider) DefaultModel() string   { return "test-model" }
func (m *MockProvider) Timeout() time.Duration { return 2 * time.Minute }

func TestRegistry(t *testing.T) {
	t.Run("RegisterAndGet", func(t *testing.T) {
		r := NewRegistry()

		mock := &MockProvider{
			name:        "mock",
			displayName: "Mock Provider",
			available:   true,
		}

		r.Register(mock)

		got, err := r.Get("mock")
		if err != nil {
			t.Fatalf("failed to get provider: %v", err)
		}

		if got.Name() != "mock" {
			t.Errorf("wrong name: got %s, want mock", got.Name())
		}
	})

	t.Run("GetNonexistent", func(t *testing.T) {
		r := NewRegistry()

		_, err := r.Get("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent provider")
		}
	})

	t.Run("List", func(t *testing.T) {
		r := NewRegistry()
		r.Register(&MockProvider{name: "a", available: true})
		r.Register(&MockProvider{name: "b", available: false})

		list := r.List()
		if len(list) != 2 {
			t.Errorf("wrong count: got %d, want 2", len(list))
		}
	})

	t.Run("Available", func(t *testing.T) {
		r := NewRegistry()
		r.Register(&MockProvider{name: "a", available: true})
		r.Register(&MockProvider{name: "b", available: false})
		r.Register(&MockProvider{name: "c", available: true})

		available := r.Available()
		if len(available) != 2 {
			t.Errorf("wrong count: got %d, want 2", len(available))
		}
	})
}

func TestDefaultRegistry(t *testing.T) {
	r := DefaultRegistry()

	// Should have all 4 providers registered
	providers := r.List()
	if len(providers) != 4 {
		t.Errorf("wrong count: got %d, want 4", len(providers))
	}

	// Check each provider exists
	for _, name := range []string{"claude", "codex", "gemini", "qwen"} {
		_, err := r.Get(name)
		if err != nil {
			t.Errorf("provider %s not found", name)
		}
	}
}

func TestCLIError(t *testing.T) {
	err := &CLIError{
		Provider: "test",
		Message:  "test error",
		Err:      nil,
	}

	expected := "test CLI error: test error"
	if err.Error() != expected {
		t.Errorf("wrong error: got %s, want %s", err.Error(), expected)
	}
}
