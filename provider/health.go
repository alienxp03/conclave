package provider

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	// HealthCheckPrompt is the prompt sent to providers for health checks.
	HealthCheckPrompt = "1+1? One digit answer only"
)

// HealthCheckWithExecute runs a provider health check using the provided execute function.
func HealthCheckWithExecute(ctx context.Context, model string, exec func(context.Context, *Request) (*Response, error)) HealthStatus {
	start := time.Now()

	// 10 second timeout for health check
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req := &Request{
		Prompt: HealthCheckPrompt,
		Model:  model,
	}

	resp, err := exec(ctx, req)
	elapsed := time.Since(start)
	if err != nil {
		return HealthStatus{
			Available:    false,
			ResponseTime: elapsed,
			Error:        err.Error(),
			CheckedAt:    time.Now(),
		}
	}
	if resp == nil {
		return HealthStatus{
			Available:    false,
			ResponseTime: elapsed,
			Error:        "empty response",
			CheckedAt:    time.Now(),
		}
	}

	if err := validateHealthResponse(resp.Content); err != nil {
		return HealthStatus{
			Available:    false,
			ResponseTime: elapsed,
			Error:        err.Error(),
			CheckedAt:    time.Now(),
		}
	}

	return HealthStatus{
		Available:    true,
		ResponseTime: elapsed,
		CheckedAt:    time.Now(),
	}
}

func validateHealthResponse(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "2" {
		return nil
	}
	if trimmed == "" {
		return fmt.Errorf("unexpected response: empty")
	}
	if len(trimmed) > 120 {
		trimmed = trimmed[:120] + "..."
	}
	return fmt.Errorf("unexpected response: %q", trimmed)
}
