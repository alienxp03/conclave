package openai

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/alienxp03/conclave/provider"
)

// JSONResponse represents OpenAI/Codex CLI structured output (legacy schema mode).
type JSONResponse struct {
	Response string `json:"response,omitempty"`
	// OpenAI compatible fields
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices,omitempty"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
	Content string `json:"content,omitempty"` // Fallback for simple content field
}

// JSONEvent represents a streaming event from Codex CLI --json output.
type JSONEvent struct {
	Type      string `json:"type"`
	ThreadID  string `json:"thread_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	// Message event fields (legacy schema)
	Message *struct {
		Role    string `json:"role,omitempty"`
		Content string `json:"content,omitempty"`
	} `json:"message,omitempty"`
	// Item event fields (current Codex JSONL schema)
	Item *struct {
		ID      string `json:"id,omitempty"`
		Type    string `json:"type,omitempty"`
		Text    string `json:"text,omitempty"`
		Content string `json:"content,omitempty"`
		Role    string `json:"role,omitempty"`
	} `json:"item,omitempty"`
	// Completion/finish event fields
	Usage *struct {
		PromptTokens     int   `json:"prompt_tokens,omitempty"`
		CompletionTokens int   `json:"completion_tokens,omitempty"`
		TotalTokens      int   `json:"total_tokens,omitempty"`
		DurationMs       int64 `json:"duration_ms,omitempty"`
		InputTokens      int   `json:"input_tokens,omitempty"`
		OutputTokens     int   `json:"output_tokens,omitempty"`
		CachedInput      int   `json:"cached_input_tokens,omitempty"`
	} `json:"usage,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
	// Text content for streaming
	Text string `json:"text,omitempty"`
}

// ParseJSON parses OpenAI/Codex CLI JSON output (supports both streaming and structured formats).
func ParseJSON(data string, duration time.Duration) (*provider.Response, error) {
	resp := &provider.Response{Raw: data}

	// Try parsing as newline-delimited JSON events first (streaming format)
	lines := strings.Split(data, "\n")
	var foundEvents bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event JSONEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		foundEvents = true

		// Extract text content from message or text field
		if event.Message != nil && event.Message.Content != "" {
			resp.Content += event.Message.Content
		}
		if event.Item != nil {
			if event.Item.Role == "assistant" ||
				event.Item.Type == "agent_message" ||
				event.Item.Type == "assistant_message" {
				if event.Item.Text != "" {
					resp.Content += event.Item.Text
				} else if event.Item.Content != "" {
					resp.Content += event.Item.Content
				}
			}
		}
		if event.Text != "" {
			resp.Content += event.Text
		}

		// Extract metadata from completion/finish events
		if event.Usage != nil {
			if resp.Metadata == nil {
				resp.Metadata = &provider.Metadata{}
			}
			inputTokens := event.Usage.PromptTokens
			if inputTokens == 0 {
				inputTokens = event.Usage.InputTokens
			}
			outputTokens := event.Usage.CompletionTokens
			if outputTokens == 0 {
				outputTokens = event.Usage.OutputTokens
			}
			totalTokens := event.Usage.TotalTokens
			if totalTokens == 0 && inputTokens > 0 && outputTokens > 0 {
				totalTokens = inputTokens + outputTokens
			}

			resp.Metadata.InputTokens = inputTokens
			resp.Metadata.OutputTokens = outputTokens
			resp.Metadata.TotalTokens = totalTokens
			if event.Usage.DurationMs > 0 {
				resp.Metadata.Duration = time.Duration(event.Usage.DurationMs) * time.Millisecond
			}
		}
		if event.StopReason != "" {
			if resp.Metadata == nil {
				resp.Metadata = &provider.Metadata{}
			}
			resp.Metadata.StopReason = event.StopReason
		}
		if event.SessionID != "" {
			if resp.Metadata == nil {
				resp.Metadata = &provider.Metadata{}
			}
			resp.Metadata.SessionID = event.SessionID
		}
		if event.ThreadID != "" {
			if resp.Metadata == nil {
				resp.Metadata = &provider.Metadata{}
			}
			if resp.Metadata.SessionID == "" {
				resp.Metadata.SessionID = event.ThreadID
			}
		}
	}

	if foundEvents && resp.Content != "" {
		// Use provided duration if not in JSON
		if resp.Metadata != nil && resp.Metadata.Duration == 0 {
			resp.Metadata.Duration = duration
		}
		return resp, nil
	}

	// Fallback: try parsing as single structured JSON object (legacy schema mode)
	var raw JSONResponse
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		// If not JSON, return as plain text
		return &provider.Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	// Extract content
	if raw.Response != "" {
		resp.Content = raw.Response
	} else if len(raw.Choices) > 0 {
		resp.Content = raw.Choices[0].Message.Content
		resp.Metadata = &provider.Metadata{
			StopReason: raw.Choices[0].FinishReason,
			Duration:   duration,
		}
	} else if raw.Content != "" {
		resp.Content = raw.Content
	}

	// Extract metadata
	if raw.Usage != nil {
		if resp.Metadata == nil {
			resp.Metadata = &provider.Metadata{Duration: duration}
		}
		resp.Metadata.InputTokens = raw.Usage.PromptTokens
		resp.Metadata.OutputTokens = raw.Usage.CompletionTokens
		resp.Metadata.TotalTokens = raw.Usage.TotalTokens
	}

	return resp, nil
}
