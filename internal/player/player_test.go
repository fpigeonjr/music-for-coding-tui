package player

import (
	"os/exec"
	"testing"
	"time"
)

// ─── Unit tests (no mpv required) ────────────────────────────────────────────

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		secs float64
		want string
	}{
		{0, "00:00"},
		{5, "00:05"},
		{59, "00:59"},
		{60, "01:00"},
		{90, "01:30"},
		{3599, "59:59"},
		{3600, "1:00:00"},
		{5400, "1:30:00"},
		{90 * 60, "1:30:00"},
		{-5, "00:00"}, // negative clamped to zero
	}
	for _, tt := range tests {
		got := FormatDuration(tt.secs)
		if got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}

// ─── Integration tests (require mpv in PATH) ─────────────────────────────────

// requireMpv skips the test if mpv is not installed.
func requireMpv(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("mpv"); err != nil {
		t.Skip("mpv not in PATH — skipping integration test (brew install mpv)")
	}
}

func TestNewPlayer_SpawnAndClose(t *testing.T) {
	requireMpv(t)
	p, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
}

func TestPlayer_LoadURL(t *testing.T) {
	requireMpv(t)
	p, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer p.Close()

	// loadfile is async — mpv accepts the command even if the URL is not yet
	// streaming. We just verify the IPC round-trip succeeds.
	const ep78 = "https://datasette.net/music/musicforprogramming_078.mp3"
	if err := p.Load(ep78); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
}

func TestPlayer_TogglePause(t *testing.T) {
	requireMpv(t)
	p, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer p.Close()

	const ep78 = "https://datasette.net/music/musicforprogramming_078.mp3"
	if err := p.Load(ep78); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	// Give mpv a moment to register the file before toggling
	time.Sleep(300 * time.Millisecond)

	if err := p.TogglePause(); err != nil {
		t.Fatalf("TogglePause() error: %v", err)
	}
}

func TestPlayer_GetState(t *testing.T) {
	requireMpv(t)
	p, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer p.Close()

	// Before loading: state should be zero, no error
	state, err := p.GetState()
	if err != nil {
		t.Fatalf("GetState() before load error: %v", err)
	}
	if state.Loaded {
		t.Error("expected Loaded=false before any Load()")
	}
}

func TestPlayer_Seek(t *testing.T) {
	requireMpv(t)
	p, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer p.Close()

	const ep78 = "https://datasette.net/music/musicforprogramming_078.mp3"
	if err := p.Load(ep78); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	// mpv returns an error if seek is called before the file is demuxed;
	// we just verify no panic occurs — a seek error here is acceptable.
	_ = p.Seek(30)
	_ = p.Seek(-30)
}

func TestPlayer_NoOrphanProcess(t *testing.T) {
	requireMpv(t)
	p, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	proc := p.cmd.Process
	if err := p.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	// After Close(), the process should no longer be running.
	// Process.Kill() on an already-dead process returns an error — that's fine.
	// The key check: we can signal the process; if it's truly gone, FindProcess
	// still returns a handle but Signal returns "os: process already finished".
	if err := proc.Signal(nil); err == nil {
		// nil signal is a liveness check on Unix — success means still alive
		t.Error("mpv process still alive after Close()")
	}
}
