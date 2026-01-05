package gemini

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
			name: "cli_format_with_stats",
			input: `{
				"response": "Start small, scale later.",
				"stats": {
					"models": {
						"gemini-3-flash-preview": {
							"api": {
								"totalRequests": 1,
								"totalErrors": 0,
								"totalLatencyMs": 1234
							},
							"tokens": {
								"prompt": 30,
								"candidates": 15,
								"total": 45,
								"cached": 0,
								"thoughts": 0,
								"tool": 0
							}
						}
					}
				}
			}`,
			wantContent:    "Start small, scale later.",
			wantInputToks:  30,
			wantOutputToks: 15,
			wantTotalToks:  45,
		},
		{
			name: "cli_format_with_input_field",
			input: `{
				"session_id": "e3ecf3fe-673b-480c-8e40-41f391ca40ca",
				"response": "Start with a modular monolith.",
				"stats": {
					"models": {
						"gemini-3-flash-preview": {
							"api": {
								"totalRequests": 1,
								"totalErrors": 0,
								"totalLatencyMs": 3767
							},
							"tokens": {
								"input": 6126,
								"prompt": 6126,
								"candidates": 6,
								"total": 6346,
								"cached": 0,
								"thoughts": 214,
								"tool": 0
							}
						}
					}
				}
			}`,
			wantContent:    "Start with a modular monolith.",
			wantInputToks:  6126,
			wantOutputToks: 6,
			wantTotalToks:  6346,
		},
		{
			name: "multiple_models_in_stats",
			input: `{
				"response": "Combined response from models.",
				"stats": {
					"models": {
						"gemini-3-flash": {
							"tokens": {"prompt": 20, "candidates": 10, "total": 30}
						},
						"gemini-3-pro": {
							"tokens": {"prompt": 15, "candidates": 8, "total": 23}
						}
					}
				}
			}`,
			wantContent:    "Combined response from models.",
			wantInputToks:  35,
			wantOutputToks: 18,
			wantTotalToks:  53,
		},
		{
			name: "simple_text_response",
			input: `{
				"text": "Simple text response format."
			}`,
			wantContent: "Simple text response format.",
		},
		{
			name: "candidates_format",
			input: `{
				"candidates": [
					{
						"content": {
							"parts": [
								{"text": "Part one. "},
								{"text": "Part two."}
							]
						},
						"finishReason": "STOP"
					}
				],
				"usageMetadata": {
					"promptTokenCount": 25,
					"candidatesTokenCount": 12,
					"totalTokenCount": 37
				}
			}`,
			wantContent:    "Part one. Part two.",
			wantInputToks:  25,
			wantOutputToks: 12,
			wantTotalToks:  37,
			wantStopReason: "STOP",
		},
		{
			name: "candidates_with_max_tokens",
			input: `{
				"candidates": [
					{
						"content": {
							"parts": [{"text": "Truncated..."}]
						},
						"finishReason": "MAX_TOKENS"
					}
				]
			}`,
			wantContent:    "Truncated...",
			wantStopReason: "MAX_TOKENS",
		},
		{
			name:        "plain_text_fallback",
			input:       "Not valid JSON, plain text output.",
			wantContent: "Not valid JSON, plain text output.",
		},
		{
			name: "empty_response_with_stats",
			input: `{
				"response": "",
				"text": "Fallback text",
				"stats": {
					"models": {
						"gemini-3-flash": {
							"tokens": {"prompt": 10, "candidates": 5, "total": 15}
						}
					}
				}
			}`,
			wantContent:    "Fallback text",
			wantInputToks:  10,
			wantOutputToks: 5,
			wantTotalToks:  15,
		},
		{
			name: "response_priority_over_text",
			input: `{
				"response": "Primary response",
				"text": "Fallback text"
			}`,
			wantContent: "Primary response",
		},
		{
			name: "stats_with_tools",
			input: `{
				"response": "Used some tools.",
				"stats": {
					"models": {
						"gemini-3-flash": {
							"tokens": {"prompt": 50, "candidates": 25, "total": 75}
						}
					},
					"tools": {
						"totalCalls": 3,
						"totalSuccess": 3,
						"totalFail": 0,
						"totalDurationMs": 500
					}
				}
			}`,
			wantContent:    "Used some tools.",
			wantInputToks:  50,
			wantOutputToks: 25,
			wantTotalToks:  75,
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
	input := `{"response": "Test response"}`

	duration := 250 * time.Millisecond
	resp, err := ParseJSON(input, duration)
	if err != nil {
		t.Fatalf("ParseJSON() error = %v", err)
	}

	// For simple response without stats, metadata may be nil
	// Let's test with candidates format which sets metadata
	input = `{
		"candidates": [{"content": {"parts": [{"text": "Test"}]}, "finishReason": "STOP"}]
	}`
	resp, err = ParseJSON(input, duration)
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
