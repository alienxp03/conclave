package opencode

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
		wantStopReason string
		wantSessionID  string
	}{
		{
			name: "standard_jsonl_output",
			input: `{"type":"text","sessionID":"sess-001","part":{"type":"text","text":"Start with a "}}
{"type":"text","sessionID":"sess-001","part":{"type":"text","text":"monolith first."}}
{"type":"step_finish","sessionID":"sess-001","part":{"type":"finish","reason":"end_turn","tokens":{"input":30,"output":12}}}`,
			wantContent:    "Start with a monolith first.",
			wantInputToks:  30,
			wantOutputToks: 12,
			wantStopReason: "end_turn",
			wantSessionID:  "sess-001",
		},
		{
			name: "single_text_event",
			input: `{"type":"text","sessionID":"abc123","part":{"type":"text","text":"Simple response."}}
{"type":"step_finish","sessionID":"abc123","part":{"type":"finish","reason":"stop"}}`,
			wantContent:    "Simple response.",
			wantStopReason: "stop",
			wantSessionID:  "abc123",
		},
		{
			name: "multiple_text_chunks",
			input: `{"type":"text","sessionID":"sess-002","part":{"type":"text","text":"First "}}
{"type":"text","sessionID":"sess-002","part":{"type":"text","text":"second "}}
{"type":"text","sessionID":"sess-002","part":{"type":"text","text":"third."}}
{"type":"step_finish","sessionID":"sess-002","part":{"type":"finish","reason":"complete","tokens":{"input":50,"output":25}}}`,
			wantContent:    "First second third.",
			wantInputToks:  50,
			wantOutputToks: 25,
			wantStopReason: "complete",
			wantSessionID:  "sess-002",
		},
		{
			name: "with_empty_lines",
			input: `{"type":"text","sessionID":"sess-003","part":{"type":"text","text":"Response with "}}

{"type":"text","sessionID":"sess-003","part":{"type":"text","text":"empty lines."}}

{"type":"step_finish","sessionID":"sess-003","part":{"type":"finish","reason":"done"}}`,
			wantContent:    "Response with empty lines.",
			wantStopReason: "done",
			wantSessionID:  "sess-003",
		},
		{
			name:        "plain_text_fallback",
			input:       "Not JSON lines, plain text output.",
			wantContent: "Not JSON lines, plain text output.",
		},
		{
			name: "malformed_json_mixed",
			input: `{"type":"text","sessionID":"sess-004","part":{"type":"text","text":"Valid text."}}
{malformed json line}
{"type":"step_finish","sessionID":"sess-004","part":{"type":"finish","reason":"end"}}`,
			wantContent:    "Valid text.",
			wantStopReason: "end",
			wantSessionID:  "sess-004",
		},
		{
			name: "no_finish_event",
			input: `{"type":"text","sessionID":"sess-005","part":{"type":"text","text":"Text without finish."}}`,
			wantContent: "Text without finish.",
		},
		{
			name: "finish_without_tokens",
			input: `{"type":"text","sessionID":"sess-006","part":{"type":"text","text":"Response here."}}
{"type":"step_finish","sessionID":"sess-006","part":{"type":"finish","reason":"stopped"}}`,
			wantContent:    "Response here.",
			wantStopReason: "stopped",
			wantSessionID:  "sess-006",
		},
		{
			name: "max_tokens_reason",
			input: `{"type":"text","sessionID":"sess-007","part":{"type":"text","text":"Truncated..."}}
{"type":"step_finish","sessionID":"sess-007","part":{"type":"finish","reason":"max_tokens","tokens":{"input":100,"output":4096}}}`,
			wantContent:    "Truncated...",
			wantInputToks:  100,
			wantOutputToks: 4096,
			wantStopReason: "max_tokens",
			wantSessionID:  "sess-007",
		},
		{
			name: "other_event_types_ignored",
			input: `{"type":"thinking","sessionID":"sess-008","part":{"type":"text","text":"Thinking..."}}
{"type":"text","sessionID":"sess-008","part":{"type":"text","text":"Actual response."}}
{"type":"tool_use","sessionID":"sess-008","part":{"type":"tool","name":"search"}}
{"type":"step_finish","sessionID":"sess-008","part":{"type":"finish","reason":"done"}}`,
			wantContent:    "Actual response.",
			wantStopReason: "done",
			wantSessionID:  "sess-008",
		},
		{
			name: "actual_cli_format_with_reasoning",
			input: `{"type":"step_start","timestamp":1767565503730,"sessionID":"ses_test123","part":{"type":"step-start"}}
{"type":"text","timestamp":1767565503731,"sessionID":"ses_test123","part":{"type":"text","text":"Depends on scale and complexity."}}
{"type":"step_finish","timestamp":1767565503744,"sessionID":"ses_test123","part":{"type":"step-finish","reason":"stop","tokens":{"input":13105,"output":6,"reasoning":317,"cache":{"read":0,"write":0}}}}`,
			wantContent:    "Depends on scale and complexity.",
			wantInputToks:  13105,
			wantOutputToks: 6,
			wantStopReason: "stop",
			wantSessionID:  "ses_test123",
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
	input := `{"type":"text","sessionID":"test","part":{"type":"text","text":"Test"}}
{"type":"step_finish","sessionID":"test","part":{"type":"finish","reason":"done","tokens":{"input":10,"output":5}}}`

	duration := 750 * time.Millisecond
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
