package model

import (
	"github.com/ilyaskhan/term-agent/internal/tools"
)

// MessageRole represents the role of a message sender.
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// Message represents a prompt message in a completion request.
type Message struct {
	Role       MessageRole
	Content    string
	ToolCallID string
}

// CompletionRequest encapsulates parameters sent to a ModelProvider.
type CompletionRequest struct {
	Model       string
	Messages    []Message
	Tools       []tools.ToolSpec
	Temperature float64
	MaxTokens   int
}
