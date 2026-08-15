// Package bootstrap wires concrete model provider implementations into the model factory.
// It is imported exactly once at program startup (cmd/termagent/main.go) to register providers.
package bootstrap

import (
	"fmt"

	"github.com/ilyaskhan/term-agent/internal/model"
	"github.com/ilyaskhan/term-agent/internal/model/anthropic"
	"github.com/ilyaskhan/term-agent/internal/model/gemini"
	"github.com/ilyaskhan/term-agent/internal/model/openai"
)

// Init registers the concrete provider factory into model.DefaultFactory.
// Must be called before any research workflow executes.
func Init() {
	model.DefaultFactory = func(cfg model.ProviderConfig) (model.ModelProvider, error) {
		if err := cfg.Validate(); err != nil {
			return nil, err
		}

		switch cfg.ProviderName {
		case "openai":
			return openai.NewProvider(cfg.APIKey)
		case "anthropic":
			return anthropic.NewProvider(cfg.APIKey)
		case "gemini":
			return gemini.NewProvider(cfg.APIKey)
		default:
			return nil, fmt.Errorf("bootstrap: no factory registered for provider %q", cfg.ProviderName)
		}
	}
}

// BuildProvider is a convenience wrapper that validates, builds, and returns a provider.
func BuildProvider(providerName, modelName, apiKey string) (model.ModelProvider, error) {
	if model.DefaultFactory == nil {
		return nil, fmt.Errorf("bootstrap.Init() has not been called; provider factory is unregistered")
	}
	cfg := model.ProviderConfig{
		ProviderName: providerName,
		Model:        modelName,
		APIKey:       apiKey,
	}
	return model.DefaultFactory(cfg)
}
