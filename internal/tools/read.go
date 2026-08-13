package tools

// ReadToolSpec defines the schema for the read_file tool.
var ReadToolSpec = ToolSpec{
	Name:        "read_file",
	Description: "Reads the content of a file within the workspace boundary.",
	RiskLevel:   RiskLevelRead,
}
