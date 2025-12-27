// Package core contains the core domain types for dbate.
package core

import (
	"time"
)

// DebateStatus represents the current status of a debate.
type DebateStatus string

const (
	StatusPending    DebateStatus = "pending"
	StatusInProgress DebateStatus = "in_progress"
	StatusCompleted  DebateStatus = "completed"
	StatusFailed     DebateStatus = "failed"
)

// Debate represents a debate session between two AI agents.
type Debate struct {
	ID          string       `json:"id"`
	Topic       string       `json:"topic"`
	AgentA      Agent        `json:"agent_a"`
	AgentB      Agent        `json:"agent_b"`
	Style       string       `json:"style"`
	MaxTurns    int          `json:"max_turns"`    // Turns per agent (total = MaxTurns * 2)
	Status      DebateStatus `json:"status"`
	Conclusion  *Conclusion  `json:"conclusion,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}

// Agent represents an AI agent participating in a debate.
type Agent struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"` // claude, codex, gemini
	Persona  string `json:"persona"`  // optimist, skeptic, etc.
}

// Turn represents a single turn in the debate.
type Turn struct {
	ID        string    `json:"id"`
	DebateID  string    `json:"debate_id"`
	AgentID   string    `json:"agent_id"`
	Number    int       `json:"number"` // Turn number (1, 2, 3, ...)
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Conclusion represents the outcome of a debate.
type Conclusion struct {
	Agreed        bool   `json:"agreed"`
	Summary       string `json:"summary"`
	AgentASummary string `json:"agent_a_summary,omitempty"`
	AgentBSummary string `json:"agent_b_summary,omitempty"`
}

// DebateSummary is a lightweight representation for listing debates.
type DebateSummary struct {
	ID        string       `json:"id"`
	Topic     string       `json:"topic"`
	Status    DebateStatus `json:"status"`
	Style     string       `json:"style"`
	AgentA    string       `json:"agent_a"` // "provider:persona"
	AgentB    string       `json:"agent_b"`
	TurnCount int          `json:"turn_count"`
	CreatedAt time.Time    `json:"created_at"`
}

// NewDebateConfig holds the configuration for creating a new debate.
type NewDebateConfig struct {
	Topic          string
	AgentAProvider string
	AgentAPersona  string
	AgentBProvider string
	AgentBPersona  string
	Style          string
	MaxTurns       int
}
