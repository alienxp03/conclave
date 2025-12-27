// Package style defines debate styles and formats.
package style

// Style represents a debate format/approach.
type Style struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	OpeningPrompt   string `json:"opening_prompt"`
	ResponsePrompt  string `json:"response_prompt"`
	ConclusionPrompt string `json:"conclusion_prompt"`
}

// DefaultStyles returns the built-in debate styles.
func DefaultStyles() []Style {
	return []Style{
		{
			ID:          "adversarial",
			Name:        "Adversarial",
			Description: "Agents argue opposite sides of the topic",
			OpeningPrompt: `You are participating in an adversarial debate on the topic: "{{.Topic}}"

You must argue FOR your assigned position. Present your strongest opening argument.
Be persuasive but fair. Anticipate counterarguments.

Your opening statement (2-3 paragraphs):`,
			ResponsePrompt: `You are in an adversarial debate on: "{{.Topic}}"

Your opponent just argued:
---
{{.PreviousArgument}}
---

Counter their points and strengthen your position. Find weaknesses in their argument.
Stay focused on winning the debate while remaining intellectually honest.

Your response:`,
			ConclusionPrompt: `The debate on "{{.Topic}}" is ending. This is your final statement.

Review the full debate:
{{.DebateHistory}}

Provide your final conclusion. State whether you've found any merit in your opponent's arguments.
If you can agree on common ground, state it clearly. If not, summarize your final position.

Your conclusion:`,
		},
		{
			ID:          "collaborative",
			Name:        "Collaborative",
			Description: "Agents work together to find the best solution",
			OpeningPrompt: `You are participating in a collaborative discussion on: "{{.Topic}}"

Your goal is to work WITH your partner to find the best possible solution or understanding.
Share your initial thoughts and perspective. Be open to building on each other's ideas.

Your opening thoughts (2-3 paragraphs):`,
			ResponsePrompt: `You are in a collaborative discussion on: "{{.Topic}}"

Your partner shared:
---
{{.PreviousArgument}}
---

Build on their ideas. Add your perspective. Identify areas of agreement and explore differences constructively.
Aim to synthesize viewpoints into something better than either alone.

Your response:`,
			ConclusionPrompt: `The collaborative discussion on "{{.Topic}}" is concluding.

Review the full discussion:
{{.DebateHistory}}

Synthesize the key insights from both perspectives. What solution or understanding have you reached together?
Highlight the most valuable contributions from each side.

Your conclusion:`,
		},
		{
			ID:          "analytical",
			Name:        "Analytical",
			Description: "Systematic pros/cons evaluation",
			OpeningPrompt: `You are conducting an analytical evaluation of: "{{.Topic}}"

Provide a structured analysis from your perspective:
1. Key considerations
2. Potential benefits (pros)
3. Potential drawbacks (cons)
4. Important tradeoffs

Be objective and thorough:`,
			ResponsePrompt: `You are in an analytical discussion on: "{{.Topic}}"

Your partner's analysis:
---
{{.PreviousArgument}}
---

Review their analysis. Add perspectives they may have missed.
Challenge any assumptions you find questionable. Refine the evaluation.

Your analysis:`,
			ConclusionPrompt: `The analytical evaluation of "{{.Topic}}" is complete.

Full analysis history:
{{.DebateHistory}}

Provide a final summary:
1. Key findings both analysts agree on
2. Points of disagreement and why
3. Recommended approach or conclusion
4. Confidence level in the analysis

Your summary:`,
		},
		{
			ID:          "socratic",
			Name:        "Socratic",
			Description: "One agent probes with questions, the other defends",
			OpeningPrompt: `You are engaging in a Socratic dialogue on: "{{.Topic}}"

{{if .IsQuestioner}}
You are the QUESTIONER. Ask probing questions to explore the topic deeply.
Challenge assumptions. Seek clarity. Help uncover hidden complexities.

Your opening questions:
{{else}}
You are the RESPONDER. Present your position on the topic clearly.
Be prepared to defend and refine your thinking through questioning.

Your opening position:
{{end}}`,
			ResponsePrompt: `You are in a Socratic dialogue on: "{{.Topic}}"

{{if .IsQuestioner}}
The responder said:
---
{{.PreviousArgument}}
---

Ask follow-up questions that probe deeper. Challenge their reasoning.
Help them (and yourself) reach greater understanding.

Your questions:
{{else}}
The questioner asked:
---
{{.PreviousArgument}}
---

Answer thoughtfully. Refine your position if the questions reveal weaknesses.
Be honest about uncertainties.

Your response:
{{end}}`,
			ConclusionPrompt: `The Socratic dialogue on "{{.Topic}}" is concluding.

Full dialogue:
{{.DebateHistory}}

What insights emerged from this inquiry? What was clarified?
What questions remain open?

Your reflection:`,
		},
	}
}

// Get returns a style by ID.
func Get(id string) *Style {
	for _, s := range DefaultStyles() {
		if s.ID == id {
			return &s
		}
	}
	return nil
}

// List returns all available style IDs.
func List() []string {
	styles := DefaultStyles()
	ids := make([]string, len(styles))
	for i, s := range styles {
		ids[i] = s.ID
	}
	return ids
}

// Valid checks if a style ID is valid.
func Valid(id string) bool {
	return Get(id) != nil
}

// Default returns the default debate style.
func Default() *Style {
	return Get("collaborative")
}
