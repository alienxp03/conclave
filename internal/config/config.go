// Package config handles application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alienxp03/conclave/internal/core"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Defaults  DefaultsConfig            `yaml:"defaults"`
	Providers map[string]ProviderConfig `yaml:"providers"`
	Personas  []PersonaConfig           `yaml:"personas,omitempty"`
	Server    ServerConfig              `yaml:"server,omitempty"`
}

// ServerConfig holds server settings.
type ServerConfig struct {
	Port int `yaml:"port"`
}

// DefaultsConfig holds default settings.
type DefaultsConfig struct {
	Style    string `yaml:"style"`
	MaxTurns int    `yaml:"max_turns"`
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
}

// ProviderConfig holds provider-specific settings.
type ProviderConfig struct {
	Command      string        `yaml:"command"`
	Args         []string      `yaml:"args,omitempty"`
	DefaultModel string        `yaml:"default_model,omitempty"`
	Models       []string      `yaml:"models,omitempty"`
	Timeout      time.Duration `yaml:"timeout,omitempty"`
	Enabled      bool          `yaml:"enabled"`
}

// PersonaConfig holds custom persona definitions.
type PersonaConfig struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	SystemPrompt string `yaml:"system_prompt"`
}

// Default returns the default configuration.
func Default() *Config {
	providers := make(map[string]ProviderConfig)
	for name, cmd := range core.DefaultCommandForProvider {
		providers[name] = ProviderConfig{
			Command:      cmd,
			Args:         core.DefaultArgsForProvider[name],
			DefaultModel: core.DefaultModelForProvider[name],
			Models:       core.DefaultModelsForProvider[name],
			Timeout:      5 * time.Minute,
			Enabled:      true,
		}
	}

	// Adjust mock timeout
	if mock, ok := providers["mock"]; ok {
		mock.Timeout = 1 * time.Minute
		providers["mock"] = mock
	}

	return &Config{
		Defaults: DefaultsConfig{
			Style:    "collaborative",
			MaxTurns: 5,
			Provider: "claude",
			Model:    "",
		},
		Server: ServerConfig{
			Port: 8182,
		},
		Providers: providers,
	}
}

// Load loads configuration from the default path.
func Load() (*Config, error) {
	return LoadFrom(DefaultConfigPath())
}

// LoadFrom loads configuration from a specific path.
func LoadFrom(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// No config file, proceed with defaults
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Merge with defaults for any missing providers
	defaultCfg := Default()
	for name, defaultProvider := range defaultCfg.Providers {
		if _, exists := cfg.Providers[name]; !exists {
			cfg.Providers[name] = defaultProvider
		}
	}

	// Apply .env overrides if file exists
	if env, err := LoadEnv(".env"); err == nil {
		ApplyEnvOverrides(cfg, env)
	}

	return cfg, nil
}

// Save saves the configuration to the default path.
func (c *Config) Save() error {
	return c.SaveTo(DefaultConfigPath())
}

// SaveTo saves the configuration to a specific path.
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetProvider returns the configuration for a provider.
func (c *Config) GetProvider(name string) (ProviderConfig, bool) {
	p, ok := c.Providers[name]
	return p, ok
}

// DefaultConfigPath returns the default configuration file path.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "conclave.yaml"
	}
	return filepath.Join(home, ".conclave", "config.yaml")
}

// GenerateExample generates an example configuration file.
func GenerateExample() string {
	example := `# conclave configuration file
# Place this file at ~/.conclave/config.yaml

defaults:
  style: collaborative      # Default debate style
  max_turns: 5              # Default turns per agent
  provider: claude          # Default provider
  model: ""                 # Default model (empty = provider default)

providers:
  claude:
    command: claude
    args: ["--print"]
    default_model: ""       # e.g., "sonnet", "opus", "haiku"
    models: ["opus", "sonnet", "haiku"]
    timeout: 5m
    enabled: true

  codex:
    command: codex
    args: []
    default_model: ""
    models: ["gpt-4", "gpt-4o", "gpt-3.5-turbo"]
    timeout: 5m
    enabled: true

  gemini:
    command: gemini
    args: []
    default_model: ""
    models: ["pro", "flash", "ultra"]
    timeout: 5m
    enabled: true

  qwen:
    command: qwen
    args: []
    default_model: ""
    models: ["qwen-turbo", "qwen-plus", "qwen-max"]
    timeout: 5m
    enabled: true

# Custom personas (optional)
personas:
  - id: security_expert
    name: Security Expert
    description: Focuses on security implications and vulnerabilities
    system_prompt: |
      You are a security-focused debater. Your approach:
      - Identify potential security vulnerabilities
      - Consider attack vectors and threat models
      - Prioritize secure-by-default solutions
      - Question trust assumptions
`
	return example
}
