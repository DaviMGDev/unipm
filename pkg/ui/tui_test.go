package ui

import (
	"testing"

	"github.com/DaviMGDev/unipm/pkg/adapter"
	tea "github.com/charmbracelet/bubbletea"
)

func TestRunSelection_Empty(t *testing.T) {
	_, err := RunSelection(nil)
	if err == nil {
		t.Error("RunSelection(nil) should return error")
	}
}

func TestRunSelection_SinglePackage(t *testing.T) {
	pkgs := []adapter.Package{
		{Name: "htop", Source: "apt", Version: "3.4.1", Description: "process viewer"},
	}

	pkg, err := RunSelection(pkgs)
	if err != nil {
		t.Fatalf("RunSelection() unexpected error: %v", err)
	}
	if pkg.Name != "htop" {
		t.Errorf("Name = %q, want %q", pkg.Name, "htop")
	}
	if pkg.Source != "apt" {
		t.Errorf("Source = %q, want %q", pkg.Source, "apt")
	}
}

func TestModel_Init(t *testing.T) {
	m := model{packages: []adapter.Package{}, cursor: 0, selected: -1}
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_Update_Quit(t *testing.T) {
	pkgs := []adapter.Package{
		{Name: "htop", Source: "apt", Version: "3.4.1"},
	}

	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
		{"q", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"esc", tea.KeyMsg{Type: tea.KeyEsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{packages: pkgs, cursor: 0, selected: -1}

			newModel, _ := m.Update(tt.msg)
			updated := newModel.(model)

			if !updated.quitting {
				t.Errorf("quitting = false after %s, want true", tt.name)
			}
		})
	}
}

func TestModel_Update_EnterSelects(t *testing.T) {
	pkgs := []adapter.Package{
		{Name: "htop", Source: "apt", Version: "3.4.1"},
		{Name: "httpie", Source: "pypi", Version: "3.2.1"},
	}

	m := model{packages: pkgs, cursor: 1, selected: -1}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(msg)
	updated := newModel.(model)

	if updated.selected != 1 {
		t.Errorf("selected = %d, want 1", updated.selected)
	}
}

func TestModel_Update_CursorMovement(t *testing.T) {
	pkgs := []adapter.Package{
		{Name: "htop", Source: "apt", Version: "3.4.1"},
		{Name: "httpie", Source: "pypi", Version: "3.2.1"},
		{Name: "ripgrep", Source: "brew", Version: "14.1.0"},
	}

	t.Run("down arrow", func(t *testing.T) {
		m := model{packages: pkgs, cursor: 0, selected: -1}
		msg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := m.Update(msg)
		updated := newModel.(model)
		if updated.cursor != 1 {
			t.Errorf("cursor = %d after down, want 1", updated.cursor)
		}
	})

	t.Run("j key", func(t *testing.T) {
		m := model{packages: pkgs, cursor: 0, selected: -1}
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		newModel, _ := m.Update(msg)
		updated := newModel.(model)
		if updated.cursor != 1 {
			t.Errorf("cursor = %d after j, want 1", updated.cursor)
		}
	})

	t.Run("up arrow", func(t *testing.T) {
		m := model{packages: pkgs, cursor: 2, selected: -1}
		msg := tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ := m.Update(msg)
		updated := newModel.(model)
		if updated.cursor != 1 {
			t.Errorf("cursor = %d after up, want 1", updated.cursor)
		}
	})

	t.Run("k key", func(t *testing.T) {
		m := model{packages: pkgs, cursor: 2, selected: -1}
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
		newModel, _ := m.Update(msg)
		updated := newModel.(model)
		if updated.cursor != 1 {
			t.Errorf("cursor = %d after k, want 1", updated.cursor)
		}
	})

	t.Run("down at bottom stays", func(t *testing.T) {
		m := model{packages: pkgs, cursor: 2, selected: -1}
		msg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := m.Update(msg)
		updated := newModel.(model)
		if updated.cursor != 2 {
			t.Errorf("cursor = %d after down at bottom, want 2", updated.cursor)
		}
	})

	t.Run("up at top stays", func(t *testing.T) {
		m := model{packages: pkgs, cursor: 0, selected: -1}
		msg := tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ := m.Update(msg)
		updated := newModel.(model)
		if updated.cursor != 0 {
			t.Errorf("cursor = %d after up at top, want 0", updated.cursor)
		}
	})
}

func TestModel_View(t *testing.T) {
	pkgs := []adapter.Package{
		{Name: "htop", Source: "apt", Version: "3.4.1", Description: "process viewer"},
	}

	m := model{packages: pkgs, cursor: 0, selected: -1}
	view := m.View()

	if view == "" {
		t.Error("View() returned empty string")
	}

	// Check key elements are present
	checks := []string{
		"Select one:",
		"[apt]",
		"htop",
		"3.4.1",
		"process viewer",
		"enter: select",
		"q/esc: cancel",
	}

	for _, check := range checks {
		if !contains(view, check) {
			t.Errorf("View() missing %q", check)
		}
	}
}

func TestModel_View_Quitting(t *testing.T) {
	m := model{quitting: true}
	view := m.View()
	if !contains(view, "cancelled") {
		t.Errorf("View() for quitting state missing 'cancelled': %q", view)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
