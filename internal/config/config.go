// Package config handles application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alienxp03/conclave/internal/core"
	intprovider "github.com/alienxp03/conclave/internal/provider"
	"github.com/alienxp03/conclave/provider"
	"github.com/alienxp03/conclave/provider/claude"
	"github.com/alienxp03/conclave/provider/gemini"
	"github.com/alienxp03/conclave/provider/generic"
	"github.com/alienxp03/conclave/provider/openai"
	"github.com/alienxp03/conclave/provider/opencode"
	"github.com/alienxp03/conclave/provider/qwen"
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
	MaxRetries   int           `yaml:"max_retries,omitempty"`
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
			MaxRetries:   2,
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

// ToProviderConfig converts a ProviderConfig to provider.Config.
func (p ProviderConfig) ToProviderConfig(name string) provider.Config {
	return provider.Config{
		Name:         name,
		Command:      p.Command,
		Args:         p.Args,
		DefaultModel: p.DefaultModel,
		Models:       p.Models,
		Timeout:      p.Timeout,
		MaxRetries:   p.MaxRetries,
	}
}

// createProviderFromName creates a provider instance based on the provider name.
func createProviderFromName(name string, cfg provider.Config) (provider.Provider, error) {
	switch name {
	case "claude":
		return claude.New(cfg), nil
	case "gemini":
		return gemini.New(cfg), nil
	case "openai", "codex":
		// Support both "openai" and "codex" for backward compatibility
		if cfg.Name == "" || cfg.Name == "codex" {
			cfg.Name = "codex" // Keep "codex" as the name for backward compatibility
		}
		return openai.New(cfg), nil
	case "qwen":
		return qwen.New(cfg), nil
	case "opencode":
		return opencode.New(cfg), nil
	case "mock":
		return intprovider.NewMockProvider(cfg), nil
	default:
		// Unknown providers fall back to generic
		return generic.New(cfg), nil
	}
}

// CreateProvider creates a provider instance from this configuration.
func (c *Config) CreateProvider(name string) (provider.Provider, error) {
	provCfg, ok := c.GetProvider(name)
	if !ok {
		return nil, fmt.Errorf("provider %s not found in config", name)
	}
	if !provCfg.Enabled {
		return nil, fmt.Errorf("provider %s is disabled", name)
	}
	return createProviderFromName(name, provCfg.ToProviderConfig(name))
}

// CreateRegistry creates a provider registry from this configuration.
func (c *Config) CreateRegistry() (*intprovider.Registry, error) {
	registry := intprovider.NewRegistry()

	for name, provCfg := range c.Providers {
		if !provCfg.Enabled {
			continue
		}

		p, err := createProviderFromName(name, provCfg.ToProviderConfig(name))
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %w", name, err)
		}

		registry.Register(p)
	}

	return registry, nil
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
    max_retries: 2          # Retry failed commands (default: 2, total 3 attempts)
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
