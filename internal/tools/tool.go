package tools

import (
	"context"
	"encoding/json"
	"time"
)

// ToolRiskLevel defines the risk level of executing a tool call.
type ToolRiskLevel string

const (
	RiskLevelRead     ToolRiskLevel = "READ_ONLY"       // Low risk: read files, search, list dir
	RiskLevelMutation ToolRiskLevel = "MUTATION"        // Medium risk: write/modify file (transactional)
	RiskLevelShell    ToolRiskLevel = "SHELL_EXEC"      // High risk: execute shell command
	RiskLevelAdmin    ToolRiskLevel = "ADMIN_DANGEROUS" // Critical risk: system level changes
)

// ToolSpec describes a tool available to AI models (JSON Schema parameters).
type ToolSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	RiskLevel   ToolRiskLevel   `json:"risk_level"`
}

// ToolCallSpec represents an invoked tool call proposed by an AI model.
type ToolCallSpec struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult represents the execution result of a tool call.
type ToolResult struct {
	ToolCallID string        `json:"tool_call_id"`
	Output     string        `json:"output"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	IsError    bool          `json:"is_error"`
}

// Tool defines the contract for runtime-owned tool implementations.
type Tool interface {
	Spec() ToolSpec
	ValidateArgs(args json.RawMessage) error
	Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}
