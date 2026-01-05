package claude

import (
	"testing"
	"time"
)

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantContent    string
		wantModel      string
		wantInputToks  int
		wantOutputToks int
		wantStopReason string
		wantSessionID  string
	}{
		{
			name: "full_message_response",
			input: `{
				"type": "result",
				"role": "assistant",
				"model": "claude-sonnet-4-20250514",
				"content": [
					{"type": "text", "text": "Consider scalability and team size."}
				],
				"stop_reason": "end_turn",
				"usage": {
					"input_tokens": 25,
					"output_tokens": 10
				},
				"session_id": "abc123"
			}`,
			wantContent:    "Consider scalability and team size.",
			wantModel:      "claude-sonnet-4-20250514",
			wantInputToks:  25,
			wantOutputToks: 10,
			wantStopReason: "end_turn",
			wantSessionID:  "abc123",
		},
		{
			name: "multiple_content_blocks",
			input: `{
				"type": "result",
				"model": "claude-sonnet-4-20250514",
				"content": [
					{"type": "text", "text": "First part. "},
					{"type": "text", "text": "Second part."}
				],
				"stop_reason": "end_turn",
				"usage": {
					"input_tokens": 50,
					"output_tokens": 20
				}
			}`,
			wantContent:    "First part. Second part.",
			wantModel:      "claude-sonnet-4-20250514",
			wantInputToks:  50,
			wantOutputToks: 20,
			wantStopReason: "end_turn",
		},
		{
			name: "simple_result_response",
			input: `{
				"type": "result",
				"result": "Depends on complexity.",
				"model": "claude-sonnet-4-20250514"
			}`,
			wantContent: "Depends on complexity.",
			wantModel:   "claude-sonnet-4-20250514",
		},
		{
			name: "content_with_non_text_blocks",
			input: `{
				"type": "result",
				"model": "claude-sonnet-4-20250514",
				"content": [
					{"type": "thinking", "text": "Let me think about this..."},
					{"type": "text", "text": "Here's my answer."}
				],
				"stop_reason": "end_turn"
			}`,
			wantContent: "Here's my answer.",
			wantModel:   "claude-sonnet-4-20250514",
		},
		{
			name:        "plain_text_response",
			input:       "This is not JSON, just plain text.",
			wantContent: "This is not JSON, just plain text.",
		},
		{
			name: "empty_content_array",
			input: `{
				"type": "result",
				"model": "claude-sonnet-4-20250514",
				"content": [],
				"result": "Fallback result"
			}`,
			wantContent: "Fallback result",
			wantModel:   "claude-sonnet-4-20250514",
		},
		{
			name: "max_tokens_stop_reason",
			input: `{
				"type": "result",
				"model": "claude-sonnet-4-20250514",
				"content": [
					{"type": "text", "text": "Truncated response..."}
				],
				"stop_reason": "max_tokens",
				"usage": {
					"input_tokens": 100,
					"output_tokens": 4096
				}
			}`,
			wantContent:    "Truncated response...",
			wantModel:      "claude-sonnet-4-20250514",
			wantInputToks:  100,
			wantOutputToks: 4096,
			wantStopReason: "max_tokens",
		},
		{
			name: "actual_cli_format_with_cache",
			input: `{
				"type": "result",
				"subtype": "success",
				"is_error": false,
				"duration_ms": 35230,
				"num_turns": 1,
				"result": "Probably not. Start monolith first.",
				"session_id": "047959ab-15e4-4aa7-b7ef-db7a5f2dddd6",
				"total_cost_usd": 0.12558,
				"usage": {
					"input_tokens": 3,
					"cache_creation_input_tokens": 32789,
					"cache_read_input_tokens": 0,
					"output_tokens": 13
				}
			}`,
			wantContent:    "Probably not. Start monolith first.",
			wantInputToks:  32792,
			wantOutputToks: 13,
			wantSessionID:  "047959ab-15e4-4aa7-b7ef-db7a5f2dddd6",
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

			if resp.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", resp.Model, tt.wantModel)
			}

			if tt.wantInputToks > 0 || tt.wantOutputToks > 0 {
				if resp.Metadata == nil {
					t.Fatal("Metadata is nil, expected token counts")
				}
				if resp.Metadata.InputTokens != tt.wantInputToks {
					t.Errorf("InputTokens = %d, want %d", resp.Metadata.InputTokens, tt.wantInputToks)
				}
				if resp.Metadata.OutputTokens != tt.wantOutputToks {
					t.Errorf("OutputTokens = %d, want %d", resp.Metadata.OutputTokens, tt.wantOutputToks)
				}
				wantTotal := tt.wantInputToks + tt.wantOutputToks
				if resp.Metadata.TotalTokens != wantTotal {
					t.Errorf("TotalTokens = %d, want %d", resp.Metadata.TotalTokens, wantTotal)
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

			if tt.wantSessionID != "" {
				if resp.Metadata == nil {
					t.Fatal("Metadata is nil, expected session_id")
				}
				if resp.Metadata.SessionID != tt.wantSessionID {
					t.Errorf("SessionID = %q, want %q", resp.Metadata.SessionID, tt.wantSessionID)
				}
			}

			if resp.Raw != tt.input {
				t.Errorf("Raw = %q, want %q", resp.Raw, tt.input)
			}
		})
	}
}

func TestParseJSON_Duration(t *testing.T) {
	// Test fallback to measured duration
	input := `{
		"type": "result",
		"content": [{"type": "text", "text": "Test"}],
		"usage": {"input_tokens": 10, "output_tokens": 5}
	}`

	duration := 500 * time.Millisecond
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

func TestParseJSON_CLIDuration(t *testing.T) {
	// Test CLI's duration_ms takes precedence
	input := `{
		"type": "result",
		"result": "Test response",
		"duration_ms": 1234,
		"usage": {"input_tokens": 10, "output_tokens": 5}
	}`

	measuredDuration := 500 * time.Millisecond
	resp, err := ParseJSON(input, measuredDuration)
	if err != nil {
		t.Fatalf("ParseJSON() error = %v", err)
	}

	if resp.Metadata == nil {
		t.Fatal("Metadata is nil")
	}

	wantDuration := 1234 * time.Millisecond
	if resp.Metadata.Duration != wantDuration {
		t.Errorf("Duration = %v, want %v (CLI's duration_ms should take precedence)", resp.Metadata.Duration, wantDuration)
	}
}
