package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fpigeonjr/music-for-coding-tui/internal/feed"
	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
)

// ─── Init ────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return tea.Batch(spawnPlayer(), loadFeed())
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
		if len(m.episodes) > 0 {
			m.loading = false
			return m, loadEpisode(m.pl, m.currentEpisode())
		}

	case feedLoadedMsg:
		m.episodes = msg.episodes
		if m.playerReady {
			m.loading = false
			return m, loadEpisode(m.pl, m.currentEpisode())
		}

	case feedErrMsg:
		m.err = msg.err
		m.loading = false

	case playerErrMsg:
		m.err = msg.err
		m.loading = false

	case tracklistLoadedMsg:
		m.tracks = msg.tracks
		m.tracksFetching = false

	case tracklistErrMsg:
		m.tracksFetching = false // non-fatal; tracklist stays empty

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

		// prev/next: move the playing episode AND sync cursor
		case "p", "[":
			return m, m.changeEpisode(m.currentIdx - 1)

		case "n", "]":
			return m, m.changeEpisode(m.currentIdx + 1)

		// j/k: browse the list without changing what's playing
		case "j", "down":
			if m.selectedIdx < len(m.episodes)-1 {
				m.selectedIdx++
				m.adjustScroll()
			}

		case "k", "up":
			if m.selectedIdx > 0 {
				m.selectedIdx--
				m.adjustScroll()
			}

		// enter: play the highlighted episode
		case "enter":
			return m, m.changeEpisode(m.selectedIdx)
		}
	}
	return m, nil
}

// ─── changeEpisode ───────────────────────────────────────────────────────────

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
	m.selectedIdx = newIdx
	m.adjustScroll()
	m.state = player.State{}
	m.tracks = nil
	m.tracksFetching = true
	if m.pl == nil {
		return nil
	}
	return loadEpisode(m.pl, m.currentEpisode())
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

func loadEpisode(p *player.Player, ep feed.Episode) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			if err := p.Load(ep.URL); err != nil {
				return playerErrMsg{err}
			}
			return tickMsg(time.Now())
		},
		fetchTracklistCmd(ep.Slug),
	)
}

func fetchTracklistCmd(slug string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := feed.FetchTracklist(slug)
		if err != nil {
			return tracklistErrMsg{err}
		}
		return tracklistLoadedMsg{tracks}
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
