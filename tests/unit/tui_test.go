package unit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilyaskhan/term-agent/internal/tui"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

func TestTUI_ModelInitialization(t *testing.T) {
	m := tui.NewModel(nil)

	if m.ActiveView != tui.ViewAgentView {
		t.Errorf("expected initial active view ViewAgentView, got %s", m.ActiveView)
	}

	if m.Width != 120 || m.Height != 35 {
		t.Errorf("expected dimensions 120x35, got %dx%d", m.Width, m.Height)
	}

	if m.Header == nil || m.StatusBar == nil || m.Prompt == nil {
		t.Fatal("expected non-nil TUI subcomponents")
	}
}

func TestTUI_TabNavigation(t *testing.T) {
	m := tui.NewModel(nil)

	// Simulate key press
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	updated, ok := newModel.(tui.Model)
	if !ok {
		t.Fatal("failed type assertion to tui.Model")
	}

	if updated.ActiveView != tui.ViewPlanView {
		t.Errorf("expected ViewPlanView after pressing 2, got %s", updated.ActiveView)
	}

	// Press 5 for Research View
	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("5")})
	updated = newModel.(tui.Model)
	if updated.ActiveView != tui.ViewResearchView {
		t.Errorf("expected ViewResearchView after pressing 5, got %s", updated.ActiveView)
	}
}

func TestTUI_DiffApprovalRejection(t *testing.T) {
	m := tui.NewModel(nil)
	m.ActiveView = tui.ViewDiffView
	m.DiffView.SetDiff("tx-test-100", "test.go", "+ new line")

	if !m.DiffView.HasPending {
		t.Fatal("expected pending diff")
	}

	// Press 'y' to approve
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	updated := newModel.(tui.Model)

	if updated.DiffView.HasPending {
		t.Error("expected HasPending to be false after approval")
	}
}

func TestTUI_RenderingDoesNotPanic(t *testing.T) {
	m := tui.NewModel(nil)
	views := []tui.ViewMode{
		tui.ViewAgentView,
		tui.ViewPlanView,
		tui.ViewDiffView,
		tui.ViewLogView,
		tui.ViewResearchView,
	}

	for _, v := range views {
		m.ActiveView = v
		output := m.View()
		if len(output) == 0 {
			t.Errorf("expected non-empty rendered string for view %s", v)
		}
	}
}

func TestTUI_DesignTokens(t *testing.T) {
	tokens := styles.DefaultTokens()
	if tokens.PrimaryColor == "" || tokens.SecondaryColor == "" {
		t.Error("expected non-empty design token color hex values")
	}
}
