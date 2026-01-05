package qwen

import (
	"testing"
	"time"
)

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantContent    string
		wantInputToks  int
		wantOutputToks int
		wantTotalToks  int
		wantStopReason string
	}{
		{
			name: "event_array_with_result",
			input: `[
				{"type": "assistant", "message": {"content": [{"type": "text", "text": "Intermediate response."}]}},
				{"type": "result", "result": "Consider team expertise.", "usage": {"input_tokens": 25, "output_tokens": 10, "total_tokens": 35}}
			]`,
			wantContent:    "Consider team expertise.",
			wantInputToks:  25,
			wantOutputToks: 10,
			wantTotalToks:  35,
		},
		{
			name: "event_array_assistant_only",
			input: `[
				{"type": "assistant", "message": {"content": [{"type": "text", "text": "First part. "}, {"type": "text", "text": "Second part."}]}, "usage": {"input_tokens": 30, "output_tokens": 15, "total_tokens": 45}}
			]`,
			wantContent:    "First part. Second part.",
			wantInputToks:  30,
			wantOutputToks: 15,
			wantTotalToks:  45,
		},
		{
			name: "event_array_result_priority",
			input: `[
				{"type": "assistant", "message": {"content": [{"type": "text", "text": "Assistant text."}]}},
				{"type": "result", "result": "Final result text.", "usage": {"input_tokens": 20, "output_tokens": 8, "total_tokens": 28}}
			]`,
			wantContent:    "Final result text.",
			wantInputToks:  20,
			wantOutputToks: 8,
			wantTotalToks:  28,
		},
		{
			name: "legacy_output_format",
			input: `{
				"output": {
					"text": "Legacy format response.",
					"finish_reason": "stop"
				},
				"usage": {
					"input_tokens": 40,
					"output_tokens": 20,
					"total_tokens": 60
				}
			}`,
			wantContent:    "Legacy format response.",
			wantInputToks:  40,
			wantOutputToks: 20,
			wantTotalToks:  60,
			wantStopReason: "stop",
		},
		{
			name: "legacy_simple_text",
			input: `{
				"text": "Simple text response."
			}`,
			wantContent: "Simple text response.",
		},
		{
			name: "legacy_output_without_usage",
			input: `{
				"output": {
					"text": "Response without usage.",
					"finish_reason": "complete"
				}
			}`,
			wantContent:    "Response without usage.",
			wantStopReason: "complete",
		},
		{
			name: "legacy_usage_calculates_total",
			input: `{
				"output": {"text": "Test"},
				"usage": {
					"input_tokens": 15,
					"output_tokens": 7
				}
			}`,
			wantContent:    "Test",
			wantInputToks:  15,
			wantOutputToks: 7,
			wantTotalToks:  22,
		},
		{
			name:        "plain_text_fallback",
			input:       "Not valid JSON at all.",
			wantContent: "Not valid JSON at all.",
		},
		{
			name: "empty_event_array",
			input: `[]`,
			wantContent: "[]",
		},
		{
			name: "event_array_non_text_content",
			input: `[
				{"type": "assistant", "message": {"content": [{"type": "thinking", "text": "Thinking..."}, {"type": "text", "text": "Actual response."}]}}
			]`,
			wantContent: "Actual response.",
		},
		{
			name: "multiple_assistant_events",
			input: `[
				{"type": "assistant", "message": {"content": [{"type": "text", "text": "First "}]}},
				{"type": "assistant", "message": {"content": [{"type": "text", "text": "second."}]}}
			]`,
			wantContent: "First second.",
		},
		{
			name: "result_with_empty_string",
			input: `[
				{"type": "assistant", "message": {"content": [{"type": "text", "text": "Fallback content."}]}},
				{"type": "result", "result": ""}
			]`,
			wantContent: "Fallback content.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ParseJSON(tt.input, 100*time.Millisecond)
			if err != nil {
				t.Fatalf("ParseJSON() error = %v", err)
			}

			if resp.Content != tt.wantContent {
				t.Errorf("Content = %q, want %q", resp.Content, tt.wantContent)
			}

			if tt.wantInputToks > 0 || tt.wantOutputToks > 0 || tt.wantTotalToks > 0 {
				if resp.Metadata == nil {
					t.Fatal("Metadata is nil, expected token counts")
				}
				if resp.Metadata.InputTokens != tt.wantInputToks {
					t.Errorf("InputTokens = %d, want %d", resp.Metadata.InputTokens, tt.wantInputToks)
				}
				if resp.Metadata.OutputTokens != tt.wantOutputToks {
					t.Errorf("OutputTokens = %d, want %d", resp.Metadata.OutputTokens, tt.wantOutputToks)
				}
				if resp.Metadata.TotalTokens != tt.wantTotalToks {
					t.Errorf("TotalTokens = %d, want %d", resp.Metadata.TotalTokens, tt.wantTotalToks)
				}
			}

			if tt.wantStopReason != "" {
				if resp.Metadata == nil {
					t.Fatal("Metadata is nil, expected stop_reason")
				}
				if resp.Metadata.StopReason != tt.wantStopReason {
					t.Errorf("StopReason = %q, want %q", resp.Metadata.StopReason, tt.wantStopReason)
				}
			}

			if resp.Raw != tt.input {
				t.Errorf("Raw = %q, want %q", resp.Raw, tt.input)
			}
		})
	}
}

func TestParseJSON_Duration(t *testing.T) {
	input := `[
		{"type": "result", "result": "Test", "usage": {"input_tokens": 10, "output_tokens": 5, "total_tokens": 15}}
	]`

	duration := 300 * time.Millisecond
	resp, err := ParseJSON(input, duration)
	if err != nil {
		t.Fatalf("ParseJSON() error = %v", err)
	}

	if resp.Metadata == nil {
		t.Fatal("Metadata is nil")
	}

	if resp.Metadata.Duration != duration {
		t.Errorf("Duration = %v, want %v", resp.Metadata.Duration, duration)
	}
}
