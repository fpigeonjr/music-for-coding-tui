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
	m.episodes = []feed.Episode{
		{Number: 78, Title: "Datassette", URL: "https://example.com/78.mp3", Duration: "1:30:00"},
		{Number: 77, Title: "Phonaut", URL: "https://example.com/77.mp3", Duration: "2:00:00"},
		{Number: 76, Title: "Material Object", URL: "https://example.com/76.mp3", Duration: "2:14:02"},
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
	if !strings.Contains(got, "1:29:59") {
		t.Errorf("expected 1:29:59, got %q", got)
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

// ─── renderNav ───────────────────────────────────────────────────────────────

func TestRenderNav_MiddleEpisode(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 1 // episode 77, between 78 and 76
	got := m.renderNav()
	if !strings.Contains(got, "78") {
		t.Errorf("expected ep 78 in nav, got %q", got)
	}
	if !strings.Contains(got, "76") {
		t.Errorf("expected ep 76 in nav, got %q", got)
	}
}

func TestRenderNav_FirstEpisode(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 0 // newest — no next
	got := m.renderNav()
	if strings.Contains(got, "→") {
		t.Errorf("newest episode should not show next arrow, got %q", got)
	}
}

func TestRenderNav_LastEpisode(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 2 // oldest — no prev
	got := m.renderNav()
	if strings.Contains(got, "←") {
		t.Errorf("oldest episode should not show prev arrow, got %q", got)
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
		t.Error("expected err to be set after playerErrMsg")
	}
	if fm.loading {
		t.Error("expected loading=false after playerErrMsg")
	}
}

func TestUpdate_FeedLoaded(t *testing.T) {
	m := initialModel()
	eps := []feed.Episode{
		{Number: 78, Title: "Datassette", URL: "https://example.com/78.mp3"},
	}
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

func TestUpdate_StateMsg(t *testing.T) {
	m := initialModel()
	expected := player.State{Loaded: true, Position: 42, Duration: 5400}
	result, _ := m.Update(stateMsg(expected))
	fm := result.(model)
	if fm.state != expected {
		t.Errorf("state = %+v, want %+v", fm.state, expected)
	}
}

// ─── changeEpisode ────────────────────────────────────────────────────────────

func TestChangeEpisode_Forward(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 0
	_ = m.changeEpisode(1) // pl is nil in tests; index still updates
	if m.currentIdx != 1 {
		t.Errorf("currentIdx = %d, want 1", m.currentIdx)
	}
}

func TestChangeEpisode_ClampMin(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 0
	m.changeEpisode(-1)
	if m.currentIdx != 0 {
		t.Errorf("currentIdx = %d, want 0 (clamped)", m.currentIdx)
	}
}

func TestChangeEpisode_ClampMax(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 2
	m.changeEpisode(99)
	if m.currentIdx != 2 {
		t.Errorf("currentIdx = %d, want 2 (clamped)", m.currentIdx)
	}
}

func TestChangeEpisode_SameIdx(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 1
	cmd := m.changeEpisode(1)
	if cmd != nil {
		t.Error("expected nil command when idx unchanged")
	}
}
