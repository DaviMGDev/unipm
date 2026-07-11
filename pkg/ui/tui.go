// Package ui provides the Bubbletea TUI components for unipm.
// Currently implements the collision-resolution prompt shown when
// a package exists in multiple sources during `unipm install`.
package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DaviMGDev/unipm/pkg/adapter"
	tea "github.com/charmbracelet/bubbletea"
)

// model is the Bubbletea model for the collision-resolution TUI.
type model struct {
	packages []adapter.Package
	cursor   int
	selected int // -1 means no selection (cancelled)
	quitting bool
}

// RunSelection displays an interactive TUI list of packages and returns
// the selected package or an error if the user cancels.
//
// Packages are displayed alphabetically by source name (per ADR-0002).
// Keys:
//   - j / ↓ : move cursor down
//   - k / ↑ : move cursor up
//   - enter  : confirm selection
//   - q / esc / ctrl+c : cancel
func RunSelection(packages []adapter.Package) (*adapter.Package, error) {
	if len(packages) == 0 {
		return nil, fmt.Errorf("no packages to select from")
	}

	if len(packages) == 1 {
		return &packages[0], nil
	}

	// Sort by source name then package name for deterministic display
	sort.Slice(packages, func(i, j int) bool {
		if packages[i].Source != packages[j].Source {
			return packages[i].Source < packages[j].Source
		}
		return packages[i].Name < packages[j].Name
	})

	m := model{
		packages: packages,
		cursor:   0,
		selected: -1,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	m = finalModel.(model)
	if m.selected < 0 {
		return nil, fmt.Errorf("installation cancelled")
	}

	return &m.packages[m.selected], nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			m.selected = m.cursor
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.packages)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Installation cancelled.\n"
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("  Multiple packages found. Select one:\n\n")

	for i, pkg := range m.packages {
		cursor := "  "
		if m.cursor == i {
			cursor = "❯ "
		}

		line := fmt.Sprintf("%s[%s] %s (%s)",
			cursor, pkg.Source, pkg.Name, pkg.Version)

		if pkg.Description != "" {
			line += fmt.Sprintf(" — %s", pkg.Description)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("  ↑/↓ or j/k: navigate  •  enter: select  •  q/esc: cancel\n")

	return b.String()
}
