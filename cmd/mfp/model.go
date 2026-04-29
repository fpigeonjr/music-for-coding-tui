package main

import (
	"errors"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/fpigeonjr/music-for-coding-tui/internal/feed"
	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
	"github.com/fpigeonjr/music-for-coding-tui/internal/store"
)

// ─── Constants ───────────────────────────────────────────────────────────────

const (
	seekDelta    = 30.0
	tickInterval = 250 * time.Millisecond
	minWidth     = 80
	minHeight    = 20
)

// errMpvNotFound is a sentinel used in tests.

var errMpvNotFound = errors.New("mpv not found — install with: brew install mpv")

// ─── Messages ────────────────────────────────────────────────────────────────

type tickMsg            time.Time
type playerReadyMsg     struct{ p *player.Player }
type playerErrMsg       struct{ err error }
type stateMsg           player.State
type feedLoadedMsg      struct{ episodes []feed.Episode }
type feedErrMsg         struct{ err error }
type tracklistLoadedMsg struct{ tracks []feed.Track }
type tracklistErrMsg    struct{ err error }
type clearThemeMsgMsg   struct{}
type clearCommandBarMsg struct{}
type showHelpMsg        struct{} // unused, toggled directly in model
type commandResultMsg   struct{ output string }
type commandErrMsg      struct{ err error }

// ─── Model ───────────────────────────────────────────────────────────────────
type model struct {
	width int
	height int

	// player
	pl *player.Player
	state player.State
	playerReady bool

	// feed
	episodes []feed.Episode
	currentIdx int // which episode is playing
	selectedIdx int // cursor in right pane (can differ from currentIdx)
	listOffset int // scroll offset for right pane

	// tracklist for the current episode (fetched async)
	tracks []feed.Track
	tracksFetching bool

	// niceties
	favourites map[int]bool
	positions store.Positions
	volume int // 0-150
	pendingResume float64 // seek to this position on next loaded tick (0 = no resume)
	theme Theme // active colour theme
	themeMsg string // flashes theme name briefly after switching
	pendingEpisodeNum int // episode number to restore when feed loads (0 = newest)
	showHelp bool // ? overlay visible
	loading bool
	err error

	// command bar
	commandMode   bool            // true when actively typing a command
	commandResult bool            // true when showing result/error (bar visible, read-only)
	commandInput  textinput.Model // textinput for command bar
	commandError  string          // error from last command
	commandOutput string          // success message from last command
}

func initialModel() model {
	favs, _ := store.LoadFavourites()
	pos, _ := store.LoadPositions()
	vol, _ := store.LoadVolume()
	themeName, _ := store.LoadTheme()
	lastEp, _ := store.LoadLastEpisode()

	active := ThemeDracula
	for _, t := range Themes {
		if t.Name == themeName {
			active = t
			break
		}
	}
	setTheme(active)

	// Setup command input
	ti := textinput.New()
	ti.Prompt = ":"
	ti.Placeholder = ""
	ti.CharLimit = 156
	ti.Width = 40

	return model{
		loading:           true,
		favourites:        favs,
		positions:         pos,
		volume:            vol,
		theme:             active,
		pendingEpisodeNum: lastEp,
		commandInput:      ti,
	}
}

// currentEpisode returns the episode currently playing.
func (m model) currentEpisode() feed.Episode {
	if len(m.episodes) == 0 || m.currentIdx < 0 || m.currentIdx >= len(m.episodes) {
		return feed.Episode{}
	}
	return m.episodes[m.currentIdx]
}

// adjustScroll keeps selectedIdx visible in the right pane.
func (m *model) adjustScroll() {
	visible := m.rightPaneHeight()
	if m.selectedIdx < m.listOffset {
		m.listOffset = m.selectedIdx
	}
	if m.selectedIdx >= m.listOffset+visible {
		m.listOffset = m.selectedIdx - visible + 1
	}
	if m.listOffset < 0 {
		m.listOffset = 0
	}
}

// rightPaneHeight returns the number of lines available for the episode list.
func (m model) rightPaneHeight() int {
	h := m.height - 4
	if h < 1 {
		h = 1
	}
	return h
}

// paneWidths returns (left, center, right) widths based on terminal width.
func (m model) paneWidths() (int, int, int) {
	left := m.width / 4
	right := m.width / 4
	center := m.width - left - right
	return left, center, right
}
