package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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
		m.commandInput.Width = m.width - 4

	case playerReadyMsg:
		m.pl = msg.p
		m.playerReady = true
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
		m.themeMsg = ""

	case clearCommandBarMsg:
		m.commandOutput = ""
		m.commandError = ""
		m.commandResult = false

	case commandResultMsg:
		m.commandOutput = msg.output
		m.commandError = ""
		m.commandResult = true
		return m, clearCommandBarCmd()

	case commandErrMsg:
		m.commandError = msg.err.Error()
		m.commandOutput = ""
		m.commandResult = true
		return m, clearCommandBarCmd()

	case tickMsg:
		return m, tea.Batch(pollState(m.pl), scheduleTick())

	case stateMsg:
		wasLoaded := m.state.Loaded
		m.state = player.State(msg)
		if !wasLoaded && m.state.Loaded && m.pendingResume > 5 {
			resume := m.pendingResume
			m.pendingResume = 0
			if m.pl != nil {
				_ = m.pl.SeekAbsolute(resume)
			}
		}
		if m.state.Loaded && m.state.Position > 5 {
			ep := m.currentEpisode()
			go func() { _ = store.SavePosition(ep.Number, m.state.Position) }()
		}

	case tea.KeyMsg:
		// Help overlay: only close keys active
		if m.showHelp {
			switch msg.String() {
			case "?", "esc", "q", "ctrl+c":
				m.showHelp = false
			}
			return m, nil
		}

		// ESC clears the result bar when not actively typing
		if m.commandResult && !m.commandMode && msg.String() == "esc" {
			m.commandResult = false
			m.commandOutput = ""
			m.commandError = ""
			return m, nil
		}

		// Command mode: route all keys to the textinput
		if m.commandMode {
			switch msg.String() {
			case "esc":
				m.commandMode = false
				m.commandResult = false
				m.commandInput.SetValue("")
				m.commandError = ""
				m.commandOutput = ""
				return m, nil
			case "enter":
				raw := m.commandInput.Value()
				m.commandInput.SetValue("")
				m.commandMode = false
				return m, m.executeCommand(raw)
			default:
				var cmd tea.Cmd
				m.commandInput, cmd = m.commandInput.Update(msg)
				return m, cmd
			}
		}

		// Normal mode
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.showHelp = true
			return m, nil

		case "/":
			m.commandMode = true
			m.commandResult = false
			m.commandOutput = ""
			m.commandError = ""
			m.commandInput.Focus()
			m.commandInput.SetValue("")
			return m, nil

		case "t":
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

		case "p", "[":
			return m, m.changeEpisode(m.currentIdx - 1)

		case "n", "]":
			return m, m.changeEpisode(m.currentIdx + 1)

		case "r":
			if len(m.episodes) > 1 {
				newIdx := rand.Intn(len(m.episodes))
				for newIdx == m.currentIdx {
					newIdx = rand.Intn(len(m.episodes))
				}
				return m, m.changeEpisode(newIdx)
			}

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

		case "enter":
			return m, m.changeEpisode(m.selectedIdx)

		case "f":
			ep := m.currentEpisode()
			if ep.Number > 0 {
				m.favourites[ep.Number] = !m.favourites[ep.Number]
				go func() { _ = store.SaveFavourites(m.favourites) }()
			}

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
		return m, nil
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

// ─── Async commands ──────────────────────────────────────────────────────────

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
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return clearThemeMsgMsg{}
	})
}

func clearCommandBarCmd() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(time.Time) tea.Msg {
		return clearCommandBarMsg{}
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

// ─── Command execution ───────────────────────────────────────────────────────

func (m *model) executeCommand(rawCmd string) tea.Cmd {
	cmd := strings.TrimSpace(rawCmd)
	if cmd == "" {
		return nil
	}
	parts := strings.Fields(cmd)
	command := parts[0]
	args := parts[1:]

	switch command {
	case "theme":
		return m.cmdTheme(args)
	case "jump":
		return m.cmdJump(args)
	case "vol", "volume":
		return m.cmdVolume(args)
	case "fav", "favourite", "favorite":
		return m.cmdFavourite()
	case "random":
		return m.cmdRandom()
	case "help":
		return m.cmdHelp()
	case "quit", "exit":
		return tea.Quit
	default:
		return m.cmdError("unknown command: " + command)
	}
}

func (m *model) cmdTheme(args []string) tea.Cmd {
	if len(args) == 0 {
		return m.cmdError("usage: theme <dracula|nord|gruvbox|onedark|everforest>")
	}
	raw := strings.Join(args, " ")
	input := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(raw, "-", " "), "_", " "))

	// "onedark" has no separator so prefix matching against "one dark" won't work
	if input == "onedark" {
		input = "one dark"
	}

	// Exact match first
	for _, t := range Themes {
		if strings.ToLower(t.Name) == input {
			return m.applyTheme(t)
		}
	}
	// Prefix match: "gruvbox" → "Gruvbox Dark", "everforest" → "Everforest Dark", etc.
	for _, t := range Themes {
		if strings.HasPrefix(strings.ToLower(t.Name), input) {
			return m.applyTheme(t)
		}
	}

	var names []string
	for _, t := range Themes {
		names = append(names, strings.ToLower(t.Name))
	}
	return m.cmdError("unknown theme — available: " + strings.Join(names, ", "))
}

func (m *model) applyTheme(t Theme) tea.Cmd {
	m.theme = t
	setTheme(t)
	m.themeMsg = t.Name
	go func() { _ = store.SaveTheme(t.Name) }()
	return tea.Batch(clearThemeMsgCmd(), func() tea.Msg {
		return commandResultMsg{output: "theme: " + t.Name}
	})
}

func (m *model) cmdJump(args []string) tea.Cmd {
	if len(args) == 0 {
		return m.cmdError("usage: jump <episode-number or title>")
	}

	// Numeric: jump by episode number
	if num, err := strconv.Atoi(args[0]); err == nil {
		for i, ep := range m.episodes {
			if ep.Number == num {
				title := ep.Title
				loadCmd := m.changeEpisode(i)
				return tea.Batch(loadCmd, func() tea.Msg {
					return commandResultMsg{output: fmt.Sprintf("→ ep %d: %s", num, title)}
				})
			}
		}
		return m.cmdError(fmt.Sprintf("no episode with number %d", num))
	}

	// Fuzzy: match by title
	query := strings.ToLower(strings.Join(args, " "))
	bestIdx, bestScore := -1, 0.0
	for i, ep := range m.episodes {
		if s := fuzzyMatch(strings.ToLower(ep.Title), query); s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	if bestIdx >= 0 && bestScore > 0.3 {
		ep := m.episodes[bestIdx]
		loadCmd := m.changeEpisode(bestIdx)
		return tea.Batch(loadCmd, func() tea.Msg {
			return commandResultMsg{output: fmt.Sprintf("→ ep %d: %s", ep.Number, ep.Title)}
		})
	}

	return m.cmdError("no episode matching: " + strings.Join(args, " "))
}

func (m *model) cmdVolume(args []string) tea.Cmd {
	if len(args) == 0 {
		return m.cmdError("usage: vol <0-150>")
	}
	vol, err := strconv.Atoi(args[0])
	if err != nil {
		return m.cmdError("volume must be a number between 0 and 150")
	}
	if vol < 0 || vol > 150 {
		return m.cmdError("volume must be between 0 and 150")
	}
	m.volume = vol
	if m.pl != nil {
		_ = m.pl.SetVolume(m.volume)
	}
	go func() { _ = store.SaveVolume(vol) }()
	return func() tea.Msg {
		return commandResultMsg{output: fmt.Sprintf("vol: %d%%", vol)}
	}
}

func (m *model) cmdFavourite() tea.Cmd {
	ep := m.currentEpisode()
	if ep.Number > 0 {
		m.favourites[ep.Number] = !m.favourites[ep.Number]
		go func() { _ = store.SaveFavourites(m.favourites) }()
	}
	isFav := m.favourites[ep.Number]
	status := "removed from"
	if isFav {
		status = "added to"
	}
	return func() tea.Msg {
		return commandResultMsg{output: fmt.Sprintf("ep %d %s favourites", ep.Number, status)}
	}
}

func (m *model) cmdRandom() tea.Cmd {
	if len(m.episodes) <= 1 {
		return func() tea.Msg {
			return commandResultMsg{output: "not enough episodes to randomise"}
		}
	}
	newIdx := rand.Intn(len(m.episodes))
	for newIdx == m.currentIdx {
		newIdx = rand.Intn(len(m.episodes))
	}
	ep := m.episodes[newIdx]
	loadCmd := m.changeEpisode(newIdx)
	return tea.Batch(loadCmd, func() tea.Msg {
		return commandResultMsg{output: fmt.Sprintf("→ ep %d: %s", ep.Number, ep.Title)}
	})
}

func (m *model) cmdHelp() tea.Cmd {
	m.showHelp = true
	return nil
}

func (m *model) cmdError(msg string) tea.Cmd {
	return func() tea.Msg {
		return commandErrMsg{err: errors.New(msg)}
	}
}

// fuzzyMatch returns ratio of matched characters of b found in-order within a.
func fuzzyMatch(a, b string) float64 {
	if len(b) == 0 {
		return 0
	}
	matches, bIdx := 0, 0
	for _, ch := range a {
		if bIdx < len(b) && byte(ch) == b[bIdx] {
			matches++
			bIdx++
		}
	}
	return float64(matches) / float64(len(b))
}
