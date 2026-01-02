package provider

import (
	"encoding/json"
	"time"
)

// Response represents a provider's response with metadata.
type Response struct {
	Content  string        `json:"content"`
	Model    string        `json:"model,omitempty"`
	Provider string        `json:"provider,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
	Metadata *ResponseMeta `json:"metadata,omitempty"`
	Raw      string        `json:"-"` // Raw response for debugging
}

// ResponseMeta contains additional response metadata.
type ResponseMeta struct {
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
	TotalTokens  int    `json:"total_tokens,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
}

// ClaudeJSONResponse represents Claude CLI JSON output.
type ClaudeJSONResponse struct {
	Type    string `json:"type"`
	Role    string `json:"role,omitempty"`
	Model   string `json:"model,omitempty"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
	Usage      *struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
	Result    string `json:"result,omitempty"` // For simpler responses
	SessionID string `json:"session_id,omitempty"`
}

// ParseClaudeJSON parses Claude CLI JSON output.
func ParseClaudeJSON(data string) (*Response, error) {
	var raw ClaudeJSONResponse
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		// Not JSON, return as plain text response
		return &Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	resp := &Response{
		Model: raw.Model,
		Raw:   data,
	}

	// Extract content from message structure
	if len(raw.Content) > 0 {
		for _, c := range raw.Content {
			if c.Type == "text" {
				resp.Content += c.Text
			}
		}
	} else if raw.Result != "" {
		resp.Content = raw.Result
	}

	// Extract metadata
	if raw.Usage != nil {
		resp.Metadata = &ResponseMeta{
			InputTokens:  raw.Usage.InputTokens,
			OutputTokens: raw.Usage.OutputTokens,
			TotalTokens:  raw.Usage.InputTokens + raw.Usage.OutputTokens,
			StopReason:   raw.StopReason,
			SessionID:    raw.SessionID,
		}
	}

	return resp, nil
}

// GeminiJSONResponse represents Gemini CLI JSON output.
type GeminiJSONResponse struct {
	Response string `json:"response,omitempty"` // Main response text from Gemini CLI
	Stats    *struct {
		Models map[string]interface{} `json:"models,omitempty"`
		Tools  map[string]interface{} `json:"tools,omitempty"`
		Files  map[string]interface{} `json:"files,omitempty"`
	} `json:"stats,omitempty"`
	// Fallback fields for different Gemini API formats
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason,omitempty"`
	} `json:"candidates,omitempty"`
	UsageMetadata *struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata,omitempty"`
	Text string `json:"text,omitempty"` // For simpler responses
}

// ParseGeminiJSON parses Gemini CLI JSON output.
func ParseGeminiJSON(data string) (*Response, error) {
	var raw GeminiJSONResponse
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return &Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	resp := &Response{
		Raw: data,
	}

	// Extract content from Gemini CLI format (priority)
	if raw.Response != "" {
		resp.Content = raw.Response
	} else if raw.Text != "" {
		// Fallback: simpler response format
		resp.Content = raw.Text
	} else if len(raw.Candidates) > 0 && len(raw.Candidates[0].Content.Parts) > 0 {
		// Fallback: traditional Gemini API format
		for _, part := range raw.Candidates[0].Content.Parts {
			resp.Content += part.Text
		}
		if len(raw.Candidates) > 0 {
			resp.Metadata = &ResponseMeta{
				StopReason: raw.Candidates[0].FinishReason,
			}
		}
	}

	// Extract metadata from stats if available (Gemini CLI format)
	if raw.Stats != nil && raw.Stats.Models != nil {
		// Parse token counts from models stats if present
		// The stats structure contains model-specific data with tokens
		if resp.Metadata == nil {
			resp.Metadata = &ResponseMeta{}
		}
	}

	// Fallback metadata extraction from usageMetadata
	if raw.UsageMetadata != nil {
		if resp.Metadata == nil {
			resp.Metadata = &ResponseMeta{}
		}
		resp.Metadata.InputTokens = raw.UsageMetadata.PromptTokenCount
		resp.Metadata.OutputTokens = raw.UsageMetadata.CandidatesTokenCount
		resp.Metadata.TotalTokens = raw.UsageMetadata.TotalTokenCount
	}

	return resp, nil
}

// CodexJSONResponse represents OpenAI/Codex CLI structured output.
type CodexJSONResponse struct {
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

// ParseCodexJSON parses OpenAI/Codex CLI structured JSON output.
func ParseCodexJSON(data string) (*Response, error) {
	var raw CodexJSONResponse
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		// If not JSON, return as plain text
		return &Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	resp := &Response{
		Raw: data,
	}

	// Extract content
	if raw.Response != "" {
		resp.Content = raw.Response
	} else if len(raw.Choices) > 0 {
		resp.Content = raw.Choices[0].Message.Content
		resp.Metadata = &ResponseMeta{
			StopReason: raw.Choices[0].FinishReason,
		}
	} else if raw.Content != "" {
		resp.Content = raw.Content
	}

	// Extract metadata
	if raw.Usage != nil {
		if resp.Metadata == nil {
			resp.Metadata = &ResponseMeta{}
		}
		resp.Metadata.InputTokens = raw.Usage.PromptTokens
		resp.Metadata.OutputTokens = raw.Usage.CompletionTokens
		resp.Metadata.TotalTokens = raw.Usage.TotalTokens
	}

	return resp, nil
}

// QwenJSONResponse represents legacy Qwen CLI JSON output.
type QwenJSONResponse struct {
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

// QwenEvent represents an event in the newer Qwen CLI JSON array output.
type QwenEvent struct {
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

// ParseQwenJSON parses Qwen CLI JSON output.
func ParseQwenJSON(data string) (*Response, error) {
	// Try parsing as array of events first (newer format)
	var events []QwenEvent
	if err := json.Unmarshal([]byte(data), &events); err == nil && len(events) > 0 {
		resp := &Response{Raw: data}
		
		var resultText string
		var assistantText string
		var usage *ResponseMeta

		// Scan all events to collect information
		for _, event := range events {
			if event.Type == "result" && event.Result != "" {
				resultText = event.Result
				if event.Usage != nil {
					usage = &ResponseMeta{
						InputTokens:  event.Usage.InputTokens,
						OutputTokens: event.Usage.OutputTokens,
						TotalTokens:  event.Usage.TotalTokens,
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
					usage = &ResponseMeta{
						InputTokens:  event.Usage.InputTokens,
						OutputTokens: event.Usage.OutputTokens,
						TotalTokens:  event.Usage.TotalTokens,
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
	var raw QwenJSONResponse
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return &Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	resp := &Response{
		Raw: data,
	}

	// Extract content
	if raw.Output.Text != "" {
		resp.Content = raw.Output.Text
		resp.Metadata = &ResponseMeta{
			StopReason: raw.Output.FinishReason,
		}
	} else if raw.Text != "" {
		resp.Content = raw.Text
	}

	// Extract metadata
	if raw.Usage != nil {
		if resp.Metadata == nil {
			resp.Metadata = &ResponseMeta{}
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
