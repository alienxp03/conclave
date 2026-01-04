package opencode

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/alienxp03/conclave/provider"
)

// JSONEvent represents an event in the Opencode CLI JSON lines output.
type JSONEvent struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionID"`
	Part      *struct {
		Type   string `json:"type"`
		Text   string `json:"text,omitempty"`
		Reason string `json:"reason,omitempty"`
		Tokens *struct {
			Input  int `json:"input"`
			Output int `json:"output"`
		} `json:"tokens,omitempty"`
	} `json:"part,omitempty"`
}

// ParseJSON parses Opencode CLI JSON lines output.
func ParseJSON(data string, duration time.Duration) (*provider.Response, error) {
	resp := &provider.Response{Raw: data}
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event JSONEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			slog.Debug("Failed to unmarshal Opencode JSON line", "line", line, "error", err)
			continue
		}

		if event.Type == "text" && event.Part != nil {
			resp.Content += event.Part.Text
		}

		if event.Type == "step_finish" && event.Part != nil {
			if resp.Metadata == nil {
				resp.Metadata = &provider.Metadata{Duration: duration}
			}
			resp.Metadata.StopReason = event.Part.Reason
			resp.Metadata.SessionID = event.SessionID
			if event.Part.Tokens != nil {
				resp.Metadata.InputTokens = event.Part.Tokens.Input
				resp.Metadata.OutputTokens = event.Part.Tokens.Output
				resp.Metadata.TotalTokens = event.Part.Tokens.Input + event.Part.Tokens.Output
			}
		}
	}

	if resp.Content == "" {
		slog.Debug("No text content found in Opencode output, using raw output as fallback")
		// Fallback if no text events found or parsing failed
		return &provider.Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	return resp, nil
}
