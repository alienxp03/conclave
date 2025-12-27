// Package export handles exporting debates to various formats.
package export

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alienxp03/dbate/internal/core"
)

// Format represents an export format.
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatPDF      Format = "pdf"
	FormatJSON     Format = "json"
)

// Exporter defines the interface for exporting debates.
type Exporter interface {
	Export(debate *core.Debate, turns []*core.Turn, w io.Writer) error
	FileExtension() string
}

// GetExporter returns an exporter for the given format.
func GetExporter(format Format) (Exporter, error) {
	switch format {
	case FormatMarkdown:
		return &MarkdownExporter{}, nil
	case FormatPDF:
		return &PDFExporter{}, nil
	case FormatJSON:
		return &JSONExporter{}, nil
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// GenerateFilename creates a filename for the export.
func GenerateFilename(debate *core.Debate, ext string) string {
	// Sanitize topic for filename
	topic := debate.Topic
	if len(topic) > 50 {
		topic = topic[:50]
	}

	// Replace unsafe characters
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
	)
	topic = replacer.Replace(topic)

	timestamp := debate.CreatedAt.Format("20060102")
	return fmt.Sprintf("debate_%s_%s.%s", timestamp, topic, ext)
}

// Helper to format agent name
func formatAgentName(agent core.Agent) string {
	return fmt.Sprintf("%s (%s/%s)", agent.Name, agent.Provider, agent.Persona)
}

// Helper to format duration
func formatDuration(start, end time.Time) string {
	d := end.Sub(start)
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	return fmt.Sprintf("%.1f hours", d.Hours())
}
