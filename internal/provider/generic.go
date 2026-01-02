package provider

import (
	"context"
	"strings"

	"github.com/alienxp03/conclave/internal/config"
)

// GenericProvider is a configurable provider for custom CLI tools.
type GenericProvider struct {
	BaseProvider
}

// NewGenericProvider creates a generic provider from config.
func NewGenericProvider(name string, cfg config.ProviderConfig) *GenericProvider {
	displayName := strings.Title(name)
	return &GenericProvider{
		BaseProvider: NewBaseProvider(name, displayName, cfg),
	}
}

// Generate sends a prompt and returns the response.
func (p *GenericProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithModel(ctx, prompt, p.defaultModel)
}

// GenerateWithModel sends a prompt with a specific model.
func (p *GenericProvider) GenerateWithModel(ctx context.Context, prompt, model string) (string, error) {
	args := []string{}

	// Add model flag if specified
	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)
	return p.Execute(ctx, args...)
}
