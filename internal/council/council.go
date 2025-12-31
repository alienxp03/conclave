// Package council implements the multi-agent council system.
package council

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/alienxp03/dbate/internal/core"
	"github.com/alienxp03/dbate/internal/persona"
	"github.com/alienxp03/dbate/internal/provider"
	"github.com/alienxp03/dbate/internal/storage"
)

// Engine orchestrates council sessions.
type Engine struct {
	storage  storage.Storage
	registry *provider.Registry
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
			ID:       uuid.New().String(),
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
			ID:       uuid.New().String(),
			Name:     fmt.Sprintf("Chairman (%s)", chairmanSpec.Provider),
			Provider: chairmanSpec.Provider,
			Model:    chairmanSpec.Model,
			Persona:  "chairman",
		}
	} else {
		// Use default chairman (first member's provider with best model)
		defaultChairman := core.GetDefaultChairman(members)
		chairman = core.Agent{
			ID:       uuid.New().String(),
			Name:     fmt.Sprintf("Chairman (%s)", defaultChairman.Provider),
			Provider: defaultChairman.Provider,
			Model:    defaultChairman.Model,
			Persona:  "chairman",
		}
	}

	// Create council
	now := time.Now()
	council := &core.Council{
		ID:        uuid.New().String(),
		Topic:     config.Topic,
		Members:   agents,
		Chairman:  chairman,
		Status:    core.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return council, nil
}

// RunCouncil executes all 3 stages of the council process.
func (e *Engine) RunCouncil(ctx context.Context, council *core.Council) error {
	// Save council first
	if err := e.storage.CreateCouncil(council); err != nil {
		return fmt.Errorf("failed to create council: %w", err)
	}

	// Update status
	council.Status = core.StatusInProgress
	council.UpdatedAt = time.Now()
	e.storage.UpdateCouncil(council)

	// Stage 1: Collect responses
	responses, err := e.CollectResponses(ctx, council)
	if err != nil {
		council.Status = core.StatusFailed
		e.storage.UpdateCouncil(council)
		return fmt.Errorf("stage 1 failed: %w", err)
	}

	// Save responses
	for _, r := range responses {
		if err := e.storage.AddResponse(&r); err != nil {
			return fmt.Errorf("failed to save response: %w", err)
		}
	}

	// Stage 2: Collect rankings
	rankings, err := e.CollectRankings(ctx, council, responses)
	if err != nil {
		council.Status = core.StatusFailed
		e.storage.UpdateCouncil(council)
		return fmt.Errorf("stage 2 failed: %w", err)
	}

	// Save rankings
	for _, r := range rankings {
		if err := e.storage.AddRanking(&r); err != nil {
			return fmt.Errorf("failed to save ranking: %w", err)
		}
	}

	// Stage 3: Generate synthesis
	synthesis, err := e.GenerateSynthesis(ctx, council, responses, rankings)
	if err != nil {
		council.Status = core.StatusFailed
		e.storage.UpdateCouncil(council)
		return fmt.Errorf("stage 3 failed: %w", err)
	}

	// Update council
	council.Synthesis = synthesis
	council.Status = core.StatusCompleted
	now := time.Now()
	council.CompletedAt = &now
	council.UpdatedAt = now

	if err := e.storage.UpdateCouncil(council); err != nil {
		return fmt.Errorf("failed to update council: %w", err)
	}

	return nil
}

// CollectResponses implements Stage 1: collect independent responses from all members.
func (e *Engine) CollectResponses(ctx context.Context, council *core.Council) ([]core.Response, error) {
	var wg sync.WaitGroup
	responseChan := make(chan core.Response, len(council.Members))
	errorChan := make(chan error, len(council.Members))

	for _, member := range council.Members {
		wg.Add(1)
		go func(agent core.Agent) {
			defer wg.Done()

			// Build prompt
			prompt, err := e.buildResponsePrompt(council.Topic, agent)
			if err != nil {
				errorChan <- fmt.Errorf("failed to build prompt for %s: %w", agent.Name, err)
				return
			}

			// Execute provider
			prov, err := e.registry.Get(agent.Provider)
			if err != nil {
				errorChan <- fmt.Errorf("provider not found for %s: %w", agent.Name, err)
				return
			}

			content, err := prov.GenerateWithModel(ctx, prompt, agent.Model)
			if err != nil {
				errorChan <- fmt.Errorf("generation failed for %s: %w", agent.Name, err)
				return
			}

			response := core.Response{
				ID:        uuid.New().String(),
				CouncilID: council.ID,
				MemberID:  agent.ID,
				Content:   content,
				CreatedAt: time.Now(),
			}

			responseChan <- response
		}(member)
	}

	// Wait for all goroutines
	wg.Wait()
	close(responseChan)
	close(errorChan)

	// Check for errors
	if len(errorChan) > 0 {
		err := <-errorChan
		return nil, err
	}

	// Collect all responses
	responses := make([]core.Response, 0, len(council.Members))
	for response := range responseChan {
		responses = append(responses, response)
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("no responses collected")
	}

	return responses, nil
}

// CollectRankings implements Stage 2: collect rankings from all members.
func (e *Engine) CollectRankings(ctx context.Context, council *core.Council, responses []core.Response) ([]core.Ranking, error) {
	// Anonymize responses for unbiased ranking
	anonymized := e.anonymizeResponses(responses)

	var wg sync.WaitGroup
	rankingChan := make(chan core.Ranking, len(council.Members))
	errorChan := make(chan error, len(council.Members))

	for _, member := range council.Members {
		wg.Add(1)
		go func(agent core.Agent) {
			defer wg.Done()

			// Build ranking prompt
			prompt := e.buildRankingPrompt(council.Topic, anonymized)

			// Execute provider
			prov, err := e.registry.Get(agent.Provider)
			if err != nil {
				errorChan <- fmt.Errorf("provider not found for %s: %w", agent.Name, err)
				return
			}

			content, err := prov.GenerateWithModel(ctx, prompt, agent.Model)
			if err != nil {
				errorChan <- fmt.Errorf("generation failed for %s: %w", agent.Name, err)
				return
			}

			// Parse rankings from response
			rankedIDs, err := e.parseRankingsFromText(content, responses)
			if err != nil {
				errorChan <- fmt.Errorf("failed to parse rankings for %s: %w", agent.Name, err)
				return
			}

			ranking := core.Ranking{
				ID:         uuid.New().String(),
				CouncilID:  council.ID,
				ReviewerID: agent.ID,
				Rankings:   rankedIDs,
				Reasoning:  content,
				CreatedAt:  time.Now(),
			}

			rankingChan <- ranking
		}(member)
	}

	// Wait for all goroutines
	wg.Wait()
	close(rankingChan)
	close(errorChan)

	// Check for errors
	if len(errorChan) > 0 {
		err := <-errorChan
		return nil, err
	}

	// Collect all rankings
	rankings := make([]core.Ranking, 0, len(council.Members))
	for ranking := range rankingChan {
		rankings = append(rankings, ranking)
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
	prompt := e.buildSynthesisPrompt(council.Topic, responses, aggregateRanks, council.Members)

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
	if p := persona.GetBuiltinPersona(name); p != nil {
		return p
	}

	// Check storage for custom personas
	customPersona, err := e.storage.GetPersona(name)
	if err == nil {
		return &persona.Persona{
			ID:           customPersona.ID,
			Name:         customPersona.Name,
			Description:  customPersona.Description,
			SystemPrompt: customPersona.SystemPrompt,
		}
	}

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

Your response:`, personaDef.SystemPrompt, topic)

	return prompt, nil
}

func (e *Engine) buildRankingPrompt(topic string, anonymizedResponses string) string {
	return fmt.Sprintf(`You are evaluating multiple responses to the following topic:

Topic: %s

Here are the anonymized responses:

%s

Analyze each response for accuracy, insight, and quality.
Provide detailed critique of each.

Then provide your FINAL RANKING at the end, listing responses from BEST to WORST.

Format your final ranking as:
FINAL RANKING:
1. Response A
2. Response B
3. Response C
...

Your evaluation:`, topic, anonymizedResponses)
}

func (e *Engine) buildSynthesisPrompt(topic string, responses []core.Response, aggregateRanks []core.AggregateRanking, members []core.Agent) string {
	// Map member IDs to names
	memberNames := make(map[string]string)
	for _, m := range members {
		memberNames[m.ID] = m.Name
	}

	// Format responses
	var responsesText strings.Builder
	for _, r := range responses {
		memberName := memberNames[r.MemberID]
		responsesText.WriteString(fmt.Sprintf("\n[%s]\n%s\n", memberName, r.Content))
	}

	// Format rankings
	var rankingsText strings.Builder
	rankingsText.WriteString("\nAggregate Rankings (by quality):\n")
	for i, ar := range aggregateRanks {
		memberName := memberNames[ar.MemberID]
		rankingsText.WriteString(fmt.Sprintf("%d. %s - Avg rank: %.2f\n", i+1, memberName, ar.AvgRank))
	}

	return fmt.Sprintf(`You are the Chairman synthesizing a council discussion.

Topic: %s

Individual Responses:
%s

%s

Your task:
1. Synthesize all perspectives and rankings into a coherent conclusion
2. Identify areas of agreement and disagreement
3. Highlight the strongest arguments (based on rankings)
4. Provide a balanced recommendation

Your synthesis:`, topic, responsesText.String(), rankingsText.String())
}

func (e *Engine) anonymizeResponses(responses []core.Response) string {
	labels := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	var result strings.Builder

	for i, r := range responses {
		if i < len(labels) {
			result.WriteString(fmt.Sprintf("\nResponse %s:\n%s\n", labels[i], r.Content))
		}
	}

	return result.String()
}

func (e *Engine) parseRankingsFromText(text string, responses []core.Response) ([]string, error) {
	// Find FINAL RANKING section
	rankingSection := ""
	lines := strings.Split(text, "\n")
	inRanking := false

	for _, line := range lines {
		if strings.Contains(strings.ToUpper(line), "FINAL RANKING") {
			inRanking = true
			continue
		}
		if inRanking {
			rankingSection += line + "\n"
		}
	}

	if rankingSection == "" {
		return nil, fmt.Errorf("no FINAL RANKING section found")
	}

	// Parse rankings: "1. Response A", "2. Response B", etc.
	labels := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	labelToID := make(map[string]string)
	for i, r := range responses {
		if i < len(labels) {
			labelToID[labels[i]] = r.ID
		}
	}

	// Extract ordered response labels
	re := regexp.MustCompile(`\d+\.\s+Response\s+([A-H])`)
	matches := re.FindAllStringSubmatch(rankingSection, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("no rankings found in expected format")
	}

	rankedIDs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			label := match[1]
			if id, ok := labelToID[label]; ok {
				rankedIDs = append(rankedIDs, id)
			}
		}
	}

	if len(rankedIDs) == 0 {
		return nil, fmt.Errorf("failed to parse any valid rankings")
	}

	return rankedIDs, nil
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
