package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alienxp03/conclave/internal/core"
	"github.com/alienxp03/conclave/internal/provider"
	"github.com/alienxp03/conclave/internal/storage"
	"github.com/alienxp03/conclave/internal/workspace"
)

// setupTestHandler creates a handler with in-memory storage for testing.
func setupTestHandler(t *testing.T) (*Handler, func()) {
	t.Helper()

	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "conclave-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := tmpDir + "/test.db"
	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Initialize schema
	if err := store.Initialize(); err != nil {
		store.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	registry := provider.NewRegistry()

	// Create workspace manager - it loads from ~/.conclave/workspaces.json by default
	workspaces, err := workspace.NewManager()
	if err != nil {
		store.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	handler := New(store, registry, workspaces)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return handler, cleanup
}

func TestHandleAPIDebate_ReturnsStatsWithMetadata(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create a test debate directly in storage
	debate := &core.Debate{
		ID:     "test-debate-001",
		Title:  "Test Debate",
		Topic:  "Test Topic",
		CWD:    "/tmp/test",
		AgentA: core.Agent{ID: "agent-a", Name: "Agent A", Provider: "mock", Persona: "optimist"},
		AgentB: core.Agent{ID: "agent-b", Name: "Agent B", Provider: "mock", Persona: "skeptic"},
		Style:  "collaborative",
		Status: core.StatusCompleted,
	}
	if err := handler.storage.CreateDebate(debate); err != nil {
		t.Fatalf("Failed to create test debate: %v", err)
	}

	// Create test turns with metadata
	turns := []*core.Turn{
		{
			ID:           "turn-1",
			DebateID:     debate.ID,
			AgentID:      "agent-a",
			Number:       1,
			Round:        1,
			Content:      "First response",
			TurnType:     core.TurnTypeDebate,
			InputTokens:  100,
			OutputTokens: 50,
			TotalTokens:  150,
			DurationMs:   1234,
			Model:        "test-model",
			StopReason:   "end_turn",
		},
		{
			ID:           "turn-2",
			DebateID:     debate.ID,
			AgentID:      "agent-b",
			Number:       2,
			Round:        1,
			Content:      "Second response",
			TurnType:     core.TurnTypeDebate,
			InputTokens:  200,
			OutputTokens: 75,
			TotalTokens:  275,
			DurationMs:   2345,
			Model:        "test-model",
			StopReason:   "end_turn",
		},
	}
	for _, turn := range turns {
		if err := handler.storage.AddTurn(turn); err != nil {
			t.Fatalf("Failed to add turn: %v", err)
		}
	}

	// Create request and response recorder
	req := httptest.NewRequest("GET", "/api/debates/test-debate-001", nil)
	req.SetPathValue("id", "test-debate-001")
	w := httptest.NewRecorder()

	// Call handler
	handler.handleAPIDebate(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Parse response
	var result struct {
		Debate *core.Debate      `json:"debate"`
		Turns  []*core.Turn      `json:"turns"`
		Stats  *core.DebateStats `json:"stats"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify debate
	if result.Debate == nil {
		t.Fatal("Expected debate in response")
	}
	if result.Debate.ID != debate.ID {
		t.Errorf("Expected debate ID %q, got %q", debate.ID, result.Debate.ID)
	}

	// Verify turns
	if len(result.Turns) != 2 {
		t.Errorf("Expected 2 turns, got %d", len(result.Turns))
	}

	// Verify individual turn metadata
	if result.Turns[0].InputTokens != 100 {
		t.Errorf("Turn 1 InputTokens = %d, want 100", result.Turns[0].InputTokens)
	}
	if result.Turns[0].OutputTokens != 50 {
		t.Errorf("Turn 1 OutputTokens = %d, want 50", result.Turns[0].OutputTokens)
	}
	if result.Turns[0].DurationMs != 1234 {
		t.Errorf("Turn 1 DurationMs = %d, want 1234", result.Turns[0].DurationMs)
	}

	// Verify stats
	if result.Stats == nil {
		t.Fatal("Expected stats in response")
	}

	// Check aggregated stats
	expectedTotalInput := 100 + 200  // 300
	expectedTotalOutput := 50 + 75   // 125
	expectedTotalTokens := 150 + 275 // 425

	if result.Stats.TotalInputTokens != expectedTotalInput {
		t.Errorf("Stats TotalInputTokens = %d, want %d", result.Stats.TotalInputTokens, expectedTotalInput)
	}
	if result.Stats.TotalOutputTokens != expectedTotalOutput {
		t.Errorf("Stats TotalOutputTokens = %d, want %d", result.Stats.TotalOutputTokens, expectedTotalOutput)
	}
	if result.Stats.TotalTokens != expectedTotalTokens {
		t.Errorf("Stats TotalTokens = %d, want %d", result.Stats.TotalTokens, expectedTotalTokens)
	}
	if result.Stats.TurnCount != 2 {
		t.Errorf("Stats TurnCount = %d, want 2", result.Stats.TurnCount)
	}

	// Check per-agent stats
	if result.Stats.AgentATurnCount != 1 {
		t.Errorf("Stats AgentATurnCount = %d, want 1", result.Stats.AgentATurnCount)
	}
	if result.Stats.AgentBTurnCount != 1 {
		t.Errorf("Stats AgentBTurnCount = %d, want 1", result.Stats.AgentBTurnCount)
	}
}

func TestHandleAPIProject_UpdateReinjectsInstructions(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create project
	project := &core.Project{
		ID:           "project-001",
		Name:         "Alpha",
		Description:  "Test project",
		Instructions: "Initial instructions",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := handler.storage.CreateProject(project); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create debate and council tied to project
	debate := &core.Debate{
		ID:                  "debate-001",
		Title:               "Project Debate",
		Topic:               "Project topic",
		CWD:                 "/tmp/test",
		ProjectID:           project.ID,
		ProjectInstructions: project.Instructions,
		AgentA:              core.Agent{ID: "agent-a", Name: "Agent A", Provider: "mock", Persona: "optimist"},
		AgentB:              core.Agent{ID: "agent-b", Name: "Agent B", Provider: "mock", Persona: "skeptic"},
		Style:               "collaborative",
		MaxTurns:            3,
		Status:              core.StatusPending,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	if err := handler.storage.CreateDebate(debate); err != nil {
		t.Fatalf("Failed to create debate: %v", err)
	}

	council := &core.Council{
		ID:                  "council-001",
		Title:               "Project Council",
		Topic:               "Council topic",
		CWD:                 "/tmp/test",
		ProjectID:           project.ID,
		ProjectInstructions: project.Instructions,
		Chairman:            core.Agent{ID: "chairman", Name: "Chairman", Provider: "mock", Persona: "chairman"},
		Status:              core.StatusPending,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	if err := handler.storage.CreateCouncil(council); err != nil {
		t.Fatalf("Failed to create council: %v", err)
	}

	// Update project instructions via API
	payload := `{"name":"Alpha","description":"Test project","instructions":"Updated instructions"}`
	req := httptest.NewRequest("PUT", "/api/projects/project-001", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "project-001")
	w := httptest.NewRecorder()

	handler.handleAPIUpdateProject(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	updatedDebate, err := handler.storage.GetDebate(debate.ID)
	if err != nil {
		t.Fatalf("Failed to load debate: %v", err)
	}
	if updatedDebate.ProjectInstructions != "Updated instructions" {
		t.Errorf("Debate project instructions = %q, want %q", updatedDebate.ProjectInstructions, "Updated instructions")
	}

	updatedCouncil, err := handler.storage.GetCouncil(council.ID)
	if err != nil {
		t.Fatalf("Failed to load council: %v", err)
	}
	if updatedCouncil.ProjectInstructions != "Updated instructions" {
		t.Errorf("Council project instructions = %q, want %q", updatedCouncil.ProjectInstructions, "Updated instructions")
	}
}
