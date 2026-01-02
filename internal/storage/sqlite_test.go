package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alienxp03/conclave/internal/core"
)

func TestSQLiteStorage(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "conclave-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Create storage
	store, err := NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	// Initialize schema
	if err := store.Initialize(); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	t.Run("CreateAndGetDebate", func(t *testing.T) {
		now := time.Now()
		debate := &core.Debate{
			ID:    "test-debate-1",
			Topic: "Test Topic",
			AgentA: core.Agent{
				ID:       "agent-a-1",
				Name:     "Agent A",
				Provider: "claude",
				Persona:  "optimist",
			},
			AgentB: core.Agent{
				ID:       "agent-b-1",
				Name:     "Agent B",
				Provider: "gemini",
				Persona:  "skeptic",
			},
			Style:     "collaborative",
			MaxTurns:  5,
			Status:    core.StatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := store.CreateDebate(debate); err != nil {
			t.Fatalf("failed to create debate: %v", err)
		}

		// Get debate
		got, err := store.GetDebate(debate.ID)
		if err != nil {
			t.Fatalf("failed to get debate: %v", err)
		}

		if got == nil {
			t.Fatal("debate not found")
		}

		if got.ID != debate.ID {
			t.Errorf("ID mismatch: got %s, want %s", got.ID, debate.ID)
		}
		if got.Topic != debate.Topic {
			t.Errorf("Topic mismatch: got %s, want %s", got.Topic, debate.Topic)
		}
		if got.AgentA.Provider != debate.AgentA.Provider {
			t.Errorf("AgentA.Provider mismatch: got %s, want %s", got.AgentA.Provider, debate.AgentA.Provider)
		}
	})

	t.Run("UpdateDebate", func(t *testing.T) {
		debate, _ := store.GetDebate("test-debate-1")
		debate.Status = core.StatusInProgress

		if err := store.UpdateDebate(debate); err != nil {
			t.Fatalf("failed to update debate: %v", err)
		}

		got, _ := store.GetDebate(debate.ID)
		if got.Status != core.StatusInProgress {
			t.Errorf("Status not updated: got %s, want %s", got.Status, core.StatusInProgress)
		}
	})

	t.Run("AddAndGetTurns", func(t *testing.T) {
		turn1 := &core.Turn{
			ID:        "turn-1",
			DebateID:  "test-debate-1",
			AgentID:   "agent-a-1",
			Number:    1,
			Content:   "First argument",
			CreatedAt: time.Now(),
		}

		turn2 := &core.Turn{
			ID:        "turn-2",
			DebateID:  "test-debate-1",
			AgentID:   "agent-b-1",
			Number:    2,
			Content:   "Second argument",
			CreatedAt: time.Now(),
		}

		if err := store.AddTurn(turn1); err != nil {
			t.Fatalf("failed to add turn 1: %v", err)
		}
		if err := store.AddTurn(turn2); err != nil {
			t.Fatalf("failed to add turn 2: %v", err)
		}

		turns, err := store.GetTurns("test-debate-1")
		if err != nil {
			t.Fatalf("failed to get turns: %v", err)
		}

		if len(turns) != 2 {
			t.Errorf("wrong number of turns: got %d, want 2", len(turns))
		}

		if turns[0].Number != 1 || turns[1].Number != 2 {
			t.Error("turns not in correct order")
		}
	})

	t.Run("GetLatestTurn", func(t *testing.T) {
		turn, err := store.GetLatestTurn("test-debate-1")
		if err != nil {
			t.Fatalf("failed to get latest turn: %v", err)
		}

		if turn.Number != 2 {
			t.Errorf("wrong turn number: got %d, want 2", turn.Number)
		}
	})

	t.Run("ListDebates", func(t *testing.T) {
		summaries, err := store.ListDebates(10, 0)
		if err != nil {
			t.Fatalf("failed to list debates: %v", err)
		}

		if len(summaries) != 1 {
			t.Errorf("wrong number of debates: got %d, want 1", len(summaries))
		}

		if summaries[0].TurnCount != 2 {
			t.Errorf("wrong turn count: got %d, want 2", summaries[0].TurnCount)
		}
	})

	t.Run("DeleteDebate", func(t *testing.T) {
		if err := store.DeleteDebate("test-debate-1"); err != nil {
			t.Fatalf("failed to delete debate: %v", err)
		}

		got, _ := store.GetDebate("test-debate-1")
		if got != nil {
			t.Error("debate still exists after deletion")
		}

		// Turns should also be deleted (cascade)
		turns, _ := store.GetTurns("test-debate-1")
		if len(turns) != 0 {
			t.Error("turns still exist after debate deletion")
		}
	})

	t.Run("GetNonexistentDebate", func(t *testing.T) {
		got, err := store.GetDebate("nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Error("expected nil for nonexistent debate")
		}
	})
}
