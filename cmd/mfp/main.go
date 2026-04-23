package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fpigeonjr/music-for-coding-tui/internal/feed"
	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
)

// ─── Constants ───────────────────────────────────────────────────────────────

const (
	seekDelta    = 30.0
	tickInterval = 250 * time.Millisecond
)

// errMpvNotFound is returned (and used in tests) when mpv cannot be started.
var errMpvNotFound = errors.New("mpv not found — install with: brew install mpv")

// ─── Messages ────────────────────────────────────────────────────────────────

type tickMsg time.Time
type playerReadyMsg struct{ p *player.Player }
type playerErrMsg struct{ err error }
type stateMsg player.State
type feedLoadedMsg struct{ episodes []feed.Episode }
type feedErrMsg struct{ err error }

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

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))
)

// ─── Model ───────────────────────────────────────────────────────────────────

type model struct {
	width  int
	height int

	// player
	pl          *player.Player
	state       player.State
	playerReady bool

	// feed
	episodes   []feed.Episode
	currentIdx int // index into episodes slice (0 = newest)

	loading bool
	err     error
}

func initialModel() model {
	return model{loading: true}
}

// currentEpisode returns the episode being played, or a zero Episode.
func (m model) currentEpisode() feed.Episode {
	if len(m.episodes) == 0 || m.currentIdx < 0 || m.currentIdx >= len(m.episodes) {
		return feed.Episode{}
	}
	return m.episodes[m.currentIdx]
}

// ─── Init ────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return tea.Batch(spawnPlayer(), loadFeed())
}

// ─── Commands ────────────────────────────────────────────────────────────────

func spawnPlayer() tea.Cmd {
	return func() tea.Msg {
		p, err := player.New()
		if err != nil {
			return playerErrMsg{err}
		}
		return playerReadyMsg{p}
	}
}

func loadFeed() tea.Cmd {
	return func() tea.Msg {
		eps, err := feed.Fetch()
		if err != nil {
			return feedErrMsg{err}
		}
		return feedLoadedMsg{eps}
	}
}

func loadEpisode(p *player.Player, url string) tea.Cmd {
	return func() tea.Msg {
		if err := p.Load(url); err != nil {
			return playerErrMsg{err}
		}
		return tickMsg(time.Now())
	}
}

func scheduleTick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

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
		m.playerReady = true
		// If feed is already loaded, start playing episode 0 immediately.
		if len(m.episodes) > 0 {
			m.loading = false
			return m, loadEpisode(m.pl, m.currentEpisode().URL)
		}
		// Otherwise wait for feedLoadedMsg.

	case feedLoadedMsg:
		m.episodes = msg.episodes
		// If player is already ready, kick off playback now.
		if m.playerReady {
			m.loading = false
			return m, loadEpisode(m.pl, m.currentEpisode().URL)
		}
		// Otherwise wait for playerReadyMsg.

	case feedErrMsg:
		m.err = msg.err
		m.loading = false

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

		case "n", "]":
			return m, m.changeEpisode(m.currentIdx + 1)

		case "p", "[":
			return m, m.changeEpisode(m.currentIdx - 1)
		}
	}
	return m, nil
}

// changeEpisode loads the episode at newIdx, clamped to valid range.
func (m *model) changeEpisode(newIdx int) tea.Cmd {
	if len(m.episodes) == 0 {
		return nil
	}
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(m.episodes) {
		newIdx = len(m.episodes) - 1
	}
	if newIdx == m.currentIdx {
		return nil
	}
	m.currentIdx = newIdx
	m.state = player.State{} // reset display while buffering
	if m.pl == nil {
		return nil // index updated; playback will start when player is ready
	}
	return loadEpisode(m.pl, m.currentEpisode().URL)
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m model) View() string {
	ep := m.currentEpisode()
	epTitle := ep.Title
	if ep.Number > 0 {
		epTitle = fmt.Sprintf("Episode %d: %s", ep.Number, ep.Title)
	}

	title := titleStyle.Render(epTitle)
	status := m.renderStatus()
	nav := m.renderNav()
	help := helpStyle.Render("space play/pause   ← / → seek ±30s   p / n prev/next   q quit")

	return fmt.Sprintf("\n  %s\n\n  %s\n  %s\n\n  %s\n", title, status, nav, help)
}

func (m model) renderStatus() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("[error] %v", m.err))
	}
	if m.loading || !m.playerReady {
		return loadingStyle.Render("[starting] ...")
	}
	if !m.state.Loaded {
		return loadingStyle.Render("[loading] ...")
	}

	pos := player.FormatDuration(m.state.Position)
	dur := player.FormatDuration(m.state.Duration)
	elapsed := timeStyle.Render(fmt.Sprintf("%s / %s", pos, dur))

	if m.state.Paused {
		return fmt.Sprintf("%s  %s", pausedStyle.Render("[paused]"), elapsed)
	}
	return fmt.Sprintf("%s  %s", playingStyle.Render("[playing]"), elapsed)
}

func (m model) renderNav() string {
	if len(m.episodes) == 0 {
		return ""
	}
	total := len(m.episodes)
	idx := m.currentIdx

	var prev, next string
	if idx < total-1 {
		prev = fmt.Sprintf("← %d: %s", m.episodes[idx+1].Number, m.episodes[idx+1].Title)
	}
	if idx > 0 {
		next = fmt.Sprintf("%d: %s →", m.episodes[idx-1].Number, m.episodes[idx-1].Title)
	}

	if prev == "" && next == "" {
		return ""
	}
	if prev == "" {
		return dimStyle.Render(fmt.Sprintf("                   %s", next))
	}
	if next == "" {
		return dimStyle.Render(prev)
	}
	return dimStyle.Render(fmt.Sprintf("%-30s  %s", prev, next))
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if fm, ok := finalModel.(model); ok && fm.pl != nil {
		_ = fm.pl.Close()
	}
}
