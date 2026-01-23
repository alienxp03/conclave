package handlers

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alienxp03/conclave/provider"
)

const (
	providerHealthCacheFilename = "conclave-provider-health.json"
	providerHealthCacheTTL      = 30 * time.Minute
)

type providerHealthCache struct {
	mu     sync.Mutex
	path   string
	ttl    time.Duration
	loaded bool
	data   map[string]provider.HealthStatus
}

func newProviderHealthCache(path string, ttl time.Duration) *providerHealthCache {
	if ttl <= 0 {
		ttl = providerHealthCacheTTL
	}
	return &providerHealthCache{
		path: path,
		ttl:  ttl,
		data: make(map[string]provider.HealthStatus),
	}
}

func defaultProviderHealthCachePath() string {
	return filepath.Join(os.TempDir(), providerHealthCacheFilename)
}

func (c *providerHealthCache) GetFresh(name string) (provider.HealthStatus, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ensureLoaded()
	status, ok := c.data[name]
	if !ok {
		return provider.HealthStatus{}, false
	}
	if !status.Available {
		return provider.HealthStatus{}, false
	}
	if status.CheckedAt.IsZero() {
		return provider.HealthStatus{}, false
	}
	if time.Since(status.CheckedAt) > c.ttl {
		return provider.HealthStatus{}, false
	}
	return status, true
}

func (c *providerHealthCache) Set(name string, status provider.HealthStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ensureLoaded()
	c.data[name] = status
	c.persist()
}

func (c *providerHealthCache) ensureLoaded() {
	if c.loaded {
		return
	}
	c.loaded = true

	data, err := os.ReadFile(c.path)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("Failed to read provider health cache", "path", c.path, "error", err)
		}
		return
	}

	if err := json.Unmarshal(data, &c.data); err != nil {
		slog.Warn("Failed to parse provider health cache", "path", c.path, "error", err)
		c.data = make(map[string]provider.HealthStatus)
	}
}

func (c *providerHealthCache) persist() {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		slog.Warn("Failed to create provider health cache directory", "path", c.path, "error", err)
		return
	}

	payload, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		slog.Warn("Failed to encode provider health cache", "path", c.path, "error", err)
		return
	}

	if err := os.WriteFile(c.path, payload, 0o644); err != nil {
		slog.Warn("Failed to write provider health cache", "path", c.path, "error", err)
	}
}
