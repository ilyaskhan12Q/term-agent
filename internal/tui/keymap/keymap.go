package keymap

// KeyBinding represents a keyboard shortcut binding.
type KeyBinding struct {
	Keys        []string
	Description string
}

// KeyMap defines the keybindings contract for term-agent TUI.
type KeyMap struct {
	Quit        KeyBinding
	ApproveDiff KeyBinding
	RejectDiff  KeyBinding
	TogglePlan  KeyBinding
	CancelAgent KeyBinding
}

// DefaultKeyMap returns the default keyboard shortcuts.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:        KeyBinding{Keys: []string{"ctrl+c", "q"}, Description: "Quit application"},
		ApproveDiff: KeyBinding{Keys: []string{"y", "enter"}, Description: "Approve transaction diff"},
		RejectDiff:  KeyBinding{Keys: []string{"n", "esc"}, Description: "Reject transaction diff"},
		TogglePlan:  KeyBinding{Keys: []string{"tab"}, Description: "Toggle task plan panel"},
		CancelAgent: KeyBinding{Keys: []string{"ctrl+k"}, Description: "Cancel current agent execution"},
	}
}
