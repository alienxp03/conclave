package claude

import (
	"encoding/json"
	"time"

	"github.com/alienxp03/conclave/provider"
)

// JSONResponse represents Claude CLI JSON output.
type JSONResponse struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`
	Role    string `json:"role,omitempty"`
	Model   string `json:"model,omitempty"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
	Usage      *struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	} `json:"usage,omitempty"`
	Result      string  `json:"result,omitempty"`    // For simpler responses
	SessionID   string  `json:"session_id,omitempty"`
	DurationMs  int64   `json:"duration_ms,omitempty"`  // CLI adds duration at top level
	NumTurns    int     `json:"num_turns,omitempty"`
	TotalCostUSD float64 `json:"total_cost_usd,omitempty"`
}

// ParseJSON parses Claude CLI JSON output.
func ParseJSON(data string, duration time.Duration) (*provider.Response, error) {
	var raw JSONResponse
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		// Not JSON, return as plain text response
		return &provider.Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	resp := &provider.Response{
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
		// Calculate total input including cache tokens
		totalInputTokens := raw.Usage.InputTokens + raw.Usage.CacheCreationInputTokens + raw.Usage.CacheReadInputTokens
		
		// Use CLI's duration_ms if available, otherwise fall back to measured duration
		actualDuration := duration
		if raw.DurationMs > 0 {
			actualDuration = time.Duration(raw.DurationMs) * time.Millisecond
		}
		
		resp.Metadata = &provider.Metadata{
			InputTokens:  totalInputTokens,
			OutputTokens: raw.Usage.OutputTokens,
			TotalTokens:  totalInputTokens + raw.Usage.OutputTokens,
			StopReason:   raw.StopReason,
			SessionID:    raw.SessionID,
			Duration:     actualDuration,
		}
	}

	return resp, nil
}
