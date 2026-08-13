package tools

// ShellToolSpec defines the schema for the execute_shell tool.
var ShellToolSpec = ToolSpec{
	Name:        "execute_shell",
	Description: "Proposes executing a shell command subject to Security Policy classification.",
	RiskLevel:   RiskLevelShell,
}
