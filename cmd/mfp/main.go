package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
)

// ─── Hard-coded Phase 1 episode ──────────────────────────────────────────────

const (
	episodeURL   = "https://datashat.net/music_for_programming_78-datassette.mp3"
	episodeTitle = "Episode 78: Datassette"
	seekDelta    = 30.0 // seconds
	tickInterval = 250 * time.Millisecond
)

// ─── Messages ────────────────────────────────────────────────────────────────

type tickMsg time.Time

type playerReadyMsg struct{ p *player.Player }

type playerErrMsg struct{ err error }

type stateMsg player.State

// ─── Styles ──────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF"))

	playingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF87"))

	pausedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444"))

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))
)

// ─── Model ───────────────────────────────────────────────────────────────────

type model struct {
	width  int
	height int

	pl    *player.Player
	state player.State

	// loading = player not ready yet; err != nil = fatal error
	loading bool
	err     error
}

func initialModel() model {
	return model{loading: true}
}

// ─── Init ────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return spawnPlayer()
}

// spawnPlayer starts mpv in a background goroutine and returns the handle.
func spawnPlayer() tea.Cmd {
	return func() tea.Msg {
		p, err := player.New()
		if err != nil {
			return playerErrMsg{err}
		}
		return playerReadyMsg{p}
	}
}

// loadEpisode tells the player to start streaming the episode.
func loadEpisode(p *player.Player) tea.Cmd {
	return func() tea.Msg {
		if err := p.Load(episodeURL); err != nil {
			return playerErrMsg{err}
		}
		return tickMsg(time.Now())
	}
}

// scheduleTick queues the next state poll.
func scheduleTick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// pollState fetches current position/duration/pause from mpv.
func pollState(p *player.Player) tea.Cmd {
	return func() tea.Msg {
		s, err := p.GetState()
		if err != nil {
			return playerErrMsg{err}
		}
		return stateMsg(s)
	}
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case playerReadyMsg:
		m.pl = msg.p
		m.loading = false
		return m, loadEpisode(m.pl)

	case playerErrMsg:
		m.err = msg.err
		m.loading = false

	case tickMsg:
		return m, tea.Batch(pollState(m.pl), scheduleTick())

	case stateMsg:
		m.state = player.State(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case " ":
			if m.pl != nil {
				_ = m.pl.TogglePause()
			}

		case "left", "h":
			if m.pl != nil {
				_ = m.pl.Seek(-seekDelta)
			}

		case "right", "l":
			if m.pl != nil {
				_ = m.pl.Seek(seekDelta)
			}
		}
	}
	return m, nil
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m model) View() string {
	title := titleStyle.Render(episodeTitle)
	status := m.renderStatus()
	help := helpStyle.Render("space play/pause   ← / → seek ±30s   q quit")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, status, help)
}

func (m model) renderStatus() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("[error] %v", m.err))
	}
	if m.loading || (!m.state.Loaded && m.pl != nil) {
		return loadingStyle.Render("[loading] ...")
	}
	if m.pl == nil {
		return loadingStyle.Render("[starting] ...")
	}

	pos := player.FormatDuration(m.state.Position)
	dur := player.FormatDuration(m.state.Duration)
	elapsed := timeStyle.Render(fmt.Sprintf("%s / %s", pos, dur))

	if m.state.Paused {
		return fmt.Sprintf("%s  %s", pausedStyle.Render("[paused]"), elapsed)
	}
	return fmt.Sprintf("%s  %s", playingStyle.Render("[playing]"), elapsed)
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Clean shutdown: close mpv after the TUI exits.
	if fm, ok := finalModel.(model); ok && fm.pl != nil {
		_ = fm.pl.Close()
	}
}
