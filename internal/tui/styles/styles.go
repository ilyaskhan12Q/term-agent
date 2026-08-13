package styles

// DesignTokens holds color palette and visual styling definitions.
type DesignTokens struct {
	PrimaryColor    string
	SecondaryColor  string
	SuccessColor    string
	WarningColor    string
	ErrorColor      string
	BackgroundColor string
}

// DefaultTokens returns the standard term-agent design token system.
func DefaultTokens() DesignTokens {
	return DesignTokens{
		PrimaryColor:    "#7D56F4",
		SecondaryColor:  "#04B575",
		SuccessColor:    "#00D787",
		WarningColor:    "#FFB800",
		ErrorColor:      "#FF5555",
		BackgroundColor: "#1E1E2E",
	}
}
