package provider

import (
	"testing"
)

func TestParseClaudeJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantContent string
		wantMeta    bool
	}{
		{
			name: "full response with content array",
			input: `{
				"type": "message",
				"role": "assistant",
				"model": "claude-3-sonnet",
				"content": [
					{"type": "text", "text": "Hello, this is a test response."}
				],
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 10, "output_tokens": 20}
			}`,
			wantContent: "Hello, this is a test response.",
			wantMeta:    true,
		},
		{
			name:        "simple result field",
			input:       `{"result": "Simple response text"}`,
			wantContent: "Simple response text",
			wantMeta:    false,
		},
		{
			name:        "plain text fallback",
			input:       "This is plain text, not JSON",
			wantContent: "This is plain text, not JSON",
			wantMeta:    false,
		},
		{
			name: "multiple text blocks",
			input: `{
				"content": [
					{"type": "text", "text": "First part. "},
					{"type": "text", "text": "Second part."}
				]
			}`,
			wantContent: "First part. Second part.",
			wantMeta:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ParseClaudeJSON(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Content != tt.wantContent {
				t.Errorf("content mismatch: got %q, want %q", resp.Content, tt.wantContent)
			}
			if tt.wantMeta && resp.Metadata == nil {
				t.Error("expected metadata but got nil")
			}
			if !tt.wantMeta && resp.Metadata != nil {
				t.Error("unexpected metadata")
			}
		})
	}
}

func TestParseGeminiJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantContent string
		wantMeta    bool
	}{
		{
			name: "full response",
			input: `{
				"candidates": [{
					"content": {
						"parts": [{"text": "Gemini response text"}]
					},
					"finishReason": "STOP"
				}],
				"usageMetadata": {
					"promptTokenCount": 5,
					"candidatesTokenCount": 15,
					"totalTokenCount": 20
				}
			}`,
			wantContent: "Gemini response text",
			wantMeta:    true,
		},
		{
			name:        "simple text field",
			input:       `{"text": "Simple text"}`,
			wantContent: "Simple text",
			wantMeta:    false,
		},
		{
			name:        "plain text fallback",
			input:       "Plain text response",
			wantContent: "Plain text response",
			wantMeta:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ParseGeminiJSON(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Content != tt.wantContent {
				t.Errorf("content mismatch: got %q, want %q", resp.Content, tt.wantContent)
			}
			if tt.wantMeta && resp.Metadata == nil {
				t.Error("expected metadata but got nil")
			}
		})
	}
}

func TestParseCodexJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantContent string
		wantMeta    bool
	}{
		{
			name: "full response",
			input: `{
				"id": "test-123",
				"model": "gpt-4",
				"choices": [{
					"message": {"role": "assistant", "content": "Codex response"},
					"finish_reason": "stop"
				}],
				"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
			}`,
			wantContent: "Codex response",
			wantMeta:    true,
		},
		{
			name:        "simple content field",
			input:       `{"content": "Direct content"}`,
			wantContent: "Direct content",
			wantMeta:    false,
		},
		{
			name:        "plain text fallback",
			input:       "Just plain text",
			wantContent: "Just plain text",
			wantMeta:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ParseCodexJSON(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Content != tt.wantContent {
				t.Errorf("content mismatch: got %q, want %q", resp.Content, tt.wantContent)
			}
			if tt.wantMeta && resp.Metadata == nil {
				t.Error("expected metadata but got nil")
			}
		})
	}
}

func TestParseQwenJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantContent string
		wantMeta    bool
	}{
		{
			name: "full response",
			input: `{
				"output": {"text": "Qwen response", "finish_reason": "stop"},
				"usage": {"input_tokens": 10, "output_tokens": 20}
			}`,
			wantContent: "Qwen response",
			wantMeta:    true,
		},
		{
			name:        "simple text field",
			input:       `{"text": "Simple text"}`,
			wantContent: "Simple text",
			wantMeta:    false,
		},
		{
			name:        "plain text fallback",
			input:       "Plain Qwen response",
			wantContent: "Plain Qwen response",
			wantMeta:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ParseQwenJSON(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Content != tt.wantContent {
				t.Errorf("content mismatch: got %q, want %q", resp.Content, tt.wantContent)
			}
			if tt.wantMeta && resp.Metadata == nil {
				t.Error("expected metadata but got nil")
			}
		})
	}
}

func TestResponseMetadata(t *testing.T) {
	input := `{
		"content": [{"type": "text", "text": "Test"}],
		"usage": {"input_tokens": 100, "output_tokens": 50}
	}`

	resp, _ := ParseClaudeJSON(input)

	if resp.Metadata == nil {
		t.Fatal("metadata is nil")
	}
	if resp.Metadata.InputTokens != 100 {
		t.Errorf("input tokens: got %d, want 100", resp.Metadata.InputTokens)
	}
	if resp.Metadata.OutputTokens != 50 {
		t.Errorf("output tokens: got %d, want 50", resp.Metadata.OutputTokens)
	}
	if resp.Metadata.TotalTokens != 150 {
		t.Errorf("total tokens: got %d, want 150", resp.Metadata.TotalTokens)
	}
}
