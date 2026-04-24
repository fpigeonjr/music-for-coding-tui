package main

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fpigeonjr/music-for-coding-tui/internal/feed"
	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
	"github.com/fpigeonjr/music-for-coding-tui/internal/store"
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
		// Restore saved volume immediately
		if m.pl != nil {
			_ = m.pl.SetVolume(m.volume)
		}
		if len(m.episodes) > 0 {
			m.loading = false
			m.pendingResume = m.positions[m.currentEpisode().Number]
			return m, loadEpisode(m.pl, m.currentEpisode())
		}

	case feedLoadedMsg:
		m.episodes = msg.episodes
		// Restore last-played episode if we have one
		if m.pendingEpisodeNum > 0 {
			for i, ep := range m.episodes {
				if ep.Number == m.pendingEpisodeNum {
					m.currentIdx = i
					m.selectedIdx = i
					m.adjustScroll()
					break
				}
			}
		}
		if m.playerReady {
			m.loading = false
			m.pendingResume = m.positions[m.currentEpisode().Number]
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
		m.tracksFetching = false

	case clearThemeMsgMsg:
		m.themeMsg = "" // non-fatal; tracklist stays empty

	case tickMsg:
		return m, tea.Batch(pollState(m.pl), scheduleTick())

	case stateMsg:
		wasLoaded := m.state.Loaded
		m.state = player.State(msg)
		// First loaded tick: seek to saved position if one exists
		if !wasLoaded && m.state.Loaded && m.pendingResume > 5 {
			resume := m.pendingResume
			m.pendingResume = 0
			if m.pl != nil {
				_ = m.pl.SeekAbsolute(resume)
			}
		}
		// Persist position every tick (best-effort, non-blocking)
		if m.state.Loaded && m.state.Position > 5 {
			ep := m.currentEpisode()
			go func() { _ = store.SavePosition(ep.Number, m.state.Position) }()
		}

	case tea.KeyMsg:
		// When help overlay is open, only handle close keys
		if m.showHelp {
			switch msg.String() {
			case "?", "esc", "q", "ctrl+c":
				m.showHelp = false
			}
			return m, nil
		}

		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.showHelp = true
			return m, nil

		case "t":
			// Cycle to next theme
			currentIdx := 0
			for i, th := range Themes {
				if th.Name == m.theme.Name {
					currentIdx = i
					break
				}
			}
			next := Themes[(currentIdx+1)%len(Themes)]
			m.theme = next
			setTheme(next)
			m.themeMsg = next.Name
			go func() { _ = store.SaveTheme(next.Name) }()
			return m, clearThemeMsgCmd()

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

		// r: random episode
		case "r":
			if len(m.episodes) > 1 {
				newIdx := rand.Intn(len(m.episodes))
				for newIdx == m.currentIdx {
					newIdx = rand.Intn(len(m.episodes))
				}
				return m, m.changeEpisode(newIdx)
			}

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

		// f: toggle favourite
		case "f":
			ep := m.currentEpisode()
			if ep.Number > 0 {
				m.favourites[ep.Number] = !m.favourites[ep.Number]
				go func() { _ = store.SaveFavourites(m.favourites) }()
			}

		// volume
		case "-", "_":
			if m.pl != nil {
				m.volume -= 10
				if m.volume < 0 {
					m.volume = 0
				}
				_ = m.pl.SetVolume(m.volume)
				go func() { _ = store.SaveVolume(m.volume) }()
			}
		case "=", "+":
			if m.pl != nil {
				m.volume += 10
				if m.volume > 150 {
					m.volume = 150
				}
				_ = m.pl.SetVolume(m.volume)
				go func() { _ = store.SaveVolume(m.volume) }()
			}
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
	m.pendingResume = m.positions[m.currentEpisode().Number]
	go func() { _ = store.SaveLastEpisode(m.currentEpisode().Number) }()
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

func clearThemeMsgCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearThemeMsgMsg{}
	})
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
