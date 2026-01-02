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
			name: "full response (legacy)",
			input: `{
				"output": {"text": "Qwen response", "finish_reason": "stop"},
				"usage": {"input_tokens": 10, "output_tokens": 20}
			}`,
			wantContent: "Qwen response",
			wantMeta:    true,
		},
		{
			name: "newer array format",
			input: `[
				{"type": "system", "subtype": "init"},
				{
					"type": "assistant",
					"message": {
						"content": [{"type": "text", "text": "Assistant thinking..."}]
					}
				},
				{
					"type": "result",
					"result": "Final Qwen answer",
					"usage": {"input_tokens": 100, "output_tokens": 50, "total_tokens": 150}
				}
			]`,
			wantContent: "Final Qwen answer",
			wantMeta:    true,
		},
		{
			name: "peer review stage format",
			input: `[
    {
        "type": "system",
        "subtype": "init",
        "uuid": "17d4257f-8df7-4884-b5f9-50ecf1e23722",
        "session_id": "17d4257f-8df7-4884-b5f9-50ecf1e23722",
        "cwd": "/Users/azuan/Workspace/Projects/conclave",
        "tools": ["task", "list_directory"],
        "mcp_servers": [{"name": "playwright", "status": "connected"}],
        "model": "coder-model",
        "permission_mode": "default",
        "slash_commands": ["compress", "init", "summary"],
        "qwen_code_version": "0.6.0",
        "agents": ["general-purpose"]
    },
    {
        "type": "assistant",
        "uuid": "33a5ff01-1128-4e91-8444-49dbba09226c",
        "session_id": "17d4257f-8df7-4884-b5f9-50ecf1e23722",
        "parent_tool_use_id": null,
        "message": {
            "id": "33a5ff01-1128-4e91-8444-49dbba09226c",
            "type": "message",
            "role": "assistant",
            "model": "coder-model",
            "content": [
                {
                    "type": "text",
                    "text": "I'll evaluate both responses..."
                }
            ],
            "stop_reason": null,
            "usage": {
                "input_tokens": 16080,
                "output_tokens": 343,
                "cache_read_input_tokens": 14838,
                "total_tokens": 16423
            }
        }
    },
    {
        "type": "result",
        "subtype": "success",
        "uuid": "4f714e6b-78d6-4155-8fc0-5be4a5749a8b",
        "session_id": "17d4257f-8df7-4884-b5f9-50ecf1e23722",
        "is_error": false,
        "duration_ms": 8572,
        "duration_api_ms": 8536,
        "num_turns": 1,
        "result": "I'll evaluate both responses on their accuracy, insight, and quality for the given topic.\n\n**Response by qwen (Pragmatist):**\n- **Accuracy**: High - Correctly identifies that a 3-person team would struggle with microservice complexity, deployment overhead, and coordination\n- **Insight**: Strong - Recognizes the resource constraints of a small team and the importance of focusing on the core product\n- **Quality**: Excellent - Provides a clear, direct answer that addresses the specific context (3-person team) with practical reasoning\n- **Word count**: 20 words as requested: \"No, stick with monolith. 3-person team can't handle microservice complexity, deployment overhead, and coordination challenges. Focus resources on core product.\"\n\n**Response by codex (Optimist):**\n- **Accuracy**: Moderate - While microservices can offer benefits, it oversimplifies the challenges for a 3-person team\n- **Insight**: Limited - Focuses on potential benefits but doesn't adequately address the significant overhead and complexity for a small team\n- **Quality**: Lower - Lacks the practical context needed for a 3-person team; \"disciplined automation and observability\" requires significant resources a 3-person team likely doesn't have\n- **Word count**: Under 20 words but doesn't fully address the practical constraints\n\nThe pragmatist response correctly identifies that microservices introduce complexity that would overwhelm a small team, while the optimist response focuses on potential benefits without adequately addressing the practical challenges for a 3-person team.\n\nFINAL RANKING:\n1. qwen (Pragmatist)\n2. codex (Optimist)",
        "usage": {
            "input_tokens": 16080,
            "output_tokens": 343,
            "cache_read_input_tokens": 14838,
            "total_tokens": 16423
        },
        "permission_denials": [],
        "stats": {
            "models": {
                "coder-model": {
                    "api": {
                        "totalRequests": 1,
                        "totalErrors": 0,
                        "totalLatencyMs": 8495
                    },
                    "tokens": {
                        "prompt": 16080,
                        "candidates": 343,
                        "total": 16423,
                        "cached": 14838,
                        "thoughts": 0,
                        "tool": 0
                    }
                }
            },
            "tools": {
                "totalCalls": 0,
                "totalSuccess": 0,
                "totalFail": 0,
                "totalDurationMs": 0,
                "totalDecisions": {
                    "accept": 0,
                    "reject": 0,
                    "modify": 0,
                    "auto_accept": 0
                },
                "byName": {}
            },
            "files": {
                "totalLinesAdded": 0,
                "totalLinesRemoved": 0
            }
        }
    }
]`,
			wantContent: "I'll evaluate both responses on their accuracy, insight, and quality for the given topic.\n\n**Response by qwen (Pragmatist):**\n- **Accuracy**: High - Correctly identifies that a 3-person team would struggle with microservice complexity, deployment overhead, and coordination\n- **Insight**: Strong - Recognizes the resource constraints of a small team and the importance of focusing on the core product\n- **Quality**: Excellent - Provides a clear, direct answer that addresses the specific context (3-person team) with practical reasoning\n- **Word count**: 20 words as requested: \"No, stick with monolith. 3-person team can't handle microservice complexity, deployment overhead, and coordination challenges. Focus resources on core product.\"\n\n**Response by codex (Optimist):**\n- **Accuracy**: Moderate - While microservices can offer benefits, it oversimplifies the challenges for a 3-person team\n- **Insight**: Limited - Focuses on potential benefits but doesn't adequately address the significant overhead and complexity for a small team\n- **Quality**: Lower - Lacks the practical context needed for a 3-person team; \"disciplined automation and observability\" requires significant resources a 3-person team likely doesn't have\n- **Word count**: Under 20 words but doesn't fully address the practical constraints\n\nThe pragmatist response correctly identifies that microservices introduce complexity that would overwhelm a small team, while the optimist response focuses on potential benefits without adequately addressing the practical challenges for a 3-person team.\n\nFINAL RANKING:\n1. qwen (Pragmatist)\n2. codex (Optimist)",
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
