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

// TurnType represents the type of turn for separate tracking.
type TurnType string

const (
	TurnTypeDebate     TurnType = "debate"     // Regular debate turn
	TurnTypeConclusion TurnType = "conclusion" // Conclusion generation
	TurnTypeVote       TurnType = "vote"       // Voting turn
	TurnTypeUser       TurnType = "user"       // User input (follow-up)
)

// Turn represents a single turn in the debate.
type Turn struct {
	ID        string    `json:"id"`
	DebateID  string    `json:"debate_id"`
	AgentID   string    `json:"agent_id"` // "user" for user turns
	Number    int       `json:"number"`   // Sequential turn number
	Round     int       `json:"round"`    // Round number
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`

	// Metadata from provider response
	TurnType     TurnType `json:"turn_type,omitempty"`      // Type of turn for separate tracking
	InputTokens  int      `json:"input_tokens,omitempty"`   // Input tokens from provider
	OutputTokens int      `json:"output_tokens,omitempty"`  // Output tokens from provider
	TotalTokens  int      `json:"total_tokens,omitempty"`   // Total tokens
	DurationMs   int64    `json:"duration_ms,omitempty"`    // Duration in milliseconds from provider
	Model        string   `json:"model,omitempty"`          // Model used for this turn
	StopReason   string   `json:"stop_reason,omitempty"`    // Stop reason from provider
}

// DebateStats contains aggregated usage statistics for a debate.
type DebateStats struct {
	// Overall totals
	TotalInputTokens  int   `json:"total_input_tokens"`
	TotalOutputTokens int   `json:"total_output_tokens"`
	TotalTokens       int   `json:"total_tokens"`
	TotalDurationMs   int64 `json:"total_duration_ms"`
	TurnCount         int   `json:"turn_count"`

	// Per-agent breakdown
	AgentAInputTokens  int   `json:"agent_a_input_tokens"`
	AgentAOutputTokens int   `json:"agent_a_output_tokens"`
	AgentATotalTokens  int   `json:"agent_a_total_tokens"`
	AgentADurationMs   int64 `json:"agent_a_duration_ms"`
	AgentATurnCount    int   `json:"agent_a_turn_count"`

	AgentBInputTokens  int   `json:"agent_b_input_tokens"`
	AgentBOutputTokens int   `json:"agent_b_output_tokens"`
	AgentBTotalTokens  int   `json:"agent_b_total_tokens"`
	AgentBDurationMs   int64 `json:"agent_b_duration_ms"`
	AgentBTurnCount    int   `json:"agent_b_turn_count"`

	// Conclusion/voting stats (tracked separately)
	ConclusionInputTokens  int   `json:"conclusion_input_tokens"`
	ConclusionOutputTokens int   `json:"conclusion_output_tokens"`
	ConclusionTotalTokens  int   `json:"conclusion_total_tokens"`
	ConclusionDurationMs   int64 `json:"conclusion_duration_ms"`
	ConclusionTurnCount    int   `json:"conclusion_turn_count"`
}

// ComputeDebateStats computes aggregated statistics from turns.
func ComputeDebateStats(turns []*Turn, agentAID, agentBID string) *DebateStats {
	stats := &DebateStats{}

	for _, turn := range turns {
		// Overall totals
		stats.TotalInputTokens += turn.InputTokens
		stats.TotalOutputTokens += turn.OutputTokens
		stats.TotalTokens += turn.TotalTokens
		stats.TotalDurationMs += turn.DurationMs
		stats.TurnCount++

		// Categorize by turn type and agent
		switch turn.TurnType {
		case TurnTypeVote, TurnTypeConclusion:
			stats.ConclusionInputTokens += turn.InputTokens
			stats.ConclusionOutputTokens += turn.OutputTokens
			stats.ConclusionTotalTokens += turn.TotalTokens
			stats.ConclusionDurationMs += turn.DurationMs
			stats.ConclusionTurnCount++
		default:
			// Regular debate turns - attribute to agent
			if turn.AgentID == agentAID {
				stats.AgentAInputTokens += turn.InputTokens
				stats.AgentAOutputTokens += turn.OutputTokens
				stats.AgentATotalTokens += turn.TotalTokens
				stats.AgentADurationMs += turn.DurationMs
				stats.AgentATurnCount++
			} else if turn.AgentID == agentBID {
				stats.AgentBInputTokens += turn.InputTokens
				stats.AgentBOutputTokens += turn.OutputTokens
				stats.AgentBTotalTokens += turn.TotalTokens
				stats.AgentBDurationMs += turn.DurationMs
				stats.AgentBTurnCount++
			}
		}
	}

	return stats
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

// ResponseType represents the type of council response for tracking.
type ResponseType string

const (
	ResponseTypeResponse  ResponseType = "response"  // Stage 1: Member response
	ResponseTypeRanking   ResponseType = "ranking"   // Stage 2: Ranking
	ResponseTypeSynthesis ResponseType = "synthesis" // Stage 3: Chairman synthesis
)

// Response represents a council member's response in Stage 1.
type Response struct {
	ID        string    `json:"id"`
	CouncilID string    `json:"council_id"`
	MemberID  string    `json:"member_id"`
	Round     int       `json:"round"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`

	// Metadata from provider response
	ResponseType ResponseType `json:"response_type,omitempty"` // Type of response for tracking
	InputTokens  int          `json:"input_tokens,omitempty"`
	OutputTokens int          `json:"output_tokens,omitempty"`
	TotalTokens  int          `json:"total_tokens,omitempty"`
	DurationMs   int64        `json:"duration_ms,omitempty"`
	Model        string       `json:"model,omitempty"`
	StopReason   string       `json:"stop_reason,omitempty"`
}

// CouncilStats contains aggregated usage statistics for a council session.
type CouncilStats struct {
	// Overall totals
	TotalInputTokens  int   `json:"total_input_tokens"`
	TotalOutputTokens int   `json:"total_output_tokens"`
	TotalTokens       int   `json:"total_tokens"`
	TotalDurationMs   int64 `json:"total_duration_ms"`
	ResponseCount     int   `json:"response_count"`

	// Per-member breakdown (map of member_id to stats)
	MemberStats map[string]*MemberStats `json:"member_stats,omitempty"`

	// Stage breakdown
	Stage1InputTokens  int   `json:"stage1_input_tokens"`  // Response collection
	Stage1OutputTokens int   `json:"stage1_output_tokens"`
	Stage1TotalTokens  int   `json:"stage1_total_tokens"`
	Stage1DurationMs   int64 `json:"stage1_duration_ms"`

	Stage2InputTokens  int   `json:"stage2_input_tokens"` // Rankings
	Stage2OutputTokens int   `json:"stage2_output_tokens"`
	Stage2TotalTokens  int   `json:"stage2_total_tokens"`
	Stage2DurationMs   int64 `json:"stage2_duration_ms"`

	Stage3InputTokens  int   `json:"stage3_input_tokens"` // Synthesis
	Stage3OutputTokens int   `json:"stage3_output_tokens"`
	Stage3TotalTokens  int   `json:"stage3_total_tokens"`
	Stage3DurationMs   int64 `json:"stage3_duration_ms"`
}

// MemberStats contains usage statistics for a single council member.
type MemberStats struct {
	MemberID     string `json:"member_id"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	TotalTokens  int    `json:"total_tokens"`
	DurationMs   int64  `json:"duration_ms"`
	ResponseCount int   `json:"response_count"`
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
