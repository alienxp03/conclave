package provider

import (
	"encoding/json"
	"time"
)

// Response represents a provider's response with metadata.
type Response struct {
	Content   string         `json:"content"`
	Model     string         `json:"model,omitempty"`
	Provider  string         `json:"provider,omitempty"`
	Duration  time.Duration  `json:"duration,omitempty"`
	Metadata  *ResponseMeta  `json:"metadata,omitempty"`
	Raw       string         `json:"-"` // Raw response for debugging
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

	// Extract content
	if len(raw.Candidates) > 0 && len(raw.Candidates[0].Content.Parts) > 0 {
		for _, part := range raw.Candidates[0].Content.Parts {
			resp.Content += part.Text
		}
		if len(raw.Candidates) > 0 {
			resp.Metadata = &ResponseMeta{
				StopReason: raw.Candidates[0].FinishReason,
			}
		}
	} else if raw.Text != "" {
		resp.Content = raw.Text
	}

	// Extract metadata
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

// CodexJSONResponse represents OpenAI/Codex CLI JSON output.
type CodexJSONResponse struct {
	ID      string `json:"id,omitempty"`
	Model   string `json:"model,omitempty"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices,omitempty"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
	Content string `json:"content,omitempty"` // For simpler responses
}

// ParseCodexJSON parses OpenAI/Codex CLI JSON output.
func ParseCodexJSON(data string) (*Response, error) {
	var raw CodexJSONResponse
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return &Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	resp := &Response{
		Model: raw.Model,
		Raw:   data,
	}

	// Extract content
	if len(raw.Choices) > 0 {
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

// QwenJSONResponse represents Qwen CLI JSON output.
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

// ParseQwenJSON parses Qwen CLI JSON output.
func ParseQwenJSON(data string) (*Response, error) {
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
