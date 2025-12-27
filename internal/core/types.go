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

// DebateMode represents how the debate is executed.
type DebateMode string

const (
	ModeAutomatic  DebateMode = "automatic"   // Run all turns automatically
	ModeTurnByTurn DebateMode = "turn_by_turn" // Execute one turn at a time
)

// Debate represents a debate session between two AI agents.
type Debate struct {
	ID          string       `json:"id"`
	Topic       string       `json:"topic"`
	AgentA      Agent        `json:"agent_a"`
	AgentB      Agent        `json:"agent_b"`
	Style       string       `json:"style"`
	Mode        DebateMode   `json:"mode"`
	MaxTurns    int          `json:"max_turns"`    // Turns per agent (total = MaxTurns * 2)
	Status      DebateStatus `json:"status"`
	ReadOnly    bool         `json:"read_only"`    // If true, debate cannot be modified or deleted
	Conclusion  *Conclusion  `json:"conclusion,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}

// Agent represents an AI agent participating in a debate.
type Agent struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"` // claude, codex, gemini, qwen
	Model    string `json:"model"`    // specific model (optional)
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

// Vote represents an agent's vote on the conclusion.
type Vote struct {
	AgentID   string `json:"agent_id"`
	Agrees    bool   `json:"agrees"`     // Does the agent agree with the proposed conclusion?
	Reasoning string `json:"reasoning"`  // Why they voted this way
}

// Conclusion represents the outcome of a debate.
type Conclusion struct {
	Agreed         bool   `json:"agreed"`
	Summary        string `json:"summary"`
	AgentASummary  string `json:"agent_a_summary,omitempty"`
	AgentBSummary  string `json:"agent_b_summary,omitempty"`
	EarlyConsensus bool   `json:"early_consensus,omitempty"` // True if debate ended early due to consensus
	AgentAVote    *Vote  `json:"agent_a_vote,omitempty"`
	AgentBVote    *Vote  `json:"agent_b_vote,omitempty"`
}

// DebateSummary is a lightweight representation for listing debates.
type DebateSummary struct {
	ID        string       `json:"id"`
	Topic     string       `json:"topic"`
	Status    DebateStatus `json:"status"`
	Style     string       `json:"style"`
	Mode      DebateMode   `json:"mode"`
	AgentA    string       `json:"agent_a"` // "provider:persona"
	AgentB    string       `json:"agent_b"`
	TurnCount int          `json:"turn_count"`
	ReadOnly  bool         `json:"read_only"`
	CreatedAt time.Time    `json:"created_at"`
}

// NewDebateConfig holds the configuration for creating a new debate.
type NewDebateConfig struct {
	Topic          string
	AgentAProvider string
	AgentAModel    string
	AgentAPersona  string
	AgentBProvider string
	AgentBModel    string
	AgentBPersona  string
	Style          string
	Mode           DebateMode
	MaxTurns       int
}

// IsModifiable returns true if the debate can be modified.
func (d *Debate) IsModifiable() bool {
	return !d.ReadOnly && d.Status != StatusCompleted
}

// TotalTurns returns the total number of turns (both agents).
func (d *Debate) TotalTurns() int {
	return d.MaxTurns * 2
}
