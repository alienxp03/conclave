package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	baseprovider "github.com/alienxp03/conclave/provider"
)

type countingProvider struct {
	name   string
	checks int32
}

func (p *countingProvider) Name() string {
	return p.name
}

func (p *countingProvider) Available() bool {
	return true
}

func (p *countingProvider) Execute(ctx context.Context, req *baseprovider.Request) (*baseprovider.Response, error) {
	return &baseprovider.Response{Content: "2"}, nil
}

func (p *countingProvider) HealthCheck(ctx context.Context) baseprovider.HealthStatus {
	atomic.AddInt32(&p.checks, 1)
	return baseprovider.HealthStatus{
		Available:    true,
		ResponseTime: 50 * time.Millisecond,
		CheckedAt:    time.Now(),
	}
}

func TestHandleAPIProviderHealth_UsesCache(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "provider-health.json")
	handler.healthCache = newProviderHealthCache(cachePath, 30*time.Minute)

	prov := &countingProvider{name: "counting"}
	handler.registry.Register(prov)

	req := httptest.NewRequest("GET", "/api/providers/health/counting", nil)
	req.SetPathValue("name", "counting")
	w := httptest.NewRecorder()
	handler.handleAPIProviderHealth(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if payload["name"] != "counting" {
		t.Fatalf("expected name counting, got %v", payload["name"])
	}

	req2 := httptest.NewRequest("GET", "/api/providers/health/counting", nil)
	req2.SetPathValue("name", "counting")
	w2 := httptest.NewRecorder()
	handler.handleAPIProviderHealth(w2, req2)

	if w2.Code != 200 {
		t.Fatalf("expected status 200, got %d", w2.Code)
	}

	if got := atomic.LoadInt32(&prov.checks); got != 1 {
		t.Fatalf("expected 1 health check call, got %d", got)
	}

	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file to be created, got error: %v", err)
	}
}
