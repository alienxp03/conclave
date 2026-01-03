package export

import (
	"fmt"
	"io"
	"strings"

	"github.com/alienxp03/conclave/internal/core"
)

// MarkdownExporter exports debates to Markdown format.
type MarkdownExporter struct{}

// Export writes the debate as Markdown.
func (e *MarkdownExporter) Export(debate *core.Debate, turns []*core.Turn, w io.Writer) error {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", debate.Topic))

	// Metadata
	sb.WriteString("## Debate Information\n\n")
	sb.WriteString(fmt.Sprintf("- **ID:** `%s`\n", debate.ID))
	sb.WriteString(fmt.Sprintf("- **Style:** %s\n", debate.Style))
	sb.WriteString(fmt.Sprintf("- **Status:** %s\n", debate.Status))
	sb.WriteString(fmt.Sprintf("- **Created:** %s\n", debate.CreatedAt.Format("January 2, 2006 at 3:04 PM")))
	if debate.CompletedAt != nil {
		sb.WriteString(fmt.Sprintf("- **Completed:** %s\n", debate.CompletedAt.Format("January 2, 2006 at 3:04 PM")))
		sb.WriteString(fmt.Sprintf("- **Duration:** %s\n", formatDuration(debate.CreatedAt, *debate.CompletedAt)))
	}
	sb.WriteString("\n")

	// Participants
	sb.WriteString("## Participants\n\n")
	sb.WriteString(fmt.Sprintf("### Agent A\n"))
	sb.WriteString(fmt.Sprintf("- **Name:** %s\n", debate.AgentA.Name))
	sb.WriteString(fmt.Sprintf("- **Provider:** %s\n", debate.AgentA.Provider))
	sb.WriteString(fmt.Sprintf("- **Persona:** %s\n", debate.AgentA.Persona))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("### Agent B\n"))
	sb.WriteString(fmt.Sprintf("- **Name:** %s\n", debate.AgentB.Name))
	sb.WriteString(fmt.Sprintf("- **Provider:** %s\n", debate.AgentB.Provider))
	sb.WriteString(fmt.Sprintf("- **Persona:** %s\n", debate.AgentB.Persona))
	sb.WriteString("\n")

	// Debate Content
	sb.WriteString("## Debate\n\n")

	if len(turns) == 0 {
		sb.WriteString("*No turns recorded.*\n\n")
	} else {
		// Group turns by round
		rounds := make(map[int][]*core.Turn)
		maxRound := 0
		for _, t := range turns {
			rounds[t.Round] = append(rounds[t.Round], t)
			if t.Round > maxRound {
				maxRound = t.Round
			}
		}

		// Group conclusions by round
		conclusions := make(map[int]*core.Conclusion)
		for _, c := range debate.Conclusions {
			conclusions[c.Round] = c
		}

		for r := 1; r <= maxRound; r++ {
			if r > 1 || len(rounds) > 1 {
				sb.WriteString(fmt.Sprintf("### Round %d\n\n", r))
			}

			for _, turn := range rounds[r] {
				agentName := debate.AgentA.Name
				if turn.AgentID == debate.AgentB.ID {
					agentName = debate.AgentB.Name
				} else if turn.AgentID == "user" {
					agentName = "User (Follow-up)"
				}

				sb.WriteString(fmt.Sprintf("#### Turn %d - %s\n\n", turn.Number, agentName))
				sb.WriteString(fmt.Sprintf("*%s*\n\n", turn.CreatedAt.Format("3:04 PM")))
				sb.WriteString(turn.Content)
				sb.WriteString("\n\n---\n\n")
			}

			// Add conclusion for this round if exists
			if c, ok := conclusions[r]; ok {
				sb.WriteString(fmt.Sprintf("### Round %d Conclusion\n\n", r))

				if c.Agreed {
					sb.WriteString("**✅ Consensus Reached**\n\n")
				} else {
					sb.WriteString("**❌ No Consensus**\n\n")
				}

				sb.WriteString(c.Summary)
				sb.WriteString("\n\n")

				if !c.Agreed {
					if c.AgentASummary != "" {
						sb.WriteString(fmt.Sprintf("#### %s's Position\n\n", debate.AgentA.Name))
						sb.WriteString(c.AgentASummary)
						sb.WriteString("\n\n")
					}
					if c.AgentBSummary != "" {
						sb.WriteString(fmt.Sprintf("#### %s's Position\n\n", debate.AgentB.Name))
						sb.WriteString(c.AgentBSummary)
						sb.WriteString("\n\n")
					}
				}
				sb.WriteString("\n---\n\n")
			}
		}
	}

	// Footer
	sb.WriteString("---\n\n")
	sb.WriteString("*Exported from conclave*\n")

	_, err := w.Write([]byte(sb.String()))
	return err
}

// FileExtension returns the file extension for Markdown.
func (e *MarkdownExporter) FileExtension() string {
	return "md"
}
