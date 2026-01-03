// Package core contains the core domain types for conclave.
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
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Topic       string        `json:"topic"`
	CWD         string        `json:"cwd"`
	WorkspaceID string        `json:"workspace_id,omitempty"` // ID of the workspace (if any)
	AgentA      Agent         `json:"agent_a"`
	AgentB      Agent         `json:"agent_b"`
	Style       string        `json:"style"`
	MaxTurns    int           `json:"max_turns"` // Turns per agent (total = MaxTurns * 2)
	Status      DebateStatus  `json:"status"`
	ReadOnly    bool          `json:"read_only"` // If true, debate cannot be modified or deleted
	Conclusions []*Conclusion `json:"conclusions,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
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
	AgentID   string    `json:"agent_id"` // "user" for user turns
	Number    int       `json:"number"`   // Sequential turn number
	Round     int       `json:"round"`    // Round number
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Vote represents an agent's vote on the conclusion.
type Vote struct {
	AgentID   string `json:"agent_id"`
	Agrees    bool   `json:"agrees"`    // Does the agent agree with the proposed conclusion?
	Reasoning string `json:"reasoning"` // Why they voted this way
}

// Conclusion represents the outcome of a debate round.
type Conclusion struct {
	Round          int    `json:"round"`
	Agreed         bool   `json:"agreed"`
	Summary        string `json:"summary"`
	AgentASummary  string `json:"agent_a_summary,omitempty"`
	AgentBSummary  string `json:"agent_b_summary,omitempty"`
	EarlyConsensus bool   `json:"early_consensus,omitempty"` // True if debate ended early due to consensus
	AgentAVote     *Vote  `json:"agent_a_vote,omitempty"`
	AgentBVote     *Vote  `json:"agent_b_vote,omitempty"`
}

// DebateSummary is a lightweight representation for listing debates.
type DebateSummary struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Topic       string       `json:"topic"`
	CWD         string       `json:"cwd"`
	WorkspaceID string       `json:"workspace_id,omitempty"`
	Status      DebateStatus `json:"status"`
	Style       string       `json:"style"`
	AgentA      string       `json:"agent_a"` // "provider:persona"
	AgentB      string       `json:"agent_b"`
	TurnCount   int          `json:"turn_count"`
	ReadOnly    bool         `json:"read_only"`
	CreatedAt   time.Time    `json:"created_at"`
}

// NewDebateConfig holds the configuration for creating a new debate.
type NewDebateConfig struct {
	Topic          string `json:"topic"`
	WorkspaceID    string `json:"workspace_id"`
	AgentAProvider string `json:"agent_a_provider"`
	AgentAModel    string `json:"agent_a_model"`
	AgentAPersona  string `json:"agent_a_persona"`
	AgentBProvider string `json:"agent_b_provider"`
	AgentBModel    string `json:"agent_b_model"`
	AgentBPersona  string `json:"agent_b_persona"`
	Style          string `json:"style"`
	MaxTurns       int    `json:"max_turns"`
}

// IsModifiable returns true if the debate can be modified.
func (d *Debate) IsModifiable() bool {
	return !d.ReadOnly && d.Status != StatusCompleted
}

// TotalTurns returns the total number of turns (both agents).
func (d *Debate) TotalTurns() int {
	return d.MaxTurns * 2
}

// CouncilSynthesis represents a chairman's synthesis for a specific round.
type CouncilSynthesis struct {
	Round     int       `json:"round"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Council represents a multi-agent council session.
type Council struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Topic       string              `json:"topic"`
	CWD         string              `json:"cwd"`
	WorkspaceID string              `json:"workspace_id,omitempty"`
	Members     []Agent             `json:"members"`
	Chairman    Agent               `json:"chairman"`
	Status      DebateStatus        `json:"status"`
	Syntheses   []*CouncilSynthesis `json:"syntheses,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
}

// Response represents a council member's response in Stage 1.
type Response struct {
	ID        string    `json:"id"`
	CouncilID string    `json:"council_id"`
	MemberID  string    `json:"member_id"`
	Round     int       `json:"round"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Ranking represents a council member's rankings of all responses in Stage 2.
type Ranking struct {
	ID         string    `json:"id"`
	CouncilID  string    `json:"council_id"`
	ReviewerID string    `json:"reviewer_id"`
	Round      int       `json:"round"`
	Rankings   []string  `json:"rankings"` // Ordered list of response IDs (best to worst)
	Reasoning  string    `json:"reasoning,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// CouncilSummary is a lightweight representation for listing councils.
type CouncilSummary struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Topic       string       `json:"topic"`
	CWD         string       `json:"cwd"`
	WorkspaceID string       `json:"workspace_id,omitempty"`
	Status      DebateStatus `json:"status"`
	MemberCount int          `json:"member_count"`
	CreatedAt   time.Time    `json:"created_at"`
}

// NewCouncilConfig holds the configuration for creating a new council.
type NewCouncilConfig struct {
	Topic       string
	WorkspaceID string
	Members     []MemberSpec
	Chairman    *MemberSpec // Optional, defaults to first member's provider with best model
}

// MemberSpec specifies a council member: provider[/model][:persona]
type MemberSpec struct {
	Provider string
	Model    string // Optional, defaults to provider's default
	Persona  string // Optional, auto-assigned if empty
}

// AggregateRanking holds the aggregated ranking data for a response.
type AggregateRanking struct {
	ResponseID string
	MemberID   string
	Positions  []int   // Position in each reviewer's ranking (1-based)
	AvgRank    float64 // Average position across all rankings
}
