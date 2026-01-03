// Package engine orchestrates debate sessions between AI agents.
package engine

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/alienxp03/conclave/internal/core"
	"github.com/alienxp03/conclave/internal/persona"
	"github.com/alienxp03/conclave/internal/provider"
	"github.com/alienxp03/conclave/internal/storage"
	"github.com/alienxp03/conclave/internal/style"
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
	slog.Debug("Creating new debate", "topic", config.Topic, "agent_a", config.AgentAProvider, "agent_b", config.AgentBProvider)
	// Validate providers
	if config.AgentAProvider == "" {
		return nil, fmt.Errorf("agent A provider is required")
	}
	providerA, err := e.registry.Get(config.AgentAProvider)
	if err != nil {
		return nil, fmt.Errorf("invalid provider for agent A: %w", err)
	}
	if !providerA.Available() {
		return nil, fmt.Errorf("provider %s is not available (CLI not found)", config.AgentAProvider)
	}

	if config.AgentBProvider == "" {
		return nil, fmt.Errorf("agent B provider is required")
	}
	providerB, err := e.registry.Get(config.AgentBProvider)
	if err != nil {
		return nil, fmt.Errorf("invalid provider for agent B: %w", err)
	}
	if !providerB.Available() {
		return nil, fmt.Errorf("provider %s is not available (CLI not found)", config.AgentBProvider)
	}

	// Validate personas (check builtin first, then storage)
	personaADef := e.getPersona(config.AgentAPersona)
	if personaADef == nil {
		return nil, fmt.Errorf("invalid persona for agent A: %s", config.AgentAPersona)
	}
	personaBDef := e.getPersona(config.AgentBPersona)
	if personaBDef == nil {
		return nil, fmt.Errorf("invalid persona for agent B: %s", config.AgentBPersona)
	}

	// Validate style (check builtin first, then storage)
	styleDef := e.getStyle(config.Style)
	if styleDef == nil {
		return nil, fmt.Errorf("invalid debate style: %s", config.Style)
	}

	// Assign default models if empty
	agentAModel := config.AgentAModel
	if agentAModel == "" {
		agentAModel = core.DefaultModelForProvider[config.AgentAProvider]
	}
	agentBModel := config.AgentBModel
	if agentBModel == "" {
		agentBModel = core.DefaultModelForProvider[config.AgentBProvider]
	}

	// Set defaults
	maxTurns := config.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 5
	}

	now := time.Now()
	cwd, _ := os.Getwd()
	debate := &core.Debate{
		ID:    core.GenerateID(),
		Topic: config.Topic,
		CWD:   cwd,
		AgentA: core.Agent{
			ID:       core.GenerateID(),
			Name:     fmt.Sprintf("Agent A (%s)", personaADef.Name),
			Provider: config.AgentAProvider,
			Model:    agentAModel,
			Persona:  config.AgentAPersona,
		},
		AgentB: core.Agent{
			ID:       core.GenerateID(),
			Name:     fmt.Sprintf("Agent B (%s)", personaBDef.Name),
			Provider: config.AgentBProvider,
			Model:    agentBModel,
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

	// Summarize topic in background
	go e.AutoSummarize(debate.ID, config.Topic, config.AgentAProvider)

	return debate, nil
}

// AutoSummarize generates a summary title and updates the debate.
func (e *Engine) AutoSummarize(id, topic, providerName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	title, err := e.SummarizeTopic(ctx, topic, providerName)
	if err != nil {
		slog.Error("Failed to auto-summarize topic", "id", id, "error", err)
		return
	}

	if err := e.storage.UpdateDebateTitle(id, title); err != nil {
		slog.Error("Failed to update debate title", "id", id, "error", err)
	}
}

// SummarizeTopic generates a 3-4 word title from a topic.
func (e *Engine) SummarizeTopic(ctx context.Context, topic string, providerName string) (string, error) {
	prov, err := e.registry.Get(providerName)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`Summarize the following question or topic as a 3-4 word title. 
Topic: "%s"

Respond with ONLY the title. No punctuation.`, topic)

	var response string
	response, err = prov.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	title := strings.TrimSpace(response)
	title = strings.Trim(title, `"'`)
	// Ensure it's not too long
	words := strings.Fields(title)
	if len(words) > 5 {
		title = strings.Join(words[:5], " ")
	}

	return title, nil
}

// getPersona retrieves a persona by ID from builtins or storage.
func (e *Engine) getPersona(id string) *persona.Persona {
	// Check builtin first
	if p := persona.Get(id); p != nil {
		return p
	}

	// Check storage for custom persona
	stored, err := e.storage.GetPersona(id)
	if err != nil || stored == nil {
		return nil
	}

	return &persona.Persona{
		ID:           stored.ID,
		Name:         stored.Name,
		Description:  stored.Description,
		SystemPrompt: stored.SystemPrompt,
	}
}

// getStyle retrieves a style by ID from builtins or storage.
func (e *Engine) getStyle(id string) *style.Style {
	// Check builtin first
	if s := style.Get(id); s != nil {
		return s
	}

	// Check storage for custom style
	stored, err := e.storage.GetStyle(id)
	if err != nil || stored == nil {
		return nil
	}

	return &style.Style{
		ID:               stored.ID,
		Name:             stored.Name,
		Description:      stored.Description,
		OpeningPrompt:    stored.OpeningPrompt,
		ResponsePrompt:   stored.ResponsePrompt,
		ConclusionPrompt: stored.ConclusionPrompt,
	}
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

	// Get current turns to determine progress
	turns, _ := e.storage.GetTurns(debateID)
	currentRound := 1
	if len(turns) > 0 {
		currentRound = turns[len(turns)-1].Round
	}

	// Count agent turns in current round
	turnsInRound := 0
	for _, t := range turns {
		if t.Round == currentRound && t.AgentID != "user" {
			turnsInRound++
		}
	}

	// Decide who starts the round (consistent for the round)
	agents := []core.Agent{debate.AgentA, debate.AgentB}
	// Use debate ID + round as seed
	r := rand.New(rand.NewSource(int64(hashString(debate.ID) + uint32(currentRound))))
	r.Shuffle(len(agents), func(i, j int) { agents[i], agents[j] = agents[j], agents[i] })

	// Execute remaining turns in round
	totalTurnsInRound := debate.MaxTurns * 2
	earlyConsensus := false

	for i := turnsInRound + 1; i <= totalTurnsInRound; i++ {
		select {
		case <-ctx.Done():
			debate.Status = core.StatusFailed
			e.storage.UpdateDebate(debate)
			return ctx.Err()
		default:
		}

		// Alternate between agents
		currentAgent := agents[(i-1)%2]
		isLastTurn := i == totalTurnsInRound
		turnNum := len(turns) + 1

		turn, err := e.executeTurn(ctx, debate, currentAgent, turnNum, isLastTurn)
		if err != nil {
			debate.Status = core.StatusFailed
			e.storage.UpdateDebate(debate)
			return fmt.Errorf("failed to execute turn %d: %w", turnNum, err)
		}

		// Update turns list for next iteration
		turns = append(turns, turn)

		if callback != nil {
			callback(turn, debate)
		}

		// Check for early consensus after each complete round (both agents spoke)
		// Start checking after 2 turns in current round (1 round of discussion)
		if i >= 2 && i%2 == 0 && i < totalTurnsInRound {
			if e.checkEarlyConsensus(ctx, debate) {
				earlyConsensus = true
				break
			}
		}
	}

	// Generate conclusion
	conclusion, err := e.generateConclusion(ctx, debate)
	if err != nil {
		// Don't fail the whole debate, just log
		conclusion = &core.Conclusion{
			Agreed:  false,
			Summary: "Unable to generate conclusion: " + err.Error(),
		}
	}

	if earlyConsensus {
		conclusion.EarlyConsensus = true
	}

	// Set round for conclusion
	turns, _ = e.storage.GetTurns(debate.ID)
	if len(turns) > 0 {
		conclusion.Round = turns[len(turns)-1].Round
	} else {
		conclusion.Round = 1
	}

	debate.Conclusions = append(debate.Conclusions, conclusion)

	// Mark as completed
	now := time.Now()
	debate.Status = core.StatusCompleted
	debate.CompletedAt = &now
	if err := e.storage.UpdateDebate(debate); err != nil {
		return fmt.Errorf("failed to update debate: %w", err)
	}

	return nil
}

// checkEarlyConsensus checks if both agents have reached agreement.
func (e *Engine) checkEarlyConsensus(ctx context.Context, debate *core.Debate) bool {
	turns, err := e.storage.GetTurns(debate.ID)
	if err != nil || len(turns) < 2 {
		return false
	}

	// Get last two turns (one from each agent)
	lastTurn := turns[len(turns)-1]
	prevTurn := turns[len(turns)-2]

	// Check for consensus signals in recent responses
	consensusSignals := []string{
		"i agree",
		"we agree",
		"consensus",
		"common ground",
		"we've reached",
		"i concur",
		"you're right",
		"you are right",
		"that's a fair point",
		"i accept",
		"we can conclude",
		"in agreement",
	}

	lastLower := strings.ToLower(lastTurn.Content)
	prevLower := strings.ToLower(prevTurn.Content)

	lastHasSignal := false
	prevHasSignal := false

	for _, signal := range consensusSignals {
		if strings.Contains(lastLower, signal) {
			lastHasSignal = true
		}
		if strings.Contains(prevLower, signal) {
			prevHasSignal = true
		}
	}

	// If both recent turns show agreement signals, verify with a quick check
	if lastHasSignal && prevHasSignal {
		return e.verifyConsensus(ctx, debate)
	}

	return false
}

// verifyConsensus asks one agent to confirm if consensus has been reached.
func (e *Engine) verifyConsensus(ctx context.Context, debate *core.Debate) bool {
	prov, err := e.registry.Get(debate.AgentA.Provider)
	if err != nil {
		return false
	}

	turns, _ := e.storage.GetTurns(debate.ID)
	history := e.buildDebateHistory(debate, turns)

	prompt := fmt.Sprintf(`You are reviewing a debate on: "%s"

Recent discussion:
%s

Based on the last few exchanges, have both participants clearly reached a consensus or agreement on the main points?

Answer with only YES or NO.`, debate.Topic, history)

	var response string
	if debate.AgentA.Model != "" {
		response, err = prov.GenerateWithModel(ctx, prompt, debate.AgentA.Model)
	} else {
		response, err = prov.Generate(ctx, prompt)
	}

	if err != nil {
		return false
	}

	responseLower := strings.ToLower(strings.TrimSpace(response))
	return strings.HasPrefix(responseLower, "yes")
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

	// Generate response (with model if specified)
	var response string
	if agent.Model != "" {
		response, err = prov.GenerateWithModel(ctx, prompt, agent.Model)
	} else {
		response, err = prov.Generate(ctx, prompt)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Create turn
	round := 1
	if len(turns) > 0 {
		round = turns[len(turns)-1].Round
	}

	turn := &core.Turn{
		ID:        core.GenerateID(),
		DebateID:  debate.ID,
		AgentID:   agent.ID,
		Number:    turnNum,
		Round:     round,
		Content:   response,
		CreatedAt: time.Now(),
	}

	if err := e.storage.AddTurn(turn); err != nil {
		return nil, fmt.Errorf("failed to save turn: %w", err)
	}

	// Check for consensus or max turns
	if err := e.checkConclusion(ctx, debate, turns); err != nil {
		// Log error but don't fail the turn
		slog.Error("Failed to check conclusion", "error", err)
	}

	slog.Debug("Turn execution completed", "turn_id", turn.ID, "agent", turn.AgentID)
	return turn, nil
}

// buildPrompt constructs the prompt for an agent's turn.
func (e *Engine) buildPrompt(debate *core.Debate, agent core.Agent, turns []*core.Turn, turnNum int, isLastTurn bool) (string, error) {
	personaDef := e.getPersona(agent.Persona)
	styleDef := e.getStyle(debate.Style)

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
		} else if t.AgentID == debate.AgentB.ID {
			agentName = debate.AgentB.Name
		} else if t.AgentID == "user" {
			agentName = "User (Follow-up)"
		} else {
			agentName = "Unknown"
		}
		historyBuilder.WriteString(fmt.Sprintf("\n--- %s (Round %d, Turn %d) ---\n%s\n", agentName, t.Round, t.Number, t.Content))
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

%s

Use Markdown format.`, personaDef.SystemPrompt, buf.String())

	return fullPrompt, nil
}

// generateConclusion generates the final conclusion for the debate with voting.
func (e *Engine) generateConclusion(ctx context.Context, debate *core.Debate) (*core.Conclusion, error) {
	turns, err := e.storage.GetTurns(debate.ID)
	if err != nil {
		return nil, err
	}

	// Build history
	history := e.buildDebateHistory(debate, turns)

	conclusion := &core.Conclusion{}

	// Get votes from both agents
	voteA, err := e.getAgentVote(ctx, debate, debate.AgentA, history)
	if err == nil {
		conclusion.AgentAVote = voteA
	}

	voteB, err := e.getAgentVote(ctx, debate, debate.AgentB, history)
	if err == nil {
		conclusion.AgentBVote = voteB
	}

	// Determine consensus based on votes
	if conclusion.AgentAVote != nil && conclusion.AgentBVote != nil {
		conclusion.Agreed = conclusion.AgentAVote.Agrees && conclusion.AgentBVote.Agrees
	}

	// Generate summary
	summary, err := e.generateSummary(ctx, debate, history, conclusion)
	if err == nil {
		conclusion.Summary = summary
	} else {
		conclusion.Summary = "Debate concluded."
	}

	// Get individual positions if no consensus
	if !conclusion.Agreed {
		if conclusion.AgentAVote != nil {
			conclusion.AgentASummary = conclusion.AgentAVote.Reasoning
		}
		if conclusion.AgentBVote != nil {
			conclusion.AgentBSummary = conclusion.AgentBVote.Reasoning
		}
	}

	return conclusion, nil
}

// buildDebateHistory builds a formatted string of the debate history.
func (e *Engine) buildDebateHistory(debate *core.Debate, turns []*core.Turn) string {
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
	return historyBuilder.String()
}

// getAgentVote asks an agent to vote on whether consensus was reached.
func (e *Engine) getAgentVote(ctx context.Context, debate *core.Debate, agent core.Agent, history string) (*core.Vote, error) {
	prov, err := e.registry.Get(agent.Provider)
	if err != nil {
		return nil, err
	}

	votePrompt := fmt.Sprintf(`You participated in a debate on: "%s"

Here is the full debate:
%s

Now it's time to conclude. Please vote on whether you and your opponent reached a meaningful consensus.

Consider:
- Did you find common ground on the main points?
- Are there fundamental disagreements that remain?
- Would you be comfortable with a joint conclusion?

Respond in this exact format:
VOTE: [AGREE/DISAGREE]
REASONING: [Brief explanation of your vote - 1-2 sentences]`, debate.Topic, history)

	var response string
	if agent.Model != "" {
		response, err = prov.GenerateWithModel(ctx, votePrompt, agent.Model)
	} else {
		response, err = prov.Generate(ctx, votePrompt)
	}
	if err != nil {
		return nil, err
	}

	vote := &core.Vote{
		AgentID: agent.ID,
	}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "VOTE:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "VOTE:"))
			vote.Agrees = strings.EqualFold(val, "AGREE")
		} else if strings.HasPrefix(line, "REASONING:") {
			vote.Reasoning = strings.TrimSpace(strings.TrimPrefix(line, "REASONING:"))
		}
	}

	// If no reasoning found, use full response
	if vote.Reasoning == "" {
		vote.Reasoning = response
	}

	return vote, nil
}

// generateSummary generates the final summary of the debate.
func (e *Engine) generateSummary(ctx context.Context, debate *core.Debate, history string, conclusion *core.Conclusion) (string, error) {
	prov, err := e.registry.Get(debate.AgentA.Provider)
	if err != nil {
		return "", err
	}

	consensusStatus := "No consensus was reached."
	if conclusion.Agreed {
		consensusStatus = "Both agents agreed on a consensus."
	}

	summaryPrompt := fmt.Sprintf(`Summarize this debate on: "%s"

Debate history:
%s

Voting results: %s

Provide a brief, objective summary (2-3 sentences) of the key points discussed and the outcome.

Use Markdown format.`, debate.Topic, history, consensusStatus)

	var response string
	if debate.AgentA.Model != "" {
		response, err = prov.GenerateWithModel(ctx, summaryPrompt, debate.AgentA.Model)
	} else {
		response, err = prov.Generate(ctx, summaryPrompt)
	}
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response), nil
}

// ExecuteNextTurn executes just the next turn (for step-by-step execution).
func (e *Engine) ExecuteNextTurn(ctx context.Context, debateID string) (*core.Turn, error) {
	slog.Debug("Executing next turn", "debate_id", debateID)
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
		if conclusion != nil {
			conclusion.Round = turn.Round
			debate.Conclusions = append(debate.Conclusions, conclusion)
		}
		now := time.Now()
		debate.Status = core.StatusCompleted
		debate.CompletedAt = &now
		e.storage.UpdateDebate(debate)
	}

	return turn, nil
}

// checkConclusion checks if the debate should conclude early.
func (e *Engine) checkConclusion(ctx context.Context, debate *core.Debate, turns []*core.Turn) error {
	// Only check if we have enough turns in the current round
	if len(turns) < 4 {
		return nil
	}

	// Check for early consensus
	if e.checkEarlyConsensus(ctx, debate) {
		conclusion, err := e.generateConclusion(ctx, debate)
		if err != nil {
			return err
		}
		conclusion.EarlyConsensus = true
		if len(turns) > 0 {
			conclusion.Round = turns[len(turns)-1].Round
		}
		debate.Conclusions = append(debate.Conclusions, conclusion)

		now := time.Now()
		debate.Status = core.StatusCompleted
		debate.CompletedAt = &now
		return e.storage.UpdateDebate(debate)
	}

	return nil
}

// AddFollowUp adds a user follow-up question and resumes the debate.
func (e *Engine) AddFollowUp(ctx context.Context, debateID string, content string) error {
	debate, turns, err := e.GetDebateWithTurns(debateID)
	if err != nil {
		return err
	}
	if debate == nil {
		return fmt.Errorf("debate not found")
	}

	newRound := 1
	if len(turns) > 0 {
		newRound = turns[len(turns)-1].Round + 1
	}

	userTurn := &core.Turn{
		ID:        core.GenerateID(),
		DebateID:  debate.ID,
		AgentID:   "user",
		Number:    len(turns) + 1,
		Round:     newRound,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := e.storage.AddTurn(userTurn); err != nil {
		return fmt.Errorf("failed to save user turn: %w", err)
	}

	// Trigger background run
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		e.RunDebate(ctx, debateID, nil)
	}()

	return nil
}

// hashString creates a simple hash from a string.
func hashString(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}
