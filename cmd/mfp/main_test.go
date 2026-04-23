package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
)

// ─── renderStatus tests (pure, no mpv) ───────────────────────────────────────

func TestRenderStatus_Starting(t *testing.T) {
	m := initialModel() // loading=true, pl=nil
	got := m.renderStatus()
	if !strings.Contains(got, "starting") {
		t.Errorf("expected [starting] indicator, got %q", got)
	}
}

func TestRenderStatus_LoadingAfterConnect(t *testing.T) {
	m := initialModel()
	m.loading = false
	m.playerReady = true
	// state.Loaded is still false (mpv hasn't buffered yet)
	got := m.renderStatus()
	if !strings.Contains(got, "loading") {
		t.Errorf("expected [loading] indicator, got %q", got)
	}
}

func TestRenderStatus_Error(t *testing.T) {
	m := initialModel()
	m.loading = false
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
	m.state = player.State{
		Loaded:   true,
		Paused:   false,
		Position: 73,   // 01:13
		Duration: 5399, // 1:29:59
	}
	got := m.renderStatus()
	if !strings.Contains(got, "playing") {
		t.Errorf("expected [playing], got %q", got)
	}
	if !strings.Contains(got, "01:13") {
		t.Errorf("expected position 01:13, got %q", got)
	}
	if !strings.Contains(got, "1:29:59") {
		t.Errorf("expected duration 1:29:59, got %q", got)
	}
}

func TestRenderStatus_Paused(t *testing.T) {
	m := initialModel()
	m.loading = false
	m.playerReady = true
	m.state = player.State{
		Loaded:   true,
		Paused:   true,
		Position: 90,
		Duration: 5400,
	}
	got := m.renderStatus()
	if !strings.Contains(got, "paused") {
		t.Errorf("expected [paused], got %q", got)
	}
}

// ─── Update tests (message handling, no mpv) ─────────────────────────────────

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
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestUpdate_PlayerReady(t *testing.T) {
	m := initialModel()
	result, _ := m.Update(playerReadyMsg{p: nil}) // nil p: we test the flag, not the pointer
	fm := result.(model)
	if fm.loading {
		t.Error("expected loading=false after playerReadyMsg")
	}
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

func TestUpdate_StateMsg(t *testing.T) {
	m := initialModel()
	m.loading = false
	expected := player.State{Loaded: true, Paused: false, Position: 42, Duration: 5400}
	result, _ := m.Update(stateMsg(expected))
	fm := result.(model)
	if fm.state != expected {
		t.Errorf("expected state %+v, got %+v", expected, fm.state)
	}
}
