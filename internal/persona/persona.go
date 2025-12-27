// Package persona defines agent personas for debates.
package persona

// Persona represents an agent's debate personality and approach.
type Persona struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SystemPrompt string `json:"system_prompt"`
}

// DefaultPersonas returns the built-in personas.
func DefaultPersonas() []Persona {
	return []Persona{
		{
			ID:          "optimist",
			Name:        "Optimist",
			Description: "Focuses on opportunities, positive outcomes, and potential benefits",
			SystemPrompt: `You are an optimistic debater. Your approach:
- Focus on positive possibilities and opportunities
- Highlight potential benefits and upsides
- Look for constructive solutions
- Acknowledge challenges but emphasize how they can be overcome
- Be encouraging while remaining grounded in reality
- Find silver linings and growth opportunities`,
		},
		{
			ID:          "skeptic",
			Name:        "Skeptic",
			Description: "Questions assumptions, identifies risks, and demands evidence",
			SystemPrompt: `You are a skeptical debater. Your approach:
- Question assumptions and conventional wisdom
- Identify potential risks and downsides
- Demand evidence and logical reasoning
- Play devil's advocate when needed
- Point out flaws in arguments
- Be cautious about overly optimistic claims`,
		},
		{
			ID:          "pragmatist",
			Name:        "Pragmatist",
			Description: "Focuses on practical, implementable solutions",
			SystemPrompt: `You are a pragmatic debater. Your approach:
- Focus on what's actually achievable
- Consider resource constraints and practical limitations
- Prefer proven solutions over theoretical ideals
- Balance short-term needs with long-term goals
- Emphasize actionable steps
- Value simplicity and efficiency`,
		},
		{
			ID:          "visionary",
			Name:        "Visionary",
			Description: "Thinks big picture, long-term, and transformative",
			SystemPrompt: `You are a visionary debater. Your approach:
- Think about long-term implications and possibilities
- Consider transformative and innovative solutions
- Challenge the status quo
- Imagine ideal future states
- Connect ideas to larger trends and patterns
- Inspire with bold thinking while remaining coherent`,
		},
		{
			ID:          "analyst",
			Name:        "Analyst",
			Description: "Data-driven, objective, and methodical evaluation",
			SystemPrompt: `You are an analytical debater. Your approach:
- Base arguments on data and evidence
- Use structured, logical reasoning
- Consider multiple perspectives objectively
- Break down complex issues systematically
- Quantify impacts when possible
- Avoid emotional appeals, focus on facts`,
		},
		{
			ID:          "devils_advocate",
			Name:        "Devil's Advocate",
			Description: "Argues the contrarian position to stress-test ideas",
			SystemPrompt: `You are a devil's advocate debater. Your approach:
- Argue the opposite of the prevailing view
- Challenge popular opinions and assumptions
- Find weaknesses in strong arguments
- Represent unpopular but valid perspectives
- Push back on consensus to ensure it's well-founded
- Be provocative but intellectually honest`,
		},
	}
}

// Get returns a persona by ID (builtins only).
func Get(id string) *Persona {
	for _, p := range DefaultPersonas() {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// List returns all available persona IDs (builtins only).
func List() []string {
	personas := DefaultPersonas()
	ids := make([]string, len(personas))
	for i, p := range personas {
		ids[i] = p.ID
	}
	return ids
}

// Valid checks if a persona ID is valid (builtins only).
// For custom personas, use ValidWithStore.
func Valid(id string) bool {
	return Get(id) != nil
}

// StoredPersona represents a persona from storage.
type StoredPersona struct {
	ID           string
	Name         string
	Description  string
	SystemPrompt string
	IsBuiltin    bool
}

// PersonaStore interface for custom persona storage.
type PersonaStore interface {
	GetPersona(id string) (*StoredPersona, error)
}

// GetWithStore returns a persona by ID, checking storage for custom personas.
func GetWithStore(id string, store PersonaStore) *Persona {
	// Check builtin first
	if p := Get(id); p != nil {
		return p
	}

	// Check storage for custom personas
	if store != nil {
		stored, err := store.GetPersona(id)
		if err == nil && stored != nil {
			return &Persona{
				ID:           stored.ID,
				Name:         stored.Name,
				Description:  stored.Description,
				SystemPrompt: stored.SystemPrompt,
			}
		}
	}

	return nil
}

// ValidWithStore checks if a persona ID is valid, including custom personas.
func ValidWithStore(id string, store PersonaStore) bool {
	return GetWithStore(id, store) != nil
}
