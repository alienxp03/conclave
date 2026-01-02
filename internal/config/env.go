package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadEnv reads a .env file and returns a map of key-value pairs.
// It ignores comments (starting with #) and empty lines.
func LoadEnv(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	env := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove inline comments
		if idx := strings.Index(value, " #"); idx != -1 {
			value = strings.TrimSpace(value[:idx])
		}

		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || 
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		env[key] = value
	}

	return env, scanner.Err()
}

// ApplyEnvOverrides updates the configuration based on environment variables.
func ApplyEnvOverrides(cfg *Config, env map[string]string) {
	// Server
	if val, ok := env["SERVER_PORT"]; ok {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.Server.Port = port
		}
	}

	// Defaults
	if val, ok := env["DEFAULT_STYLE"]; ok {
		cfg.Defaults.Style = val
	}
	if val, ok := env["DEFAULT_PROVIDER"]; ok {
		cfg.Defaults.Provider = val
	}
	// Model isn't in .env.example but logic implies it could be
	if val, ok := env["DEFAULT_MODEL"]; ok {
		cfg.Defaults.Model = val
	}

	// Provider Enablement
	for name, provider := range cfg.Providers {
		envKey := fmt.Sprintf("PROVIDER_%s_ENABLED", strings.ToUpper(name))
		if val, ok := env[envKey]; ok {
			if boolVal, err := strconv.ParseBool(val); err == nil {
				provider.Enabled = boolVal
				cfg.Providers[name] = provider
			}
		}

		// Timeout
		if val, ok := env["PROVIDER_TIMEOUT"]; ok {
			if seconds, err := strconv.Atoi(val); err == nil {
				provider.Timeout = time.Duration(seconds) * time.Second
				cfg.Providers[name] = provider
			} else if duration, err := time.ParseDuration(val); err == nil {
				provider.Timeout = duration
				cfg.Providers[name] = provider
			}
		}
	}
}
