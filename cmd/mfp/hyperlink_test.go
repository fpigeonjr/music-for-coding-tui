package main

import (
	"os"
	"strings"
	"testing"
)

func TestTermSupportsHyperlinks_Ghostty(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "ghostty")
	os.Unsetenv("TERM")
	os.Unsetenv("VTE_VERSION")
	if !termSupportsHyperlinks() {
		t.Error("expected ghostty to support hyperlinks")
	}
}

func TestTermSupportsHyperlinks_iTerm(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	if !termSupportsHyperlinks() {
		t.Error("expected iTerm.app to support hyperlinks")
	}
}

func TestTermSupportsHyperlinks_Kitty(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "xterm-kitty")
	if !termSupportsHyperlinks() {
		t.Error("expected kitty to support hyperlinks")
	}
}

func TestTermSupportsHyperlinks_Unsupported(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("VTE_VERSION", "")
	if termSupportsHyperlinks() {
		t.Error("expected Apple_Terminal to not support hyperlinks")
	}
}

func TestHyperlink_SupportedTerminal(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "ghostty")
	got := hyperlink("about", "https://musicforprogramming.net/about")
	if !strings.Contains(got, "\033]8;;https://musicforprogramming.net/about\033\\") {
		t.Errorf("expected OSC 8 sequence, got %q", got)
	}
	if !strings.Contains(got, "about") {
		t.Errorf("expected label in output, got %q", got)
	}
}

func TestHyperlink_FallbackToTok(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("VTE_VERSION", "")
	got := hyperlink("about", "https://musicforprogramming.net/about")
	// Should not contain OSC escape sequences
	if strings.Contains(got, "\033]8;;") {
		t.Errorf("unsupported terminal should not emit OSC 8, got %q", got)
	}
	if !strings.Contains(got, "about") {
		t.Errorf("expected label in fallback output, got %q", got)
	}
}
