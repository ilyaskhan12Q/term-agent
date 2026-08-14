package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// Prompt Component manages user prompt text input.
type Prompt struct {
	textarea textarea.Model
	styles   *styles.Styles
	width    int
}

// NewPrompt constructs Prompt component with default textarea settings.
func NewPrompt(s *styles.Styles) *Prompt {
	ta := textarea.New()
	ta.Placeholder = "Ask term-agent to research a topic, build code, or run a workflow..."
	ta.Focus()
	ta.CharLimit = 2000
	ta.SetWidth(80)
	ta.SetHeight(2)
	ta.ShowLineNumbers = false

	return &Prompt{
		textarea: ta,
		styles:   s,
		width:    80,
	}
}

// SetWidth resizes text area.
func (p *Prompt) SetWidth(w int) {
	p.width = w
	p.textarea.SetWidth(w - 6)
}

// Focus enables input focus.
func (p *Prompt) Focus() tea.Cmd {
	return p.textarea.Focus()
}

// Blur disables input focus.
func (p *Prompt) Blur() {
	p.textarea.Blur()
}

// Value gets current text.
func (p *Prompt) Value() string {
	return p.textarea.Value()
}

// Reset clears input content.
func (p *Prompt) Reset() {
	p.textarea.Reset()
}

// Update handles keyboard messages.
func (p *Prompt) Update(msg tea.Msg) (*Prompt, tea.Cmd) {
	var cmd tea.Cmd
	p.textarea, cmd = p.textarea.Update(msg)
	return p, cmd
}

// Render renders prompt input widget.
func (p *Prompt) Render() string {
	prefix := p.styles.PromptPrefix.Render("❯")
	return p.styles.PromptBorder.Render(prefix + p.textarea.View())
}
