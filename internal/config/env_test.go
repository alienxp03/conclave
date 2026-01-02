package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	
	content := `
# Comment
KEY1=value1
KEY2="value 2"
KEY3='value 3'
KEY4=value 4 # inline comment
EMPTY=
`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create env file: %v", err)
	}

	env, err := LoadEnv(envFile)
	if err != nil {
		t.Fatalf("LoadEnv failed: %v", err)
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"KEY1", "value1"},
		{"KEY2", "value 2"},
		{"KEY3", "value 3"},
		{"KEY4", "value 4"},
		{"EMPTY", ""},
	}

	for _, tt := range tests {
		if got, ok := env[tt.key]; !ok || got != tt.expected {
			t.Errorf("expected %s=%q, got %q (exists=%v)", tt.key, tt.expected, got, ok)
		}
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	cfg := Default()
	
	env := map[string]string{
		"DEFAULT_STYLE":           "adversarial",
		"DEFAULT_PROVIDER":        "gemini",
		"PROVIDER_CLAUDE_ENABLED": "false",
		"PROVIDER_TIMEOUT":        "60",
		"SERVER_PORT":             "9090",
	}

	ApplyEnvOverrides(cfg, env)

	if cfg.Defaults.Style != "adversarial" {
		t.Errorf("expected style adversarial, got %s", cfg.Defaults.Style)
	}
	if cfg.Defaults.Provider != "gemini" {
		t.Errorf("expected provider gemini, got %s", cfg.Defaults.Provider)
	}
	if cfg.Providers["claude"].Enabled {
		t.Errorf("expected claude disabled")
	}
	if cfg.Providers["gemini"].Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", cfg.Providers["gemini"].Timeout)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
}
