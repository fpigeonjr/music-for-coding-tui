package main

import (
	"errors"
	"time"

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

// (styles live in styles.go)

type tickMsg            time.Time
type playerReadyMsg     struct{ p *player.Player }
type playerErrMsg       struct{ err error }
type stateMsg           player.State
type feedLoadedMsg      struct{ episodes []feed.Episode }
type feedErrMsg         struct{ err error }
type tracklistLoadedMsg struct{ tracks []feed.Track }
type tracklistErrMsg    struct{ err error }

// ─── Model ───────────────────────────────────────────────────────────────────

type model struct {
	width  int
	height int

	// player
	pl          *player.Player
	state       player.State
	playerReady bool

	// feed
	episodes    []feed.Episode
	currentIdx  int // which episode is playing
	selectedIdx int // cursor in right pane (can differ from currentIdx)
	listOffset  int // scroll offset for right pane

	// tracklist for the current episode (fetched async)
	tracks         []feed.Track
	tracksFetching bool

	// niceties
	favourites    map[int]bool
	positions     store.Positions
	volume        int     // 0-150
	pendingResume float64 // seek to this position on next loaded tick (0 = no resume)

	loading bool
	err     error
}

func initialModel() model {
	favs, _ := store.LoadFavourites()
	pos, _ := store.LoadPositions()
	vol, _ := store.LoadVolume()
	return model{
		loading:    true,
		favourites: favs,
		positions:  pos,
		volume:     vol,
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
