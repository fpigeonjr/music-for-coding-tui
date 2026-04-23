package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Styles ──────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF")) // cyan

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))
)

// ─── Model ───────────────────────────────────────────────────────────────────

type model struct {
	width  int
	height int
}

func initialModel() model {
	return model{}
}

// ─── Init ────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return nil
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m model) View() string {
	title := titleStyle.Render("music-for-coding-tui")
	subtitle := "Phase 1 scaffold — audio plumbing coming next"
	help := helpStyle.Render("q quit")

	return fmt.Sprintf("\n  %s\n  %s\n\n  %s\n", title, subtitle, help)
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
