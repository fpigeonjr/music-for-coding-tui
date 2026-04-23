package main

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fpigeonjr/music-for-coding-tui/internal/feed"
	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func modelWithEpisodes() model {
	m := initialModel()
	m.loading = false
	m.playerReady = true
	m.width = 120
	m.height = 40
	m.episodes = []feed.Episode{
		{Number: 78, Title: "Datassette", URL: "https://example.com/78.mp3", Duration: "1:30:00", Size: 158925518},
		{Number: 77, Title: "Phonaut", URL: "https://example.com/77.mp3", Duration: "2:00:00", Size: 221651679},
		{Number: 76, Title: "Material Object", URL: "https://example.com/76.mp3", Duration: "2:14:02", Size: 244308026},
	}
	return m
}

// ─── renderStatus ────────────────────────────────────────────────────────────

func TestRenderStatus_Starting(t *testing.T) {
	m := initialModel()
	got := m.renderStatus()
	if !strings.Contains(got, "starting") {
		t.Errorf("expected [starting], got %q", got)
	}
}

func TestRenderStatus_LoadingAfterConnect(t *testing.T) {
	m := initialModel()
	m.loading = false
	m.playerReady = true
	got := m.renderStatus()
	if !strings.Contains(got, "loading") {
		t.Errorf("expected [loading], got %q", got)
	}
}

func TestRenderStatus_Error(t *testing.T) {
	m := initialModel()
	m.err = errMpvNotFound
	got := m.renderStatus()
	if !strings.Contains(got, "error") {
		t.Errorf("expected error indicator, got %q", got)
	}
}

func TestRenderStatus_Playing(t *testing.T) {
	m := initialModel()
	m.loading = false
	m.playerReady = true
	m.state = player.State{Loaded: true, Paused: false, Position: 73, Duration: 5399}
	got := m.renderStatus()
	if !strings.Contains(got, "playing") {
		t.Errorf("expected [playing], got %q", got)
	}
	if !strings.Contains(got, "01:13") {
		t.Errorf("expected 01:13, got %q", got)
	}
}

func TestRenderStatus_Paused(t *testing.T) {
	m := initialModel()
	m.loading = false
	m.playerReady = true
	m.state = player.State{Loaded: true, Paused: true, Position: 90, Duration: 5400}
	got := m.renderStatus()
	if !strings.Contains(got, "paused") {
		t.Errorf("expected [paused], got %q", got)
	}
}

// ─── pane layout ─────────────────────────────────────────────────────────────

func TestPaneWidths_120col(t *testing.T) {
	m := initialModel()
	m.width = 120
	left, center, right := m.paneWidths()
	if left+center+right != 120 {
		t.Errorf("pane widths don't sum to 120: %d+%d+%d=%d", left, center, right, left+center+right)
	}
	if left != 30 || right != 30 {
		t.Errorf("expected left=30 right=30 at width 120, got left=%d right=%d", left, right)
	}
}

func TestPaneWidths_80col(t *testing.T) {
	m := initialModel()
	m.width = 80
	left, center, right := m.paneWidths()
	if left+center+right != 80 {
		t.Errorf("pane widths don't sum to 80: %d+%d+%d=%d", left, center, right, left+center+right)
	}
}

// ─── renderRight ─────────────────────────────────────────────────────────────

func TestRenderRight_ShowsEpisodes(t *testing.T) {
	m := modelWithEpisodes()
	got := m.renderRight(30)
	if !strings.Contains(got, "78") {
		t.Errorf("expected ep 78 in right pane, got %q", got)
	}
	if !strings.Contains(got, "Datassette") {
		t.Errorf("expected 'Datassette' in right pane, got %q", got)
	}
}

func TestRenderRight_SelectionMarker(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 0
	m.selectedIdx = 0
	got := m.renderRight(30)
	if !strings.Contains(got, "▶") {
		t.Errorf("expected play marker ▶ for current+selected, got %q", got)
	}
}

func TestRenderRight_BrowseWithoutPlaying(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 0  // playing ep 78
	m.selectedIdx = 1 // cursor on ep 77
	got := m.renderRight(30)
	if !strings.Contains(got, "▶") {
		t.Errorf("expected ▶ marker for playing episode, got %q", got)
	}
}

// ─── renderLeft ──────────────────────────────────────────────────────────────

func TestRenderLeft_ContainsPreamble(t *testing.T) {
	m := modelWithEpisodes()
	got := m.renderLeft(28)
	if !strings.Contains(got, "musicFor") {
		t.Errorf("expected preamble in left pane, got %q", got)
	}
}

func TestRenderLeft_ContainsTransport(t *testing.T) {
	m := modelWithEpisodes()
	got := m.renderLeft(28)
	if !strings.Contains(got, "[stop]") {
		t.Errorf("expected transport controls in left pane, got %q", got)
	}
}

func TestRenderLeft_ContainsStats(t *testing.T) {
	m := modelWithEpisodes()
	got := m.renderLeft(28)
	if !strings.Contains(got, "episodes") {
		t.Errorf("expected stats in left pane, got %q", got)
	}
}

// ─── renderCenter ────────────────────────────────────────────────────────────

func TestRenderCenter_ShowsEpisodeTitle(t *testing.T) {
	m := modelWithEpisodes()
	got := m.renderCenter(60)
	if !strings.Contains(got, "78") {
		t.Errorf("expected episode number in center, got %q", got)
	}
	if !strings.Contains(got, "Datassette") {
		t.Errorf("expected episode title in center, got %q", got)
	}
}

func TestRenderCenter_ShowsTracklist(t *testing.T) {
	m := modelWithEpisodes()
	m.tracks = []feed.Track{
		{Artist: "David Borden", Title: "Enfield In Winter"},
		{Artist: "Datassette", Title: "rain_wind_canvas"},
	}
	got := m.renderCenter(60)
	if !strings.Contains(got, "David Borden") {
		t.Errorf("expected tracklist in center, got %q", got)
	}
}

func TestRenderCenter_FetchingPlaceholder(t *testing.T) {
	m := modelWithEpisodes()
	m.tracksFetching = true
	got := m.renderCenter(60)
	if !strings.Contains(got, "fetching") {
		t.Errorf("expected fetching placeholder, got %q", got)
	}
}

// ─── truncate ────────────────────────────────────────────────────────────────

func TestTruncate(t *testing.T) {
	tests := []struct {
		s      string
		max    int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"toolongstring", 10, "toolongst…"},
		{"a", 1, "a"},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
		}
	}
}

// ─── adjustScroll ────────────────────────────────────────────────────────────

func TestAdjustScroll_ScrollsDown(t *testing.T) {
	m := modelWithEpisodes()
	m.height = 10 // rightPaneHeight = 6
	m.listOffset = 0
	m.selectedIdx = 10 // way past visible
	m.adjustScroll()
	if m.listOffset == 0 {
		t.Error("expected listOffset to advance when selection is below visible area")
	}
}

func TestAdjustScroll_ScrollsUp(t *testing.T) {
	m := modelWithEpisodes()
	m.listOffset = 5
	m.selectedIdx = 2
	m.adjustScroll()
	if m.listOffset != 2 {
		t.Errorf("listOffset = %d, want 2", m.listOffset)
	}
}

// ─── Update messages ─────────────────────────────────────────────────────────

func TestUpdate_WindowSize(t *testing.T) {
	m := initialModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	fm := result.(model)
	if fm.width != 120 || fm.height != 40 {
		t.Errorf("expected 120×40, got %d×%d", fm.width, fm.height)
	}
}

func TestUpdate_QuitKey(t *testing.T) {
	m := initialModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("expected a quit command, got nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", cmd())
	}
}

func TestUpdate_PlayerReady(t *testing.T) {
	m := initialModel()
	result, _ := m.Update(playerReadyMsg{p: nil})
	fm := result.(model)
	if !fm.playerReady {
		t.Error("expected playerReady=true after playerReadyMsg")
	}
}

func TestUpdate_PlayerError(t *testing.T) {
	m := initialModel()
	result, _ := m.Update(playerErrMsg{err: errMpvNotFound})
	fm := result.(model)
	if fm.err == nil {
		t.Error("expected err to be set")
	}
	if fm.loading {
		t.Error("expected loading=false")
	}
}

func TestUpdate_FeedLoaded(t *testing.T) {
	m := initialModel()
	eps := []feed.Episode{{Number: 78, Title: "Datassette"}}
	result, _ := m.Update(feedLoadedMsg{episodes: eps})
	fm := result.(model)
	if len(fm.episodes) != 1 {
		t.Errorf("expected 1 episode, got %d", len(fm.episodes))
	}
}

func TestUpdate_FeedError(t *testing.T) {
	m := initialModel()
	result, _ := m.Update(feedErrMsg{err: fmt.Errorf("network error")})
	fm := result.(model)
	if fm.err == nil {
		t.Error("expected err after feedErrMsg")
	}
}

func TestUpdate_TracklistLoaded(t *testing.T) {
	m := initialModel()
	tracks := []feed.Track{{Artist: "A", Title: "B"}}
	result, _ := m.Update(tracklistLoadedMsg{tracks: tracks})
	fm := result.(model)
	if len(fm.tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(fm.tracks))
	}
	if fm.tracksFetching {
		t.Error("expected tracksFetching=false after load")
	}
}

func TestUpdate_StateMsg(t *testing.T) {
	m := initialModel()
	expected := player.State{Loaded: true, Position: 42, Duration: 5400}
	result, _ := m.Update(stateMsg(expected))
	fm := result.(model)
	if fm.state != expected {
		t.Errorf("state = %+v, want %+v", fm.state, expected)
	}
}

func TestUpdate_JKScroll(t *testing.T) {
	m := modelWithEpisodes()
	m.selectedIdx = 1

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	fm := result.(model)
	if fm.selectedIdx != 2 {
		t.Errorf("j: selectedIdx = %d, want 2", fm.selectedIdx)
	}

	result, _ = fm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	fm = result.(model)
	if fm.selectedIdx != 1 {
		t.Errorf("k: selectedIdx = %d, want 1", fm.selectedIdx)
	}
}

// ─── changeEpisode ────────────────────────────────────────────────────────────

func TestChangeEpisode_UpdatesIdx(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 0
	m.changeEpisode(1)
	if m.currentIdx != 1 {
		t.Errorf("currentIdx = %d, want 1", m.currentIdx)
	}
	if m.selectedIdx != 1 {
		t.Errorf("selectedIdx = %d, want 1 (should sync with current)", m.selectedIdx)
	}
}

func TestChangeEpisode_ClampMin(t *testing.T) {
	m := modelWithEpisodes()
	m.changeEpisode(-1)
	if m.currentIdx != 0 {
		t.Errorf("currentIdx = %d, want 0 (clamped)", m.currentIdx)
	}
}

func TestChangeEpisode_ClampMax(t *testing.T) {
	m := modelWithEpisodes()
	m.changeEpisode(99)
	if m.currentIdx != 2 {
		t.Errorf("currentIdx = %d, want 2 (clamped)", m.currentIdx)
	}
}

func TestChangeEpisode_SameIdx_ReturnsNil(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 1
	cmd := m.changeEpisode(1)
	if cmd != nil {
		t.Error("expected nil cmd when idx unchanged")
	}
}

func TestChangeEpisode_ClearsTracklist(t *testing.T) {
	m := modelWithEpisodes()
	m.tracks = []feed.Track{{Artist: "A", Title: "B"}}
	m.changeEpisode(1)
	if len(m.tracks) != 0 {
		t.Error("expected tracks to be cleared on episode change")
	}
	if !m.tracksFetching {
		t.Error("expected tracksFetching=true on episode change")
	}
}
