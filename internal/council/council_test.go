package council

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alienxp03/conclave/internal/core"
	"github.com/alienxp03/conclave/internal/provider"
	"github.com/alienxp03/conclave/internal/storage"
	extprovider "github.com/alienxp03/conclave/provider"
)

type mockProvider struct {
	name      string
	available bool
	err       error
}

func (m *mockProvider) Name() string    { return m.name }
func (m *mockProvider) Available() bool { return m.available }

func (m *mockProvider) Execute(ctx context.Context, req *extprovider.Request) (*extprovider.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	content := "OK"
	prompt := req.Prompt
	switch {
	case strings.Contains(prompt, "Summarize the following question"):
		content = "Test Summary"
	case strings.Contains(prompt, "You are the Chairman synthesizing"):
		content = "Synthesis content"
	case strings.Contains(prompt, "FINAL RANKING"):
		content = "FINAL RANKING:\n1. " + m.name
	case strings.Contains(prompt, "Your response:"):
		content = "Response from " + m.name
	}

	return &extprovider.Response{
		Content:  content,
		Model:    "test-model",
		Provider: m.name,
	}, nil
}

func (m *mockProvider) HealthCheck(ctx context.Context) extprovider.HealthStatus {
	return extprovider.HealthStatus{
		Available:    m.available,
		ResponseTime: 0,
		CheckedAt:    time.Now(),
	}
}

func setupTestCouncilEngine(t *testing.T, registry *provider.Registry) (*Engine, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "conclave-council-test-*")
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

	eng := New(store, registry)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return eng, cleanup
}

func TestRunCouncilContinuesOnMemberFailure(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&mockProvider{name: "okprov", available: true})
	registry.Register(&mockProvider{name: "failprov", available: true, err: errors.New("boom")})

	eng, cleanup := setupTestCouncilEngine(t, registry)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := eng.CreateCouncil(ctx, core.NewCouncilConfig{
		Topic: "Test topic",
		Members: []core.MemberSpec{
			{Provider: "okprov"},
			{Provider: "failprov"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create council: %v", err)
	}

	if err := eng.RunCouncil(ctx, c); err != nil {
		t.Fatalf("expected council to complete, got error: %v", err)
	}

	stored, err := eng.storage.GetCouncil(c.ID)
	if err != nil {
		t.Fatalf("failed to reload council: %v", err)
	}
	if stored.Status != core.StatusCompleted {
		t.Fatalf("expected status completed, got %s", stored.Status)
	}

	responses, err := eng.storage.GetResponses(c.ID)
	if err != nil {
		t.Fatalf("failed to load responses: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}

	rankings, err := eng.storage.GetRankings(c.ID)
	if err != nil {
		t.Fatalf("failed to load rankings: %v", err)
	}
	if len(rankings) != 1 {
		t.Fatalf("expected 1 ranking, got %d", len(rankings))
	}

	stored, err = eng.storage.GetCouncil(c.ID)
	if err != nil {
		t.Fatalf("failed to reload council: %v", err)
	}
	if len(stored.Syntheses) != 1 {
		t.Fatalf("expected 1 synthesis, got %d", len(stored.Syntheses))
	}
}

func TestRunCouncilFallsBackWhenChairmanFails(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&mockProvider{name: "okprov", available: true})
	registry.Register(&mockProvider{name: "failprov", available: true, err: errors.New("boom")})

	eng, cleanup := setupTestCouncilEngine(t, registry)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chairman := core.MemberSpec{Provider: "failprov", Model: "test-model", Persona: "chairman"}
	c, err := eng.CreateCouncil(ctx, core.NewCouncilConfig{
		Topic: "Test topic",
		Members: []core.MemberSpec{
			{Provider: "okprov"},
			{Provider: "okprov"},
		},
		Chairman: &chairman,
	})
	if err != nil {
		t.Fatalf("failed to create council: %v", err)
	}

	if err := eng.RunCouncil(ctx, c); err != nil {
		t.Fatalf("expected council to complete, got error: %v", err)
	}

	stored, err := eng.storage.GetCouncil(c.ID)
	if err != nil {
		t.Fatalf("failed to reload council: %v", err)
	}
	if stored.Status != core.StatusCompleted {
		t.Fatalf("expected status completed, got %s", stored.Status)
	}
	if len(stored.Syntheses) != 1 {
		t.Fatalf("expected 1 synthesis, got %d", len(stored.Syntheses))
	}
	if !strings.Contains(stored.Syntheses[0].Content, "Synthesis unavailable") {
		t.Fatalf("expected fallback synthesis content, got: %s", stored.Syntheses[0].Content)
	}
}
