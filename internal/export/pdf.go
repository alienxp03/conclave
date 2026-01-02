package export

import (
	"fmt"
	"io"
	"strings"

	"github.com/jung-kurt/gofpdf"

	"github.com/alienxp03/conclave/internal/core"
)

// PDFExporter exports debates to PDF format.
type PDFExporter struct{}

// Export writes the debate as PDF.
func (e *PDFExporter) Export(debate *core.Debate, turns []*core.Turn, w io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Add first page
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 18)
	pdf.MultiCell(0, 10, debate.Topic, "", "C", false)
	pdf.Ln(5)

	// Metadata section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Debate Information")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	e.addMetadataRow(pdf, "ID:", debate.ID[:8]+"...")
	e.addMetadataRow(pdf, "Style:", strings.Title(debate.Style))
	e.addMetadataRow(pdf, "Status:", string(debate.Status))
	e.addMetadataRow(pdf, "Created:", debate.CreatedAt.Format("January 2, 2006 at 3:04 PM"))
	if debate.CompletedAt != nil {
		e.addMetadataRow(pdf, "Completed:", debate.CompletedAt.Format("January 2, 2006 at 3:04 PM"))
		e.addMetadataRow(pdf, "Duration:", formatDuration(debate.CreatedAt, *debate.CompletedAt))
	}
	pdf.Ln(5)

	// Participants section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Participants")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	e.addParticipantBox(pdf, "Agent A", debate.AgentA, 200, 230, 255) // Light blue
	pdf.Ln(3)
	e.addParticipantBox(pdf, "Agent B", debate.AgentB, 200, 255, 200) // Light green
	pdf.Ln(8)

	// Debate content
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Debate")
	pdf.Ln(8)

	if len(turns) == 0 {
		pdf.SetFont("Arial", "I", 10)
		pdf.Cell(0, 6, "No turns recorded.")
		pdf.Ln(6)
	} else {
		for _, turn := range turns {
			agentName := debate.AgentA.Name
			isAgentA := turn.AgentID == debate.AgentA.ID
			if !isAgentA {
				agentName = debate.AgentB.Name
			}

			// Check if we need a new page
			if pdf.GetY() > 250 {
				pdf.AddPage()
			}

			// Turn header with colored background
			if isAgentA {
				pdf.SetFillColor(200, 230, 255) // Light blue
			} else {
				pdf.SetFillColor(200, 255, 200) // Light green
			}

			pdf.SetFont("Arial", "B", 10)
			header := fmt.Sprintf("Turn %d - %s (%s)", turn.Number, agentName, turn.CreatedAt.Format("3:04 PM"))
			pdf.CellFormat(0, 7, header, "", 1, "", true, 0, "")

			// Turn content
			pdf.SetFont("Arial", "", 9)
			pdf.SetFillColor(255, 255, 255)

			// Word wrap the content
			content := e.sanitizeText(turn.Content)
			pdf.MultiCell(0, 5, content, "", "", false)
			pdf.Ln(5)
		}
	}

	// Conclusion
	if debate.Conclusion != nil {
		if pdf.GetY() > 230 {
			pdf.AddPage()
		}

		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "Conclusion")
		pdf.Ln(8)

		// Consensus status
		if debate.Conclusion.Agreed {
			pdf.SetFillColor(200, 255, 200) // Light green
			pdf.SetFont("Arial", "B", 10)
			pdf.CellFormat(0, 7, "Consensus Reached", "", 1, "", true, 0, "")
		} else {
			pdf.SetFillColor(255, 200, 200) // Light red
			pdf.SetFont("Arial", "B", 10)
			pdf.CellFormat(0, 7, "No Consensus", "", 1, "", true, 0, "")
		}

		pdf.SetFont("Arial", "", 10)
		pdf.SetFillColor(255, 255, 255)
		pdf.MultiCell(0, 5, e.sanitizeText(debate.Conclusion.Summary), "", "", false)
		pdf.Ln(3)

		if !debate.Conclusion.Agreed {
			if debate.Conclusion.AgentASummary != "" {
				pdf.SetFont("Arial", "B", 10)
				pdf.Cell(0, 6, fmt.Sprintf("%s's Position:", debate.AgentA.Name))
				pdf.Ln(6)
				pdf.SetFont("Arial", "", 9)
				pdf.MultiCell(0, 5, e.sanitizeText(debate.Conclusion.AgentASummary), "", "", false)
				pdf.Ln(3)
			}
			if debate.Conclusion.AgentBSummary != "" {
				pdf.SetFont("Arial", "B", 10)
				pdf.Cell(0, 6, fmt.Sprintf("%s's Position:", debate.AgentB.Name))
				pdf.Ln(6)
				pdf.SetFont("Arial", "", 9)
				pdf.MultiCell(0, 5, e.sanitizeText(debate.Conclusion.AgentBSummary), "", "", false)
				pdf.Ln(3)
			}
		}
	}

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 10, "Exported from conclave", "", 0, "C", false, 0, "")

	return pdf.Output(w)
}

// FileExtension returns the file extension for PDF.
func (e *PDFExporter) FileExtension() string {
	return "pdf"
}

// Helper to add a metadata row
func (e *PDFExporter) addMetadataRow(pdf *gofpdf.Fpdf, label, value string) {
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(30, 5, label)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, value)
	pdf.Ln(5)
}

// Helper to add a participant box
func (e *PDFExporter) addParticipantBox(pdf *gofpdf.Fpdf, title string, agent core.Agent, r, g, b int) {
	pdf.SetFillColor(r, g, b)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, title, "", 1, "", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)
	pdf.Cell(25, 5, "Name:")
	pdf.Cell(0, 5, agent.Name)
	pdf.Ln(5)
	pdf.Cell(25, 5, "Provider:")
	pdf.Cell(0, 5, agent.Provider)
	pdf.Ln(5)
	pdf.Cell(25, 5, "Persona:")
	pdf.Cell(0, 5, agent.Persona)
	pdf.Ln(5)
}

// Sanitize text for PDF (remove problematic characters)
func (e *PDFExporter) sanitizeText(text string) string {
	// gofpdf uses Windows-1252 encoding by default
	// Replace common Unicode characters that might cause issues
	replacer := strings.NewReplacer(
		"\u2018", "'",  // Left single quote
		"\u2019", "'",  // Right single quote
		"\u201C", "\"", // Left double quote
		"\u201D", "\"", // Right double quote
		"\u2013", "-",  // En dash
		"\u2014", "--", // Em dash
		"\u2026", "...", // Ellipsis
		"\u2022", "*",  // Bullet
		"\u00A0", " ",  // Non-breaking space
	)
	return replacer.Replace(text)
}
