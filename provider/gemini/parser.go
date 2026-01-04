package gemini

import (
	"encoding/json"
	"time"

	"github.com/alienxp03/conclave/provider"
)

// ModelStats represents per-model statistics from Gemini CLI.
type ModelStats struct {
	API *struct {
		TotalRequests  int   `json:"totalRequests"`
		TotalErrors    int   `json:"totalErrors"`
		TotalLatencyMs int64 `json:"totalLatencyMs"`
	} `json:"api,omitempty"`
	Tokens *struct {
		Prompt     int `json:"prompt"`
		Candidates int `json:"candidates"`
		Total      int `json:"total"`
		Cached     int `json:"cached"`
		Thoughts   int `json:"thoughts"`
		Tool       int `json:"tool"`
	} `json:"tokens,omitempty"`
}

// JSONResponse represents Gemini CLI JSON output.
type JSONResponse struct {
	Response string `json:"response,omitempty"` // Main response text from Gemini CLI
	Stats    *struct {
		Models map[string]*ModelStats `json:"models,omitempty"`
		Tools  *struct {
			TotalCalls      int   `json:"totalCalls"`
			TotalSuccess    int   `json:"totalSuccess"`
			TotalFail       int   `json:"totalFail"`
			TotalDurationMs int64 `json:"totalDurationMs"`
		} `json:"tools,omitempty"`
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

// ParseJSON parses Gemini CLI JSON output.
func ParseJSON(data string, duration time.Duration) (*provider.Response, error) {
	var raw JSONResponse
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return &provider.Response{
			Content: data,
			Raw:     data,
		}, nil
	}

	resp := &provider.Response{
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
			resp.Metadata = &provider.Metadata{
				StopReason: raw.Candidates[0].FinishReason,
				Duration:   duration,
			}
		}
	}

	// Extract metadata from stats if available (Gemini CLI format)
	if raw.Stats != nil && raw.Stats.Models != nil {
		if resp.Metadata == nil {
			resp.Metadata = &provider.Metadata{Duration: duration}
		}
		// Aggregate stats from all models used
		for _, modelStats := range raw.Stats.Models {
			if modelStats.Tokens != nil {
				resp.Metadata.InputTokens += modelStats.Tokens.Prompt
				resp.Metadata.OutputTokens += modelStats.Tokens.Candidates
				resp.Metadata.TotalTokens += modelStats.Tokens.Total
			}
		}
	}

	// Fallback metadata extraction from usageMetadata (traditional API format)
	if raw.UsageMetadata != nil {
		if resp.Metadata == nil {
			resp.Metadata = &provider.Metadata{Duration: duration}
		}
		// Only use if not already populated from stats
		if resp.Metadata.InputTokens == 0 {
			resp.Metadata.InputTokens = raw.UsageMetadata.PromptTokenCount
		}
		if resp.Metadata.OutputTokens == 0 {
			resp.Metadata.OutputTokens = raw.UsageMetadata.CandidatesTokenCount
		}
		if resp.Metadata.TotalTokens == 0 {
			resp.Metadata.TotalTokens = raw.UsageMetadata.TotalTokenCount
		}
	}

	return resp, nil
}
