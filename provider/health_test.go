package provider

import (
	"context"
	"testing"
)

func TestHealthCheckWithExecuteSuccess(t *testing.T) {
	var prompt string
	status := HealthCheckWithExecute(context.Background(), "", func(ctx context.Context, req *Request) (*Response, error) {
		prompt = req.Prompt
		return &Response{Content: "2"}, nil
	})

	if prompt != HealthCheckPrompt {
		t.Fatalf("expected prompt %q, got %q", HealthCheckPrompt, prompt)
	}
	if !status.Available {
		t.Fatalf("expected available=true, got false with error %q", status.Error)
	}
}

func TestHealthCheckWithExecuteInvalidResponse(t *testing.T) {
	status := HealthCheckWithExecute(context.Background(), "", func(ctx context.Context, req *Request) (*Response, error) {
		return &Response{Content: "two"}, nil
	})

	if status.Available {
		t.Fatalf("expected available=false, got true")
	}
	if status.Error == "" {
		t.Fatalf("expected error message for invalid response")
	}
}
