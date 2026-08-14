package styles

import "github.com/charmbracelet/lipgloss"

// DesignTokens holds color palette and visual styling definitions.
type DesignTokens struct {
	PrimaryColor    string
	SecondaryColor  string
	SuccessColor    string
	WarningColor    string
	ErrorColor      string
	BackgroundColor string
	SurfaceColor    string
	MutedColor      string
	BorderColor     string
	TextColor       string
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
		SurfaceColor:    "#2A2A3D",
		MutedColor:      "#6C7086",
		BorderColor:     "#45475A",
		TextColor:       "#CDD6F4",
	}
}

// Styles encapsulates pre-compiled Lip Gloss styles for TUI components.
type Styles struct {
	Tokens DesignTokens

	// Navigation & Header
	HeaderTitle  lipgloss.Style
	ActiveTab    lipgloss.Style
	InactiveTab  lipgloss.Style
	TabSeparator lipgloss.Style
	HeaderBar    lipgloss.Style

	// Panel & Boxes
	PanelBorder lipgloss.Style
	PanelTitle  lipgloss.Style
	Badge       lipgloss.Style
	BadgeActive lipgloss.Style

	// Status Bar
	StatusBar     lipgloss.Style
	StatusItem    lipgloss.Style
	StatusVal     lipgloss.Style
	StatusBusy    lipgloss.Style
	StatusSuccess lipgloss.Style

	// Views & Content
	ThoughtBox   lipgloss.Style
	ToolBox      lipgloss.Style
	DiffAdd      lipgloss.Style
	DiffRemove   lipgloss.Style
	DiffHeader   lipgloss.Style
	DiffContext  lipgloss.Style
	TaskPending  lipgloss.Style
	TaskRunning  lipgloss.Style
	TaskComplete lipgloss.Style
	TaskFailed   lipgloss.Style

	// Prompt Box
	PromptBorder lipgloss.Style
	PromptPrefix lipgloss.Style
}

// NewStyles constructs a compiled Styles instance based on design tokens.
func NewStyles() *Styles {
	t := DefaultTokens()

	return &Styles{
		Tokens: t,

		HeaderTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.PrimaryColor)).
			Padding(0, 1),

		ActiveTab: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(t.PrimaryColor)).
			Padding(0, 1),

		InactiveTab: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.MutedColor)).
			Background(lipgloss.Color(t.SurfaceColor)).
			Padding(0, 1),

		TabSeparator: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.BorderColor)),

		HeaderBar: lipgloss.NewStyle().
			Background(lipgloss.Color(t.SurfaceColor)).
			Padding(0, 1),

		PanelBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.BorderColor)).
			Padding(0, 1),

		PanelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.PrimaryColor)).
			MarginBottom(1),

		Badge: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.MutedColor)).
			Background(lipgloss.Color(t.SurfaceColor)).
			Padding(0, 1),

		BadgeActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(t.SecondaryColor)).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color(t.SurfaceColor)).
			Foreground(lipgloss.Color(t.TextColor)).
			Padding(0, 1),

		StatusItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.MutedColor)).
			MarginRight(1),

		StatusVal: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.PrimaryColor)),

		StatusBusy: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.WarningColor)),

		StatusSuccess: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.SuccessColor)),

		ThoughtBox: lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color(t.MutedColor)).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color(t.PrimaryColor)).
			PaddingLeft(1).
			MarginLeft(1),

		ToolBox: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.SecondaryColor)).
			Background(lipgloss.Color(t.SurfaceColor)).
			Padding(0, 1),

		DiffAdd: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.SuccessColor)),

		DiffRemove: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.ErrorColor)),

		DiffHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.WarningColor)),

		DiffContext: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.MutedColor)),

		TaskPending: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.MutedColor)),

		TaskRunning: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.WarningColor)),

		TaskComplete: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.SuccessColor)),

		TaskFailed: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.ErrorColor)),

		PromptBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.PrimaryColor)).
			Padding(0, 1),

		PromptPrefix: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.PrimaryColor)).
			MarginRight(1),
	}
}
