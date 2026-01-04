package qwen

import (
	"encoding/json"
	"time"

	"github.com/alienxp03/conclave/provider"
)

// JSONResponse represents legacy Qwen CLI JSON output.
type JSONResponse struct {
	Output struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"output,omitempty"`
	Usage *struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens,omitempty"`
	} `json:"usage,omitempty"`
	Text string `json:"text,omitempty"` // For simpler responses
}

// Event represents an event in the newer Qwen CLI JSON array output.
type Event struct {
	Type    string `json:"type"`
	Result  string `json:"result,omitempty"`
	Message *struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"message,omitempty"`
	Usage *struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// ParseJSON parses Qwen CLI JSON output.
func ParseJSON(data string, duration time.Duration) (*provider.Response, error) {
	// Try parsing as array of events first (newer format)
	var events []Event
	if err := json.Unmarshal([]byte(data), &events); err == nil && len(events) > 0 {
		resp := &provider.Response{Raw: data}

		var resultText string
		var assistantText string
		var usage *provider.Metadata

		// Scan all events to collect information
		for _, event := range events {
			if event.Type == "result" && event.Result != "" {
				resultText = event.Result
				if event.Usage != nil {
					usage = &provider.Metadata{
						InputTokens:  event.Usage.InputTokens,
						OutputTokens: event.Usage.OutputTokens,
						TotalTokens:  event.Usage.TotalTokens,
						Duration:     duration,
					}
				}
			}

			if event.Type == "assistant" && event.Message != nil {
				for _, c := range event.Message.Content {
					if c.Type == "text" {
						assistantText += c.Text
					}
				}
				if usage == nil && event.Usage != nil {
					usage = &provider.Metadata{
						InputTokens:  event.Usage.InputTokens,
						OutputTokens: event.Usage.OutputTokens,
						TotalTokens:  event.Usage.TotalTokens,
						Duration:     duration,
					}
				}
			}
		}

		// Prioritize resultText, then assistantText
		if resultText != "" {
			resp.Content = resultText
		} else {
			resp.Content = assistantText
		}

		resp.Metadata = usage

		if resp.Content != "" {
			return resp, nil
		}
	}

	// Fallback to legacy object format
	var raw JSONResponse
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return &provider.Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	resp := &provider.Response{
		Raw: data,
	}

	// Extract content
	if raw.Output.Text != "" {
		resp.Content = raw.Output.Text
		resp.Metadata = &provider.Metadata{
			StopReason: raw.Output.FinishReason,
			Duration:   duration,
		}
	} else if raw.Text != "" {
		resp.Content = raw.Text
	}

	// Extract metadata
	if raw.Usage != nil {
		if resp.Metadata == nil {
			resp.Metadata = &provider.Metadata{Duration: duration}
		}
		resp.Metadata.InputTokens = raw.Usage.InputTokens
		resp.Metadata.OutputTokens = raw.Usage.OutputTokens
		if raw.Usage.TotalTokens > 0 {
			resp.Metadata.TotalTokens = raw.Usage.TotalTokens
		} else {
			resp.Metadata.TotalTokens = raw.Usage.InputTokens + raw.Usage.OutputTokens
		}
	}

	return resp, nil
}
