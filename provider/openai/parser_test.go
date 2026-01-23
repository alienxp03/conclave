package openai

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
		wantSessionID  string
	}{
		{
			name: "codex_jsonl_agent_message",
			input: `{"type":"thread.started","thread_id":"thread-123"}
{"type":"turn.started"}
{"type":"item.completed","item":{"id":"item_0","type":"reasoning","text":"Thinking..."}}
{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"Probably not; keep it simple."}}
{"type":"turn.completed","usage":{"input_tokens":120,"cached_input_tokens":20,"output_tokens":5}}`,
			wantContent:    "Probably not; keep it simple.",
			wantInputToks:  120,
			wantOutputToks: 5,
			wantTotalToks:  125,
			wantSessionID:  "thread-123",
		},
		{
			name: "codex_jsonl_role_assistant",
			input: `{"type":"item.completed","item":{"id":"item_2","type":"message","role":"assistant","text":"Hi there."}}
{"type":"turn.completed","usage":{"input_tokens":10,"output_tokens":2}}`,
			wantContent:    "Hi there.",
			wantInputToks:  10,
			wantOutputToks: 2,
			wantTotalToks:  12,
		},
		{
			name:        "legacy_message_event",
			input:       `{"type":"message","message":{"role":"assistant","content":"Legacy hello."}}`,
			wantContent: "Legacy hello.",
		},
		{
			name: "legacy_structured_json",
			input: `{"choices":[{"message":{"content":"Structured response."},"finish_reason":"stop"}],
  "usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`,
			wantContent:    "Structured response.",
			wantInputToks:  3,
			wantOutputToks: 2,
			wantTotalToks:  5,
		},
		{
			name:        "plain_text_fallback",
			input:       "Not JSON at all.",
			wantContent: "Not JSON at all.",
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

			if tt.wantSessionID != "" {
				if resp.Metadata == nil || resp.Metadata.SessionID != tt.wantSessionID {
					t.Fatalf("SessionID = %v, want %v", resp.Metadata, tt.wantSessionID)
				}
			}
		})
	}
}
