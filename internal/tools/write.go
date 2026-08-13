package tools

// WriteToolSpec defines the schema for the write_file tool.
var WriteToolSpec = ToolSpec{
	Name:        "write_file",
	Description: "Proposes writing content to a file via the transactional Mutation Engine.",
	RiskLevel:   RiskLevelMutation,
}
