package keymap

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines keyboard shortcuts for the term-agent TUI.
type KeyMap struct {
	Quit        key.Binding
	TabNext     key.Binding
	TabPrev     key.Binding
	SelectTab1  key.Binding
	SelectTab2  key.Binding
	SelectTab3  key.Binding
	SelectTab4  key.Binding
	SelectTab5  key.Binding
	Submit      key.Binding
	ApproveDiff key.Binding
	RejectDiff  key.Binding
	CancelAgent key.Binding
	Clear       key.Binding
}

// DefaultKeyMap returns the configured keyboard shortcuts.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("q / ctrl+c", "quit"),
		),
		TabNext: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next view"),
		),
		TabPrev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev view"),
		),
		SelectTab1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "agent view"),
		),
		SelectTab2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "plan view"),
		),
		SelectTab3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "diff view"),
		),
		SelectTab4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "log view"),
		),
		SelectTab5: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "research view"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit prompt"),
		),
		ApproveDiff: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "approve diff"),
		),
		RejectDiff: key.NewBinding(
			key.WithKeys("n", "esc"),
			key.WithHelp("n/esc", "reject diff"),
		),
		CancelAgent: key.NewBinding(
			key.WithKeys("ctrl+k"),
			key.WithHelp("ctrl+k", "cancel task"),
		),
		Clear: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear view"),
		),
	}
}
