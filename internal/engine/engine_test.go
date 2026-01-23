package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alienxp03/conclave/internal/core"
	"github.com/alienxp03/conclave/internal/provider"
	"github.com/alienxp03/conclave/internal/storage"
	extprovider "github.com/alienxp03/conclave/provider"
)

// MockProvider for testing - implements extprovider.Provider interface
type MockProvider struct {
	name      string
	available bool
	responses []string
	callCount int
}

func (m *MockProvider) Name() string    { return m.name }
func (m *MockProvider) Available() bool { return m.available }

// Execute implements the provider.Provider interface
func (m *MockProvider) Execute(ctx context.Context, req *extprovider.Request) (*extprovider.Response, error) {
	idx := m.callCount % len(m.responses)
	m.callCount++
	return &extprovider.Response{
		Content:  m.responses[idx],
		Model:    "test-model",
		Provider: m.name,
	}, nil
}

func (m *MockProvider) HealthCheck(ctx context.Context) extprovider.HealthStatus {
	return extprovider.HealthStatus{
		Available:    m.available,
		ResponseTime: 0,
		CheckedAt:    time.Now(),
	}
}

// Legacy methods for backward compatibility with internal/provider.Provider
func (m *MockProvider) DisplayName() string { return m.name }
func (m *MockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	resp, err := m.Execute(ctx, &extprovider.Request{Prompt: prompt})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
func (m *MockProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	return m.Generate(ctx, prompt)
}
func (m *MockProvider) GenerateWithDir(ctx context.Context, prompt, model, dir string) (string, error) {
	return m.Generate(ctx, prompt)
}
func (m *MockProvider) GenerateWithResponseDir(ctx context.Context, prompt, model, dir string) (*extprovider.Response, error) {
	return m.Execute(ctx, &extprovider.Request{Prompt: prompt, Model: model, WorkingDir: dir})
}
func (m *MockProvider) Models() []string       { return []string{"test-model"} }
func (m *MockProvider) DefaultModel() string   { return "test-model" }
func (m *MockProvider) Timeout() time.Duration { return 2 * time.Minute }

func setupTestEngine(t *testing.T) (*Engine, func()) {
	tmpDir, err := os.MkdirTemp("", "conclave-engine-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	store, err := storage.NewSQLiteStorage(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	if err := store.Initialize(); err != nil {
		store.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to initialize storage: %v", err)
	}

	registry := provider.NewRegistry()
	registry.Register(&MockProvider{
		name:      "mock",
		available: true,
		responses: []string{"Response 1", "Response 2", "CONSENSUS: yes\nSUMMARY: Both agreed"},
	})

	eng := New(store, registry, nil)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return eng, cleanup
}

func TestCreateDebate(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("ValidConfig", func(t *testing.T) {
		config := core.NewDebateConfig{
			Topic:          "Test topic",
			AgentAProvider: "mock",
			AgentAPersona:  "optimist",
			AgentBProvider: "mock",
			AgentBPersona:  "skeptic",
			Style:          "collaborative",
			MaxTurns:       3,
		}

		debate, err := eng.CreateDebate(ctx, config)
		if err != nil {
			t.Fatalf("failed to create debate: %v", err)
		}

		if debate.ID == "" {
			t.Error("debate ID is empty")
		}
		if debate.Topic != "Test topic" {
			t.Errorf("wrong topic: got %s", debate.Topic)
		}
		if debate.Status != core.StatusPending {
			t.Errorf("wrong status: got %s", debate.Status)
		}
		if debate.MaxTurns != 3 {
			t.Errorf("wrong max turns: got %d", debate.MaxTurns)
		}
	})

	t.Run("InvalidProvider", func(t *testing.T) {
		config := core.NewDebateConfig{
			Topic:          "Test",
			AgentAProvider: "nonexistent",
			AgentAPersona:  "optimist",
			AgentBProvider: "mock",
			AgentBPersona:  "skeptic",
			Style:          "collaborative",
		}

		_, err := eng.CreateDebate(ctx, config)
		if err == nil {
			t.Error("expected error for invalid provider")
		}
	})

	t.Run("InvalidPersona", func(t *testing.T) {
		config := core.NewDebateConfig{
			Topic:          "Test",
			AgentAProvider: "mock",
			AgentAPersona:  "nonexistent",
			AgentBProvider: "mock",
			AgentBPersona:  "skeptic",
			Style:          "collaborative",
		}

		_, err := eng.CreateDebate(ctx, config)
		if err == nil {
			t.Error("expected error for invalid persona")
		}
	})

	t.Run("InvalidStyle", func(t *testing.T) {
		config := core.NewDebateConfig{
			Topic:          "Test",
			AgentAProvider: "mock",
			AgentAPersona:  "optimist",
			AgentBProvider: "mock",
			AgentBPersona:  "skeptic",
			Style:          "nonexistent",
		}

		_, err := eng.CreateDebate(ctx, config)
		if err == nil {
			t.Error("expected error for invalid style")
		}
	})

	t.Run("DefaultMaxTurns", func(t *testing.T) {
		config := core.NewDebateConfig{
			Topic:          "Test",
			AgentAProvider: "mock",
			AgentAPersona:  "optimist",
			AgentBProvider: "mock",
			AgentBPersona:  "skeptic",
			Style:          "collaborative",
			MaxTurns:       0, // Should default to 5
		}

		debate, err := eng.CreateDebate(ctx, config)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}

		if debate.MaxTurns != 5 {
			t.Errorf("default max turns wrong: got %d, want 5", debate.MaxTurns)
		}
	})
}

func TestGetDebate(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create a debate first
	config := core.NewDebateConfig{
		Topic:          "Test",
		AgentAProvider: "mock",
		AgentAPersona:  "optimist",
		AgentBProvider: "mock",
		AgentBPersona:  "skeptic",
		Style:          "collaborative",
	}
	created, _ := eng.CreateDebate(ctx, config)

	t.Run("ExistingDebate", func(t *testing.T) {
		debate, err := eng.GetDebate(created.ID)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if debate == nil {
			t.Fatal("debate not found")
		}
		if debate.ID != created.ID {
			t.Errorf("wrong ID: got %s", debate.ID)
		}
	})

	t.Run("NonexistentDebate", func(t *testing.T) {
		debate, err := eng.GetDebate("nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if debate != nil {
			t.Error("expected nil for nonexistent debate")
		}
	})
}

func TestListDebates(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple debates
	for i := 0; i < 3; i++ {
		config := core.NewDebateConfig{
			Topic:          "Test",
			AgentAProvider: "mock",
			AgentAPersona:  "optimist",
			AgentBProvider: "mock",
			AgentBPersona:  "skeptic",
			Style:          "collaborative",
		}
		eng.CreateDebate(ctx, config)
	}

	debates, err := eng.ListDebates(10, 0)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if len(debates) != 3 {
		t.Errorf("wrong count: got %d, want 3", len(debates))
	}
}

func TestDeleteDebate(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	config := core.NewDebateConfig{
		Topic:          "Test",
		AgentAProvider: "mock",
		AgentAPersona:  "optimist",
		AgentBProvider: "mock",
		AgentBPersona:  "skeptic",
		Style:          "collaborative",
	}
	debate, _ := eng.CreateDebate(ctx, config)

	if err := eng.DeleteDebate(debate.ID); err != nil {
		t.Fatalf("failed: %v", err)
	}

	got, _ := eng.GetDebate(debate.ID)
	if got != nil {
		t.Error("debate still exists after deletion")
	}
}

func TestRunDebate(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	config := core.NewDebateConfig{
		Topic:          "Test",
		AgentAProvider: "mock",
		AgentAPersona:  "optimist",
		AgentBProvider: "mock",
		AgentBPersona:  "skeptic",
		Style:          "collaborative",
		MaxTurns:       2, // 4 debate turns total (2 per agent)
	}
	debate, _ := eng.CreateDebate(ctx, config)

	turnCount := 0
	callback := func(turn *core.Turn, d *core.Debate) {
		turnCount++
	}

	if err := eng.RunDebate(ctx, debate.ID, callback); err != nil {
		t.Fatalf("failed: %v", err)
	}

	// Callback is only called for debate turns (4)
	if turnCount != 4 {
		t.Errorf("wrong callback count: got %d, want 4", turnCount)
	}

	// Check final state
	final, turns, _ := eng.GetDebateWithTurns(debate.ID)
	if final.Status != core.StatusCompleted {
		t.Errorf("wrong status: got %s, want completed", final.Status)
	}

	// Total stored turns includes debate (4) + vote (2) + conclusion (1) = 7
	if len(turns) != 7 {
		t.Errorf("wrong stored turn count: got %d, want 7", len(turns))
	}

	// Verify debate turn count
	debateTurnCount := 0
	for _, turn := range turns {
		if turn.TurnType == core.TurnTypeDebate {
			debateTurnCount++
		}
	}
	if debateTurnCount != 4 {
		t.Errorf("wrong debate turn count: got %d, want 4", debateTurnCount)
	}

	if len(final.Conclusions) == 0 {
		t.Error("conclusions is empty")
	}
}

func TestExecuteNextTurn(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	config := core.NewDebateConfig{
		Topic:          "Test",
		AgentAProvider: "mock",
		AgentAPersona:  "optimist",
		AgentBProvider: "mock",
		AgentBPersona:  "skeptic",
		Style:          "collaborative",
		MaxTurns:       2,
	}
	debate, _ := eng.CreateDebate(ctx, config)

	// Execute first turn
	turn1, err := eng.ExecuteNextTurn(ctx, debate.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if turn1.Number != 1 {
		t.Errorf("wrong turn number: got %d, want 1", turn1.Number)
	}

	// Check status updated
	d, _, _ := eng.GetDebateWithTurns(debate.ID)
	if d.Status != core.StatusInProgress {
		t.Errorf("wrong status: got %s, want in_progress", d.Status)
	}

	// Execute remaining turns
	for i := 0; i < 3; i++ {
		eng.ExecuteNextTurn(ctx, debate.ID)
	}

	// Should be completed now
	final, _, _ := eng.GetDebateWithTurns(debate.ID)
	if final.Status != core.StatusCompleted {
		t.Errorf("wrong final status: got %s, want completed", final.Status)
	}
}

func TestHashString(t *testing.T) {
	// Same string should produce same hash
	h1 := hashString("test")
	h2 := hashString("test")
	if h1 != h2 {
		t.Error("same string produced different hashes")
	}

	// Different strings should produce different hashes
	h3 := hashString("different")
	if h1 == h3 {
		t.Error("different strings produced same hash")
	}
}
