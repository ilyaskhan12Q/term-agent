package model

import (
	"github.com/ilyaskhan/term-agent/internal/tools"
)

// TokenUsage holds token consumption metadata.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// CompletionResponse contains output from a ModelProvider.
type CompletionResponse struct {
	Content   string
	ToolCalls []tools.ToolCallSpec
	Usage     TokenUsage
	Model     string
}
