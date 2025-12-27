package export

import (
	"encoding/json"
	"io"

	"github.com/alienxp03/dbate/internal/core"
)

// JSONExporter exports debates to JSON format.
type JSONExporter struct{}

// ExportData represents the full export structure.
type ExportData struct {
	Debate *core.Debate  `json:"debate"`
	Turns  []*core.Turn  `json:"turns"`
}

// Export writes the debate as JSON.
func (e *JSONExporter) Export(debate *core.Debate, turns []*core.Turn, w io.Writer) error {
	data := ExportData{
		Debate: debate,
		Turns:  turns,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// FileExtension returns the file extension for JSON.
func (e *JSONExporter) FileExtension() string {
	return "json"
}
