package export

import (
	"fmt"
	"io"
	"strings"

	"github.com/alienxp03/dbate/internal/core"
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
		for _, turn := range turns {
			agentName := debate.AgentA.Name
			if turn.AgentID == debate.AgentB.ID {
				agentName = debate.AgentB.Name
			}

			sb.WriteString(fmt.Sprintf("### Turn %d - %s\n\n", turn.Number, agentName))
			sb.WriteString(fmt.Sprintf("*%s*\n\n", turn.CreatedAt.Format("3:04 PM")))
			sb.WriteString(turn.Content)
			sb.WriteString("\n\n---\n\n")
		}
	}

	// Conclusion
	if debate.Conclusion != nil {
		sb.WriteString("## Conclusion\n\n")

		if debate.Conclusion.Agreed {
			sb.WriteString("**✅ Consensus Reached**\n\n")
		} else {
			sb.WriteString("**❌ No Consensus**\n\n")
		}

		sb.WriteString(debate.Conclusion.Summary)
		sb.WriteString("\n\n")

		if !debate.Conclusion.Agreed {
			if debate.Conclusion.AgentASummary != "" {
				sb.WriteString(fmt.Sprintf("### %s's Final Position\n\n", debate.AgentA.Name))
				sb.WriteString(debate.Conclusion.AgentASummary)
				sb.WriteString("\n\n")
			}
			if debate.Conclusion.AgentBSummary != "" {
				sb.WriteString(fmt.Sprintf("### %s's Final Position\n\n", debate.AgentB.Name))
				sb.WriteString(debate.Conclusion.AgentBSummary)
				sb.WriteString("\n\n")
			}
		}
	}

	// Footer
	sb.WriteString("---\n\n")
	sb.WriteString("*Exported from dbate*\n")

	_, err := w.Write([]byte(sb.String()))
	return err
}

// FileExtension returns the file extension for Markdown.
func (e *MarkdownExporter) FileExtension() string {
	return "md"
}
