// Package council implements the multi-agent council system.
package council

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alienxp03/conclave/internal/core"
	"github.com/alienxp03/conclave/internal/persona"
	"github.com/alienxp03/conclave/internal/provider"
	"github.com/alienxp03/conclave/internal/storage"
)

// Engine orchestrates council sessions.
type Engine struct {
	storage  storage.Storage
	registry *provider.Registry
	running  sync.Map // councilID -> bool
}

// New creates a new council engine.
func New(store storage.Storage, registry *provider.Registry) *Engine {
	return &Engine{
		storage:  store,
		registry: registry,
	}
}

// CreateCouncil creates a new council session.
func (e *Engine) CreateCouncil(ctx context.Context, config core.NewCouncilConfig) (*core.Council, error) {
	// Validate and prepare members
	if len(config.Members) < 2 {
		return nil, fmt.Errorf("council must have at least 2 members")
	}

	// Assign default models and personas
	members := core.AssignDefaultModels(config.Members)
	members = core.AssignDefaultPersonas(members)

	// Create agents from member specs
	agents := make([]core.Agent, len(members))
	for i, member := range members {
		// Validate provider
		prov, err := e.registry.Get(member.Provider)
		if err != nil {
			return nil, fmt.Errorf("invalid provider for member %d: %w", i+1, err)
		}
		if !prov.Available() {
			return nil, fmt.Errorf("provider %s is not available (CLI not found)", member.Provider)
		}

		// Get persona definition
		personaDef := e.getPersona(member.Persona)
		if personaDef == nil {
			return nil, fmt.Errorf("invalid persona for member %d: %s", i+1, member.Persona)
		}

		agents[i] = core.Agent{
			ID:       core.GenerateID(),
			Name:     fmt.Sprintf("%s (%s)", member.Provider, personaDef.Name),
			Provider: member.Provider,
			Model:    member.Model,
			Persona:  member.Persona,
		}
	}

	// Determine chairman
	var chairman core.Agent
	if config.Chairman != nil {
		chairmanSpec := *config.Chairman
		if chairmanSpec.Model == "" {
			chairmanSpec.Model = core.BestModelForProvider[chairmanSpec.Provider]
		}

		chairman = core.Agent{
			ID:       core.GenerateID(),
			Name:     fmt.Sprintf("Chairman (%s)", chairmanSpec.Provider),
			Provider: chairmanSpec.Provider,
			Model:    chairmanSpec.Model,
			Persona:  "chairman",
		}
	} else {
		// Use default chairman (first member's provider with best model)
		defaultChairman := core.GetDefaultChairman(members)
		chairman = core.Agent{
			ID:       core.GenerateID(),
			Name:     fmt.Sprintf("Chairman (%s)", defaultChairman.Provider),
			Provider: defaultChairman.Provider,
			Model:    defaultChairman.Model,
			Persona:  "chairman",
		}
	}

	// Create council
	now := time.Now()
	cwd, _ := os.Getwd()
	council := &core.Council{
		ID:        core.GenerateID(),
		Topic:     config.Topic,
		CWD:       cwd,
		Members:   agents,
		Chairman:  chairman,
		Status:    core.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save to storage
	if err := e.storage.CreateCouncil(council); err != nil {
		return nil, fmt.Errorf("failed to save council: %w", err)
	}

	// Summarize topic in background
	go e.AutoSummarize(council.ID, config.Topic, council.Chairman.Provider)

	return council, nil
}

// AutoSummarize generates a summary title and updates the council.
func (e *Engine) AutoSummarize(id, topic, providerName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	title, err := e.SummarizeTopic(ctx, topic, providerName)
	if err != nil {
		slog.Error("Failed to auto-summarize topic", "id", id, "error", err)
		return
	}

	if err := e.storage.UpdateCouncilTitle(id, title); err != nil {
		slog.Error("Failed to update council title", "id", id, "error", err)
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

// CouncilCallbacks contains callback functions for council progress.
type CouncilCallbacks struct {
	OnResponseCollected func(response core.Response)
	OnRankingCollected  func(ranking core.Ranking)
	OnSynthesisComplete func(synthesis core.CouncilSynthesis)
	OnStageComplete     func(stage int)
}

// RunCouncil executes all 3 stages of the council process.
func (e *Engine) RunCouncil(ctx context.Context, council *core.Council) error {
	return e.RunCouncilWithCallbacks(ctx, council, nil)
}

// RunCouncilWithCallbacks executes all 3 stages with progress callbacks.
func (e *Engine) RunCouncilWithCallbacks(ctx context.Context, council *core.Council, callbacks *CouncilCallbacks) error {
	slog.Info("Starting council execution", "council_id", council.ID, "topic", council.Topic)

	// Check if already running
	if _, loaded := e.running.LoadOrStore(council.ID, true); loaded {
		slog.Warn("Council already running, skipping duplicate execution", "council_id", council.ID)
		return nil
	}
	defer e.running.Delete(council.ID)

	// Ensure council is in storage and set to in_progress
	existing, _ := e.storage.GetCouncil(council.ID)
	if existing == nil {
		if err := e.storage.CreateCouncil(council); err != nil {
			return fmt.Errorf("failed to create council: %w", err)
		}
	}

	// Update status to in_progress if it's not already
	if council.Status != core.StatusInProgress {
		council.Status = core.StatusInProgress
		council.UpdatedAt = time.Now()
		if err := e.storage.UpdateCouncil(council); err != nil {
			slog.Error("Failed to update council status", "error", err)
		}
	}

	// Stage 1: Collect responses
	slog.Debug("Stage 1: Collecting responses", "council_id", council.ID)
	currentResponses, err := e.CollectResponsesWithCallback(ctx, council, callbacks)
	if err != nil {
		slog.Error("Stage 1 failed", "error", err)
		council.Status = core.StatusFailed
		e.storage.UpdateCouncil(council)
		return fmt.Errorf("stage 1 failed: %w", err)
	}

	if callbacks != nil && callbacks.OnStageComplete != nil {
		callbacks.OnStageComplete(1)
	}

	// Save responses
	for _, r := range currentResponses {
		r := r // capture loop variable
		if err := e.storage.AddResponse(&r); err != nil {
			slog.Error("Failed to save response", "error", err)
			return fmt.Errorf("failed to save response: %w", err)
		}
	}

	// Stage 2: Collect rankings
	slog.Debug("Stage 2: Collecting rankings", "council_id", council.ID)
	currentRankings, err := e.CollectRankingsWithCallback(ctx, council, currentResponses, callbacks)
	if err != nil {
		slog.Error("Stage 2 failed", "error", err)
		council.Status = core.StatusFailed
		e.storage.UpdateCouncil(council)
		return fmt.Errorf("stage 2 failed: %w", err)
	}

	if callbacks != nil && callbacks.OnStageComplete != nil {
		callbacks.OnStageComplete(2)
	}

	// Save rankings
	for _, r := range currentRankings {
		r := r // capture loop variable
		if err := e.storage.AddRanking(&r); err != nil {
			slog.Error("Failed to save ranking", "error", err)
			return fmt.Errorf("failed to save ranking: %w", err)
		}
	}

	// Stage 3: Synthesis
	slog.Debug("Stage 3: Generating synthesis", "council_id", council.ID)
	synthesisContent, err := e.GenerateSynthesis(ctx, council, currentResponses, currentRankings)
	if err != nil {
		slog.Error("Stage 3 failed", "error", err)
		council.Status = core.StatusFailed
		e.storage.UpdateCouncil(council)
		return fmt.Errorf("stage 3 failed: %w", err)
	}

	// Update council
	round := 1
	if len(currentResponses) > 0 {
		round = currentResponses[0].Round
	}

	synthesis := &core.CouncilSynthesis{
		Round:     round,
		Content:   synthesisContent,
		CreatedAt: time.Now(),
	}

	if callbacks != nil && callbacks.OnSynthesisComplete != nil {
		callbacks.OnSynthesisComplete(*synthesis)
	}

	council.Syntheses = append(council.Syntheses, synthesis)
	council.Status = core.StatusCompleted
	e.storage.UpdateCouncil(council)

	slog.Info("Council execution completed", "council_id", council.ID)
	return nil
}

// CollectResponses implements Stage 1: collect independent responses from all members.
func (e *Engine) CollectResponses(ctx context.Context, council *core.Council) ([]core.Response, error) {
	return e.CollectResponsesWithCallback(ctx, council, nil)
}

// CollectResponsesWithCallback collects responses with progress callbacks.
func (e *Engine) CollectResponsesWithCallback(ctx context.Context, council *core.Council, callbacks *CouncilCallbacks) ([]core.Response, error) {
	type responseResult struct {
		agent    core.Agent
		response core.Response
		err      error
	}

	resultChan := make(chan responseResult, len(council.Members))

	// Get current round
	existingResponses, _ := e.storage.GetResponses(council.ID)
	round := 1
	if len(existingResponses) > 0 {
		round = existingResponses[len(existingResponses)-1].Round
	}

	for _, member := range council.Members {
		go func(agent core.Agent) {
			// Build prompt
			// If there are previous syntheses, include them in the prompt
			prompt, err := e.buildResponsePromptWithHistory(council, agent, existingResponses)
			if err != nil {
				resultChan <- responseResult{agent: agent, err: fmt.Errorf("failed to build prompt for %s: %w", agent.Name, err)}
				return
			}

			// Execute provider
			prov, err := e.registry.Get(agent.Provider)
			if err != nil {
				resultChan <- responseResult{agent: agent, err: fmt.Errorf("provider not found for %s: %w", agent.Name, err)}
				return
			}

			content, err := prov.GenerateWithModel(ctx, prompt, agent.Model)
			if err != nil {
				resultChan <- responseResult{agent: agent, err: fmt.Errorf("generation failed for %s: %w", agent.Name, err)}
				return
			}

			response := core.Response{
				ID:        core.GenerateID(),
				CouncilID: council.ID,
				MemberID:  agent.ID,
				Round:     round,
				Content:   content,
				CreatedAt: time.Now(),
			}

			resultChan <- responseResult{agent: agent, response: response}
		}(member)
	}

	// Collect all results
	responses := make([]core.Response, 0, len(council.Members))
	var firstErr error

	for i := 0; i < len(council.Members); i++ {
		result := <-resultChan
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}

		responses = append(responses, result.response)

		// Call callback if provided
		if callbacks != nil && callbacks.OnResponseCollected != nil {
			callbacks.OnResponseCollected(result.response)
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("no responses collected")
	}

	return responses, nil
}

// CollectRankings implements Stage 2: collect rankings from all members.
func (e *Engine) CollectRankings(ctx context.Context, council *core.Council, responses []core.Response) ([]core.Ranking, error) {
	return e.CollectRankingsWithCallback(ctx, council, responses, nil)
}

// CollectRankingsWithCallback collects rankings with progress callbacks.
func (e *Engine) CollectRankingsWithCallback(ctx context.Context, council *core.Council, responses []core.Response, callbacks *CouncilCallbacks) ([]core.Ranking, error) {
	// Format responses for ranking (using names)
	formattedResponses := e.formatResponsesForRanking(responses, council.Members)

	type rankingResult struct {
		agent   core.Agent
		ranking core.Ranking
		content string
		err     error
	}

	resultChan := make(chan rankingResult, len(council.Members))

	for _, member := range council.Members {
		go func(agent core.Agent) {
			// Build ranking prompt
			prompt := e.buildRankingPrompt(council, formattedResponses)

			// Execute provider
			prov, err := e.registry.Get(agent.Provider)
			if err != nil {
				resultChan <- rankingResult{agent: agent, err: fmt.Errorf("provider not found for %s: %w", agent.Name, err)}
				return
			}

			content, err := prov.GenerateWithModel(ctx, prompt, agent.Model)
			if err != nil {
				resultChan <- rankingResult{agent: agent, err: fmt.Errorf("generation failed for %s: %w", agent.Name, err)}
				return
			}

			// Parse rankings from response
			rankedIDs, err := e.parseRankingsFromText(content, responses, council.Members)
			if err != nil {
				resultChan <- rankingResult{agent: agent, content: content, err: fmt.Errorf("failed to parse rankings for %s: %w\n\nRaw response:\n%s", agent.Name, err, content)}
				return
			}

			round := 1
			if len(responses) > 0 {
				round = responses[0].Round
			}

			ranking := core.Ranking{
				ID:         core.GenerateID(),
				CouncilID:  council.ID,
				ReviewerID: agent.ID,
				Round:      round,
				Rankings:   rankedIDs,
				Reasoning:  content,
				CreatedAt:  time.Now(),
			}

			resultChan <- rankingResult{agent: agent, ranking: ranking, content: content}
		}(member)
	}

	// Collect all results
	rankings := make([]core.Ranking, 0, len(council.Members))
	var firstErr error

	for i := 0; i < len(council.Members); i++ {
		result := <-resultChan
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}

		rankings = append(rankings, result.ranking)

		// Call callback if provided
		if callbacks != nil && callbacks.OnRankingCollected != nil {
			callbacks.OnRankingCollected(result.ranking)
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	if len(rankings) == 0 {
		return nil, fmt.Errorf("no rankings collected")
	}

	return rankings, nil
}

// GenerateSynthesis implements Stage 3: chairman synthesizes all responses and rankings.
func (e *Engine) GenerateSynthesis(ctx context.Context, council *core.Council, responses []core.Response, rankings []core.Ranking) (string, error) {
	// Calculate aggregate rankings
	aggregateRanks := e.calculateAggregateRankings(responses, rankings, council.Members)

	// Build synthesis prompt
	prompt := e.buildSynthesisPrompt(council, responses, aggregateRanks)

	// Execute chairman
	prov, err := e.registry.Get(council.Chairman.Provider)
	if err != nil {
		return "", fmt.Errorf("chairman provider not found: %w", err)
	}

	synthesis, err := prov.GenerateWithModel(ctx, prompt, council.Chairman.Model)
	if err != nil {
		return "", fmt.Errorf("synthesis generation failed: %w", err)
	}

	return synthesis, nil
}

// getPersona returns the persona definition (builtin or custom).
func (e *Engine) getPersona(name string) *persona.Persona {
	// Check builtin first
	if p := persona.Get(name); p != nil {
		return p
	}

	// Check storage for custom personas
	customPersona, err := e.storage.GetPersona(name)
	if err == nil && customPersona != nil {
		return &persona.Persona{
			ID:           customPersona.ID,
			Name:         customPersona.Name,
			Description:  customPersona.Description,
			SystemPrompt: customPersona.SystemPrompt,
		}
	}

	return nil
}

func (e *Engine) buildResponsePromptWithHistory(council *core.Council, agent core.Agent, history []*core.Response) (string, error) {
	// Get persona
	personaDef := e.getPersona(agent.Persona)
	if personaDef == nil {
		return "", fmt.Errorf("persona not found: %s", agent.Persona)
	}

	// Build history text
	var historyText strings.Builder
	if len(council.Syntheses) > 0 {
		latestSynthesis := council.Syntheses[len(council.Syntheses)-1]
		historyText.WriteString(fmt.Sprintf("\nConclusion from Chairman %s:\n%s\n", council.Chairman.Name, latestSynthesis.Content))
	}

	// Check if there's a user directive in the current round
	var directive string
	round := 1
	if len(history) > 0 {
		round = history[len(history)-1].Round
	}
	for _, r := range history {
		if r.Round == round && r.MemberID == "user" {
			directive = r.Content
			break
		}
	}

	prompt := fmt.Sprintf(`%s

Topic: %s
%s`, personaDef.SystemPrompt, council.Topic, historyText.String())

	if directive != "" {
		prompt += fmt.Sprintf("\nFollow up question by user: \"%s\"\n\nPlease answer the question given by the user, taking into account the Chairman's previous conclusion.", directive)
	} else {
		prompt += "\n\nProvide your perspective on this topic. Focus on what matters most from your viewpoint."
	}

	prompt += "\n\nYour response:\n\nUse Markdown format."

	return prompt, nil
}

// AddFollowUp adds a user follow-up and resumes the council deliberation.
func (e *Engine) AddFollowUp(ctx context.Context, councilID string, content string) error {
	council, err := e.storage.GetCouncil(councilID)
	if err != nil {
		return err
	}

	// Determine new round
	responses, _ := e.storage.GetResponses(councilID)
	newRound := 1
	if len(responses) > 0 {
		newRound = responses[len(responses)-1].Round + 1
	}

	userResponse := core.Response{
		ID:        core.GenerateID(),
		CouncilID: council.ID,
		MemberID:  "user",
		Round:     newRound,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := e.storage.AddResponse(&userResponse); err != nil {
		return fmt.Errorf("failed to save user follow-up: %w", err)
	}

	// Set status to in_progress synchronously to avoid race condition with UI re-fetch
	council.Status = core.StatusInProgress
	council.UpdatedAt = time.Now()
	if err := e.storage.UpdateCouncil(council); err != nil {
		slog.Error("Failed to update council status during follow-up", "error", err)
	}

	// Trigger background run
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		e.RunCouncil(ctx, council)
	}()

	return nil
}

// Helper functions for prompt building (to be implemented)
// These will be implemented in prompts.go
func (e *Engine) buildResponsePrompt(topic string, agent core.Agent) (string, error) {
	// Get persona
	personaDef := e.getPersona(agent.Persona)
	if personaDef == nil {
		return "", fmt.Errorf("persona not found: %s", agent.Persona)
	}

	// Build prompt
	prompt := fmt.Sprintf(`%s

Topic: %s

Provide your perspective on this topic. Focus on what matters most from your viewpoint.

Your response:

Use Markdown format.`, personaDef.SystemPrompt, topic)

	return prompt, nil
}

func (e *Engine) buildRankingPrompt(council *core.Council, formattedResponses string) string {
	var contextText strings.Builder
	if len(council.Syntheses) > 0 {
		latestSynthesis := council.Syntheses[len(council.Syntheses)-1]
		contextText.WriteString(fmt.Sprintf("\nConclusion from Chairman %s:\n%s\n", council.Chairman.Name, latestSynthesis.Content))
	}

	// Check for directive in current round
	existingResponses, _ := e.storage.GetResponses(council.ID)
	var directive string
	round := 1
	if len(existingResponses) > 0 {
		round = existingResponses[len(existingResponses)-1].Round
	}
	for _, r := range existingResponses {
		if r.Round == round && r.MemberID == "user" {
			directive = r.Content
			break
		}
	}

	if directive != "" {
		contextText.WriteString(fmt.Sprintf("\nFollow up question by user: \"%s\"\n", directive))
	}

	return fmt.Sprintf(`You are evaluating multiple responses to the following topic:

Topic: %s
%s
Here are the responses from different perspectives:

%s

Your task:
1. Briefly analyze each response for accuracy, insight, and quality (especially how well they addressed the user's follow-up if applicable)
2. Provide your FINAL RANKING at the end

IMPORTANT: You MUST end your response with a ranking section in this EXACT format:

FINAL RANKING:
1. [Name of Agent 1]
2. [Name of Agent 2]

(List the agents from best to worst based on the quality of their response)

Your evaluation:

Use Markdown format.`, council.Topic, contextText.String(), formattedResponses)
}

func (e *Engine) buildSynthesisPrompt(council *core.Council, responses []core.Response, aggregateRanks []core.AggregateRanking) string {
	// Map member IDs to names
	memberNames := make(map[string]string)
	for _, m := range council.Members {
		memberNames[m.ID] = m.Name
	}

	// Format responses
	var responsesText strings.Builder
	for _, r := range responses {
		memberName := memberNames[r.MemberID]
		if r.MemberID == "user" {
			memberName = "User Follow-up"
		}
		responsesText.WriteString(fmt.Sprintf("\n[%s]\n%s\n", memberName, r.Content))
	}

	// Format rankings
	var rankingsText strings.Builder
	rankingsText.WriteString("\nAggregate Rankings (by quality):\n")
	for i, ar := range aggregateRanks {
		memberName := memberNames[ar.MemberID]
		rankingsText.WriteString(fmt.Sprintf("%d. %s - Avg rank: %.2f\n", i+1, memberName, ar.AvgRank))
	}

	var contextText strings.Builder
	if len(council.Syntheses) > 0 {
		latestSynthesis := council.Syntheses[len(council.Syntheses)-1]
		contextText.WriteString(fmt.Sprintf("\nYour Previous Conclusion:\n%s\n", latestSynthesis.Content))
	}

	return fmt.Sprintf(`You are the Chairman synthesizing a council discussion.

Topic: %s
%s
Individual Responses:
%s

%s

Your task as Chairman is to synthesize all of this information into a single, comprehensive, accurate answer to the user's original question. Consider:
- The individual responses and their insights
- The peer rankings and what they reveal about response quality
- Any patterns of agreement or disagreement
`, council.Topic, contextText.String(), responsesText.String(), rankingsText.String())
}

func (e *Engine) formatResponsesForRanking(responses []core.Response, members []core.Agent) string {
	memberMap := make(map[string]core.Agent)
	for _, m := range members {
		memberMap[m.ID] = m
	}

	var result strings.Builder
	for _, r := range responses {
		agent, ok := memberMap[r.MemberID]
		name := "Unknown"
		if ok {
			name = agent.Name
		}
		result.WriteString(fmt.Sprintf("\nResponse by %s:\n%s\n", name, r.Content))
	}

	return result.String()
}

func (e *Engine) parseRankingsFromText(text string, responses []core.Response, members []core.Agent) ([]string, error) {
	// Map member identifiers to response IDs
	idMap := make(map[string]string)
	memberMap := make(map[string]core.Agent)
	for _, m := range members {
		memberMap[m.ID] = m
	}

	for _, r := range responses {
		if agent, ok := memberMap[r.MemberID]; ok {
			// Map various ways an agent might be identified
			idMap[strings.ToLower(agent.Name)] = r.ID
			idMap[strings.ToLower(agent.Provider)] = r.ID
			idMap[strings.ToLower(agent.Persona)] = r.ID

			// Also map provider + persona format like "qwen (Pragmatist)"
			// This matches how agents are named in Engine.CreateCouncil
			// but we handle it case-insensitively.
		}
	}

	// Try to find a ranking section
	rankingSection := ""
	lines := strings.Split(text, "\n")
	inRanking := false

	rankingHeaders := []string{
		"FINAL RANKING",
		"RANKING:",
		"MY RANKING",
		"RANKED",
		"ORDER",
		"BEST TO WORST",
		"FROM BEST",
	}

	for _, line := range lines {
		upperLine := strings.ToUpper(line)
		headerFound := false
		for _, header := range rankingHeaders {
			if strings.Contains(upperLine, header) {
				inRanking = true
				headerFound = true
				break
			}
		}
		if inRanking && !headerFound {
			rankingSection += line + "\n"
		}
	}

	if rankingSection == "" {
		rankingSection = text
	}

	var rankedIDs []string
	seen := make(map[string]bool)

	// Go through lines in ranking section and find identifiers
	rankLines := strings.Split(rankingSection, "\n")
	for _, line := range rankLines {
		lineLower := strings.ToLower(line)
		if strings.TrimSpace(lineLower) == "" {
			continue
		}

		// Check for identifiers in this line
		// We want to find the BEST match in the line
		var foundID string
		var bestMatchLen int

		for identifier, id := range idMap {
			if seen[id] {
				continue
			}

			// Check if identifier is in line
			if strings.Contains(lineLower, identifier) {
				// Keep track of longest match to handle "qwen" vs "qwen (Pragmatist)"
				if len(identifier) > bestMatchLen {
					bestMatchLen = len(identifier)
					foundID = id
				}
			}
		}

		if foundID != "" {
			rankedIDs = append(rankedIDs, foundID)
			seen[foundID] = true
		}
	}

	// If no rankings found at all, it's an error in parsing
	if len(rankedIDs) == 0 {
		slog.Warn("Failed to parse any rankings from text", "text_preview", text[:min(len(text), 200)])
		// Use response order as fallback
		for _, r := range responses {
			rankedIDs = append(rankedIDs, r.ID)
		}
	} else if len(rankedIDs) < len(responses) {
		// Append unranked responses at the end
		for _, r := range responses {
			if !seen[r.ID] {
				rankedIDs = append(rankedIDs, r.ID)
			}
		}
	}

	return rankedIDs, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (e *Engine) calculateAggregateRankings(responses []core.Response, rankings []core.Ranking, members []core.Agent) []core.AggregateRanking {
	// Map response ID to member ID
	responseToMember := make(map[string]string)
	for _, r := range responses {
		responseToMember[r.ID] = r.MemberID
	}

	// Collect positions for each response
	positionsByResponse := make(map[string][]int)
	for _, ranking := range rankings {
		for pos, responseID := range ranking.Rankings {
			positionsByResponse[responseID] = append(positionsByResponse[responseID], pos+1) // 1-based position
		}
	}

	// Calculate average ranks
	aggregateRanks := make([]core.AggregateRanking, 0, len(responses))
	for _, r := range responses {
		positions := positionsByResponse[r.ID]
		if len(positions) == 0 {
			continue
		}

		sum := 0
		for _, pos := range positions {
			sum += pos
		}
		avgRank := float64(sum) / float64(len(positions))

		aggregateRanks = append(aggregateRanks, core.AggregateRanking{
			ResponseID: r.ID,
			MemberID:   r.MemberID,
			Positions:  positions,
			AvgRank:    avgRank,
		})
	}

	// Sort by average rank (lower is better)
	sort.Slice(aggregateRanks, func(i, j int) bool {
		return aggregateRanks[i].AvgRank < aggregateRanks[j].AvgRank
	})

	return aggregateRanks
}
