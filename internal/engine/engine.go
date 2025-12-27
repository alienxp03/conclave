// Package engine orchestrates debate sessions between AI agents.
package engine

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"

	"github.com/alienxp03/dbate/internal/core"
	"github.com/alienxp03/dbate/internal/persona"
	"github.com/alienxp03/dbate/internal/provider"
	"github.com/alienxp03/dbate/internal/storage"
	"github.com/alienxp03/dbate/internal/style"
)

// Engine orchestrates debate sessions.
type Engine struct {
	storage  storage.Storage
	registry *provider.Registry
}

// New creates a new debate engine.
func New(store storage.Storage, registry *provider.Registry) *Engine {
	return &Engine{
		storage:  store,
		registry: registry,
	}
}

// CreateDebate creates a new debate session.
func (e *Engine) CreateDebate(ctx context.Context, config core.NewDebateConfig) (*core.Debate, error) {
	// Validate providers
	providerA, err := e.registry.Get(config.AgentAProvider)
	if err != nil {
		return nil, fmt.Errorf("invalid provider for agent A: %w", err)
	}
	if !providerA.Available() {
		return nil, fmt.Errorf("provider %s is not available (CLI not found)", config.AgentAProvider)
	}

	providerB, err := e.registry.Get(config.AgentBProvider)
	if err != nil {
		return nil, fmt.Errorf("invalid provider for agent B: %w", err)
	}
	if !providerB.Available() {
		return nil, fmt.Errorf("provider %s is not available (CLI not found)", config.AgentBProvider)
	}

	// Validate personas
	if !persona.Valid(config.AgentAPersona) {
		return nil, fmt.Errorf("invalid persona for agent A: %s", config.AgentAPersona)
	}
	if !persona.Valid(config.AgentBPersona) {
		return nil, fmt.Errorf("invalid persona for agent B: %s", config.AgentBPersona)
	}

	// Validate style
	if !style.Valid(config.Style) {
		return nil, fmt.Errorf("invalid debate style: %s", config.Style)
	}

	// Set defaults
	maxTurns := config.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 5
	}

	now := time.Now()
	debate := &core.Debate{
		ID:    uuid.New().String(),
		Topic: config.Topic,
		AgentA: core.Agent{
			ID:       uuid.New().String(),
			Name:     fmt.Sprintf("Agent A (%s)", persona.Get(config.AgentAPersona).Name),
			Provider: config.AgentAProvider,
			Persona:  config.AgentAPersona,
		},
		AgentB: core.Agent{
			ID:       uuid.New().String(),
			Name:     fmt.Sprintf("Agent B (%s)", persona.Get(config.AgentBPersona).Name),
			Provider: config.AgentBProvider,
			Persona:  config.AgentBPersona,
		},
		Style:     config.Style,
		MaxTurns:  maxTurns,
		Status:    core.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := e.storage.CreateDebate(debate); err != nil {
		return nil, fmt.Errorf("failed to create debate: %w", err)
	}

	return debate, nil
}

// GetDebate retrieves a debate by ID.
func (e *Engine) GetDebate(id string) (*core.Debate, error) {
	return e.storage.GetDebate(id)
}

// GetDebateWithTurns retrieves a debate with all its turns.
func (e *Engine) GetDebateWithTurns(id string) (*core.Debate, []*core.Turn, error) {
	debate, err := e.storage.GetDebate(id)
	if err != nil {
		return nil, nil, err
	}
	if debate == nil {
		return nil, nil, nil
	}

	turns, err := e.storage.GetTurns(id)
	if err != nil {
		return nil, nil, err
	}

	return debate, turns, nil
}

// ListDebates returns a list of debates.
func (e *Engine) ListDebates(limit, offset int) ([]*core.DebateSummary, error) {
	if limit <= 0 {
		limit = 20
	}
	return e.storage.ListDebates(limit, offset)
}

// DeleteDebate deletes a debate.
func (e *Engine) DeleteDebate(id string) error {
	return e.storage.DeleteDebate(id)
}

// TurnCallback is called after each turn completes.
type TurnCallback func(turn *core.Turn, debate *core.Debate)

// RunDebate executes the entire debate from start to finish.
func (e *Engine) RunDebate(ctx context.Context, debateID string, callback TurnCallback) error {
	debate, err := e.storage.GetDebate(debateID)
	if err != nil {
		return fmt.Errorf("failed to get debate: %w", err)
	}
	if debate == nil {
		return fmt.Errorf("debate not found: %s", debateID)
	}

	// Update status to in progress
	debate.Status = core.StatusInProgress
	if err := e.storage.UpdateDebate(debate); err != nil {
		return fmt.Errorf("failed to update debate status: %w", err)
	}

	// Randomly decide who starts
	agents := []core.Agent{debate.AgentA, debate.AgentB}
	rand.Shuffle(len(agents), func(i, j int) { agents[i], agents[j] = agents[j], agents[i] })

	// Execute turns
	totalTurns := debate.MaxTurns * 2
	for turnNum := 1; turnNum <= totalTurns; turnNum++ {
		select {
		case <-ctx.Done():
			debate.Status = core.StatusFailed
			e.storage.UpdateDebate(debate)
			return ctx.Err()
		default:
		}

		// Alternate between agents
		currentAgent := agents[(turnNum-1)%2]
		isLastTurn := turnNum == totalTurns

		turn, err := e.executeTurn(ctx, debate, currentAgent, turnNum, isLastTurn)
		if err != nil {
			debate.Status = core.StatusFailed
			e.storage.UpdateDebate(debate)
			return fmt.Errorf("failed to execute turn %d: %w", turnNum, err)
		}

		if callback != nil {
			callback(turn, debate)
		}
	}

	// Generate conclusion
	conclusion, err := e.generateConclusion(ctx, debate)
	if err != nil {
		// Don't fail the whole debate, just log
		debate.Conclusion = &core.Conclusion{
			Agreed:  false,
			Summary: "Unable to generate conclusion: " + err.Error(),
		}
	} else {
		debate.Conclusion = conclusion
	}

	// Mark as completed
	now := time.Now()
	debate.Status = core.StatusCompleted
	debate.CompletedAt = &now
	if err := e.storage.UpdateDebate(debate); err != nil {
		return fmt.Errorf("failed to update debate: %w", err)
	}

	return nil
}

// executeTurn executes a single turn in the debate.
func (e *Engine) executeTurn(ctx context.Context, debate *core.Debate, agent core.Agent, turnNum int, isLastTurn bool) (*core.Turn, error) {
	// Get provider
	prov, err := e.registry.Get(agent.Provider)
	if err != nil {
		return nil, err
	}

	// Get previous turns
	turns, err := e.storage.GetTurns(debate.ID)
	if err != nil {
		return nil, err
	}

	// Build prompt
	prompt, err := e.buildPrompt(debate, agent, turns, turnNum, isLastTurn)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Generate response
	response, err := prov.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Create turn
	turn := &core.Turn{
		ID:        uuid.New().String(),
		DebateID:  debate.ID,
		AgentID:   agent.ID,
		Number:    turnNum,
		Content:   response,
		CreatedAt: time.Now(),
	}

	if err := e.storage.AddTurn(turn); err != nil {
		return nil, fmt.Errorf("failed to save turn: %w", err)
	}

	return turn, nil
}

// buildPrompt constructs the prompt for an agent's turn.
func (e *Engine) buildPrompt(debate *core.Debate, agent core.Agent, turns []*core.Turn, turnNum int, isLastTurn bool) (string, error) {
	personaDef := persona.Get(agent.Persona)
	styleDef := style.Get(debate.Style)

	if personaDef == nil || styleDef == nil {
		return "", fmt.Errorf("invalid persona or style")
	}

	var promptTemplate string
	if turnNum == 1 || (turnNum == 2 && len(turns) == 0) {
		// First turn for this agent
		promptTemplate = styleDef.OpeningPrompt
	} else if isLastTurn {
		promptTemplate = styleDef.ConclusionPrompt
	} else {
		promptTemplate = styleDef.ResponsePrompt
	}

	// Get the other agent
	var otherAgent core.Agent
	if agent.ID == debate.AgentA.ID {
		otherAgent = debate.AgentB
	} else {
		otherAgent = debate.AgentA
	}

	// Build previous argument
	var previousArgument string
	if len(turns) > 0 {
		previousArgument = turns[len(turns)-1].Content
	}

	// Build debate history
	var historyBuilder strings.Builder
	for _, t := range turns {
		var agentName string
		if t.AgentID == debate.AgentA.ID {
			agentName = debate.AgentA.Name
		} else {
			agentName = debate.AgentB.Name
		}
		historyBuilder.WriteString(fmt.Sprintf("\n--- %s (Turn %d) ---\n%s\n", agentName, t.Number, t.Content))
	}

	// Template data
	data := map[string]interface{}{
		"Topic":            debate.Topic,
		"AgentName":        agent.Name,
		"OtherAgentName":   otherAgent.Name,
		"PreviousArgument": previousArgument,
		"DebateHistory":    historyBuilder.String(),
		"TurnNumber":       turnNum,
		"MaxTurns":         debate.MaxTurns * 2,
		"IsQuestioner":     turnNum%2 == 1 && debate.Style == "socratic",
	}

	// Parse and execute template
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Combine persona and style prompt
	fullPrompt := fmt.Sprintf(`%s

%s`, personaDef.SystemPrompt, buf.String())

	return fullPrompt, nil
}

// generateConclusion generates the final conclusion for the debate.
func (e *Engine) generateConclusion(ctx context.Context, debate *core.Debate) (*core.Conclusion, error) {
	turns, err := e.storage.GetTurns(debate.ID)
	if err != nil {
		return nil, err
	}

	// Build history
	var historyBuilder strings.Builder
	for _, t := range turns {
		var agentName string
		if t.AgentID == debate.AgentA.ID {
			agentName = debate.AgentA.Name
		} else {
			agentName = debate.AgentB.Name
		}
		historyBuilder.WriteString(fmt.Sprintf("\n--- %s (Turn %d) ---\n%s\n", agentName, t.Number, t.Content))
	}

	// Use agent A to summarize
	prov, err := e.registry.Get(debate.AgentA.Provider)
	if err != nil {
		return nil, err
	}

	summaryPrompt := fmt.Sprintf(`You were part of a debate on: "%s"

Here is the full debate:
%s

Analyze this debate and provide a conclusion:
1. Did the debaters reach a consensus? (yes/no)
2. If yes, what was the agreed conclusion?
3. If no, summarize each side's final position.

Respond in this exact format:
CONSENSUS: [yes/no]
SUMMARY: [overall summary of the debate outcome]
AGENT_A_POSITION: [Agent A's final position, if no consensus]
AGENT_B_POSITION: [Agent B's final position, if no consensus]`, debate.Topic, historyBuilder.String())

	response, err := prov.Generate(ctx, summaryPrompt)
	if err != nil {
		return nil, err
	}

	// Parse response (simple parsing)
	conclusion := &core.Conclusion{}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "CONSENSUS:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "CONSENSUS:"))
			conclusion.Agreed = strings.EqualFold(val, "yes")
		} else if strings.HasPrefix(line, "SUMMARY:") {
			conclusion.Summary = strings.TrimSpace(strings.TrimPrefix(line, "SUMMARY:"))
		} else if strings.HasPrefix(line, "AGENT_A_POSITION:") {
			conclusion.AgentASummary = strings.TrimSpace(strings.TrimPrefix(line, "AGENT_A_POSITION:"))
		} else if strings.HasPrefix(line, "AGENT_B_POSITION:") {
			conclusion.AgentBSummary = strings.TrimSpace(strings.TrimPrefix(line, "AGENT_B_POSITION:"))
		}
	}

	// If parsing failed, use the whole response as summary
	if conclusion.Summary == "" {
		conclusion.Summary = response
	}

	return conclusion, nil
}

// ExecuteNextTurn executes just the next turn (for step-by-step execution).
func (e *Engine) ExecuteNextTurn(ctx context.Context, debateID string) (*core.Turn, error) {
	debate, err := e.storage.GetDebate(debateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debate: %w", err)
	}
	if debate == nil {
		return nil, fmt.Errorf("debate not found: %s", debateID)
	}

	if debate.Status == core.StatusCompleted {
		return nil, fmt.Errorf("debate is already completed")
	}

	// Get current turns
	turns, err := e.storage.GetTurns(debateID)
	if err != nil {
		return nil, err
	}

	currentTurnNum := len(turns) + 1
	totalTurns := debate.MaxTurns * 2

	if currentTurnNum > totalTurns {
		return nil, fmt.Errorf("all turns completed")
	}

	// Update status if first turn
	if debate.Status == core.StatusPending {
		debate.Status = core.StatusInProgress
		if err := e.storage.UpdateDebate(debate); err != nil {
			return nil, err
		}
	}

	// Determine starting order (consistent for the debate)
	agents := []core.Agent{debate.AgentA, debate.AgentB}
	// Use debate ID as seed for consistent ordering
	r := rand.New(rand.NewSource(int64(hashString(debate.ID))))
	r.Shuffle(len(agents), func(i, j int) { agents[i], agents[j] = agents[j], agents[i] })

	currentAgent := agents[(currentTurnNum-1)%2]
	isLastTurn := currentTurnNum == totalTurns

	turn, err := e.executeTurn(ctx, debate, currentAgent, currentTurnNum, isLastTurn)
	if err != nil {
		return nil, err
	}

	// Check if debate is complete
	if currentTurnNum == totalTurns {
		conclusion, _ := e.generateConclusion(ctx, debate)
		debate.Conclusion = conclusion
		now := time.Now()
		debate.Status = core.StatusCompleted
		debate.CompletedAt = &now
		e.storage.UpdateDebate(debate)
	}

	return turn, nil
}

// hashString creates a simple hash from a string.
func hashString(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}
