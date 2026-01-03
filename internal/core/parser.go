package core

import (
	"fmt"
	"strings"
)

// ParseMemberSpec parses a member specification string.
// Format: provider[/model][:persona]
//
// Examples:
//   - "claude" -> {Provider: "claude", Model: "", Persona: ""}
//   - "claude:optimist" -> {Provider: "claude", Model: "", Persona: "optimist"}
//   - "claude/opus" -> {Provider: "claude", Model: "opus", Persona: ""}
//   - "claude/opus:optimist" -> {Provider: "claude", Model: "opus", Persona: "optimist"}
func ParseMemberSpec(spec string) (MemberSpec, error) {
	if spec == "" {
		return MemberSpec{}, fmt.Errorf("member spec cannot be empty")
	}

	var m MemberSpec

	// Split by ':' to separate provider[/model] from persona
	parts := strings.SplitN(spec, ":", 2)

	// parts[0] = provider[/model]
	// parts[1] = persona (if exists)

	// Split provider/model
	providerParts := strings.SplitN(parts[0], "/", 2)
	m.Provider = strings.TrimSpace(providerParts[0])

	if m.Provider == "" {
		return MemberSpec{}, fmt.Errorf("provider cannot be empty in spec: %s", spec)
	}

	// Extract model if specified
	if len(providerParts) == 2 {
		m.Model = strings.TrimSpace(providerParts[1])
	}

	// Extract persona if specified
	if len(parts) == 2 {
		m.Persona = strings.TrimSpace(parts[1])
	}

	return m, nil
}

// ParseMemberSpecs parses a comma-separated list of member specifications.
// Format: spec1,spec2,spec3,...
func ParseMemberSpecs(specsStr string) ([]MemberSpec, error) {
	if specsStr == "" {
		return nil, fmt.Errorf("member specs cannot be empty")
	}

	specs := strings.Split(specsStr, ",")
	members := make([]MemberSpec, 0, len(specs))

	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}

		member, err := ParseMemberSpec(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid member spec '%s': %w", spec, err)
		}

		members = append(members, member)
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("no valid member specs found")
	}

	return members, nil
}

// DefaultPersonaOrder defines the order in which personas are auto-assigned.
var DefaultPersonaOrder = []string{
	"optimist",
	"skeptic",
	"pragmatist",
	"analyst",
	"visionary",
	"devils_advocate",
}

// AssignDefaultPersonas assigns default personas to member specs that don't have one.
// Personas are assigned in order from DefaultPersonaOrder.
func AssignDefaultPersonas(members []MemberSpec) []MemberSpec {
	result := make([]MemberSpec, len(members))
	copy(result, members)

	for i := range result {
		if result[i].Persona == "" {
			// Assign persona from default order (cycle if more members than personas)
			personaIdx := i % len(DefaultPersonaOrder)
			result[i].Persona = DefaultPersonaOrder[personaIdx]
		}
	}

	return result
}

// DefaultCommandForProvider returns the default command for a provider.
var DefaultCommandForProvider = map[string]string{
	"claude":   "claude",
	"gemini":   "gemini",
	"qwen":     "qwen",
	"codex":    "codex",
	"opencode": "opencode",
	"mock":     "",
}

// DefaultArgsForProvider returns the default arguments for a provider.
var DefaultArgsForProvider = map[string][]string{
	"claude":   {"--print"},
	"gemini":   {},
	"qwen":     {},
	"codex":    {},
	"opencode": {},
	"mock":     {},
}

// DefaultModelsForProvider returns the list of supported models for a provider.
var DefaultModelsForProvider = map[string][]string{
	"claude":   {"opus-4.5", "sonnet-4.5", "haiku-4.5"},
	"gemini":   {"gemini-3-pro-preview", "gemini-3-flash-preview"},
	"qwen":     {"qwen-3-coder-plus"},
	"codex":    {"gpt-5.2-codex", "gpt-5.2"},
	"opencode": {"zai-coding-plan/glm-4.7", "google/gemini-3-flash-preview"},
	"mock":     {"mock-v1", "mock-v2"},
}

// DefaultModelForProvider returns the default model for a provider.
var DefaultModelForProvider = map[string]string{
	"claude":   "sonnet-4.5",
	"gemini":   "gemini-3-flash-preview",
	"qwen":     "qwen-3-coder-plus",
	"codex":    "gpt-5.2-codex",
	"opencode": "zai-coding-plan/glm-4.7",
	"mock":     "mock-v1",
}

// BestModelForProvider returns the best (most capable) model for a provider.
// Used for chairman selection.
var BestModelForProvider = map[string]string{
	"claude":   "opus",
	"gemini":   "pro",
	"qwen":     "max",
	"codex":    "gpt-5.2-codex",
	"opencode": "zai-coding-plan/glm-4.7",
	"mock":     "mock-v1",
}

// AssignDefaultModels assigns default models to member specs that don't have one.
func AssignDefaultModels(members []MemberSpec) []MemberSpec {
	result := make([]MemberSpec, len(members))
	copy(result, members)

	for i := range result {
		if result[i].Model == "" {
			if defaultModel, ok := DefaultModelForProvider[result[i].Provider]; ok {
				result[i].Model = defaultModel
			}
		}
	}

	return result
}

// GetDefaultChairman returns a default chairman based on the first member's provider.
// Upgrades to the best model for that provider.
func GetDefaultChairman(members []MemberSpec) MemberSpec {
	if len(members) == 0 {
		// Ultimate fallback
		return MemberSpec{
			Provider: "claude",
			Model:    "opus",
			Persona:  "chairman",
		}
	}

	provider := members[0].Provider
	bestModel := BestModelForProvider[provider]
	if bestModel == "" {
		bestModel = DefaultModelForProvider[provider]
	}

	return MemberSpec{
		Provider: provider,
		Model:    bestModel,
		Persona:  "chairman",
	}
}
