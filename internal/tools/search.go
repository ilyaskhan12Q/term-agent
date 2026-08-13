package tools

// SearchToolSpec defines the schema for the search_workspace tool.
var SearchToolSpec = ToolSpec{
	Name:        "search_workspace",
	Description: "Searches for text patterns across workspace files.",
	RiskLevel:   RiskLevelRead,
}
