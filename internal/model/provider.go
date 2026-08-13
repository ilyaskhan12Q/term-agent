package model

import (
	"context"
)

// ModelProvider defines the provider-agnostic interface for LLM completions.
type ModelProvider interface {
	Name() string
	GenerateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	Capabilities() ModelCapabilities
}
