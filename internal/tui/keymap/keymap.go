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
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
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
			key.WithKeys("alt+1"),
			key.WithHelp("alt+1", "agent view"),
		),
		SelectTab2: key.NewBinding(
			key.WithKeys("alt+2"),
			key.WithHelp("alt+2", "plan view"),
		),
		SelectTab3: key.NewBinding(
			key.WithKeys("alt+3"),
			key.WithHelp("alt+3", "diff view"),
		),
		SelectTab4: key.NewBinding(
			key.WithKeys("alt+4"),
			key.WithHelp("alt+4", "log view"),
		),
		SelectTab5: key.NewBinding(
			key.WithKeys("alt+5"),
			key.WithHelp("alt+5", "research view"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit prompt"),
		),
		ApproveDiff: key.NewBinding(
			key.WithKeys("ctrl+y", "alt+y"),
			key.WithHelp("ctrl+y", "approve diff"),
		),
		RejectDiff: key.NewBinding(
			key.WithKeys("ctrl+n", "alt+n", "esc"),
			key.WithHelp("esc/ctrl+n", "reject diff"),
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
