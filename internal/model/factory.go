// Package model provides provider-agnostic LLM abstractions and the provider factory.
package model

import (
	"fmt"
)

// SupportedProviders lists all provider names that can be instantiated by the factory.
var SupportedProviders = []string{"openai", "anthropic", "gemini"}

// ProviderConfig holds the configuration required to build a ModelProvider.
type ProviderConfig struct {
	// ProviderName is the canonical provider identifier ("openai", "anthropic", "gemini").
	ProviderName string
	// Model is the model identifier string (e.g. "gpt-4o", "claude-sonnet-4-5", "gemini-1.5-pro").
	Model string
	// APIKey is the provider API key. Never written to config files.
	APIKey string
}

// Validate checks that all required fields are present and valid for research mode.
// Returns a descriptive error suitable for display to the user.
func (c *ProviderConfig) Validate() error {
	if c.ProviderName == "" {
		return fmt.Errorf("provider is not set; use --provider or TERMAGENT_PROVIDER to select one of: %v", SupportedProviders)
	}

	known := false
	for _, p := range SupportedProviders {
		if p == c.ProviderName {
			known = true
			break
		}
	}
	if !known {
		return fmt.Errorf("unknown provider %q; supported providers are: %v", c.ProviderName, SupportedProviders)
	}

	if c.Model == "" {
		return fmt.Errorf("model is not set; use --model or TERMAGENT_MODEL to specify a model for provider %q", c.ProviderName)
	}

	if c.APIKey == "" {
		envVar := providerEnvVar(c.ProviderName)
		return fmt.Errorf("API key for provider %q is missing; set environment variable %s", c.ProviderName, envVar)
	}

	return nil
}

// providerEnvVar returns the expected environment variable name for a provider's API key.
func providerEnvVar(provider string) string {
	switch provider {
	case "openai":
		return "OPENAI_API_KEY"
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	default:
		return "PROVIDER_API_KEY"
	}
}

// ProviderFactory is a function type that constructs a ModelProvider from a ProviderConfig.
// The factory imports concrete provider packages to avoid import cycles.
type ProviderFactory func(cfg ProviderConfig) (ModelProvider, error)

// DefaultFactory is the production factory that builds real provider clients.
// It must be set by main (or the research workflow bootstrap) to inject concrete implementations.
// This avoids circular imports between the model package and concrete provider packages.
var DefaultFactory ProviderFactory
