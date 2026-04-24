package main

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// ─── Theme struct ────────────────────────────────────────────────────────────

// Theme defines the full colour palette for the TUI.
// All render functions derive their styles from the active theme via setTheme().
type Theme struct {
	Name    string
	Keyword string // keywords: function, return  (italic applied by style)
	Param   string // parameters, time display    (italic applied by style)
	Str     string // strings, template literals
	Bracket string // [tokens], episode numbers + titles
	Fg      string // main foreground text
	Dim     string // separators, very inactive text
	Comment string // // stats, secondary text, loading states
	Playing string // [playing] indicator
	Paused  string // [paused] indicator
	Error   string // errors
}

// ─── Built-in themes ─────────────────────────────────────────────────────────

var ThemeDracula = Theme{
	Name:    "Dracula",
	Keyword: "#ff79c6", // pink
	Param:   "#ffb86c", // orange
	Str:     "#f1fa8c", // yellow
	Bracket: "#8be9fd", // cyan
	Fg:      "#f8f8f2", // near-white
	Dim:     "#44475a", // dark gray
	Comment: "#6272a4", // muted blue-gray
	Playing: "#50fa7b", // green
	Paused:  "#ffb86c", // orange
	Error:   "#ff5555", // red
}

var ThemeNord = Theme{
	Name:    "Nord",
	Keyword: "#81a1c1", // nord9  blue
	Param:   "#ebcb8b", // nord13 yellow
	Str:     "#a3be8c", // nord14 green
	Bracket: "#88c0d0", // nord8  light blue
	Fg:      "#d8dee9", // nord4  light gray
	Dim:     "#3b4252", // nord1  dark
	Comment: "#616e88", // between nord2/nord3
	Playing: "#a3be8c", // nord14 green
	Paused:  "#ebcb8b", // nord13 yellow
	Error:   "#bf616a", // nord11 red
}

var ThemeGruvboxDark = Theme{
	Name:    "Gruvbox Dark",
	Keyword: "#fb4934", // bright red
	Param:   "#fabd2f", // bright yellow
	Str:     "#b8bb26", // bright green
	Bracket: "#83a598", // bright aqua
	Fg:      "#ebdbb2", // light0
	Dim:     "#504945", // dark2
	Comment: "#928374", // gray
	Playing: "#b8bb26", // bright green
	Paused:  "#fabd2f", // bright yellow
	Error:   "#fb4934", // bright red
}

var ThemeOneDark = Theme{
	Name:    "One Dark",
	Keyword: "#c678dd", // purple
	Param:   "#e5c07b", // yellow
	Str:     "#98c379", // green
	Bracket: "#61afef", // blue
	Fg:      "#abb2bf", // light gray
	Dim:     "#3e4452", // dark gray
	Comment: "#5c6370", // comment gray
	Playing: "#98c379", // green
	Paused:  "#e5c07b", // yellow
	Error:   "#e06c75", // red
}

var ThemeEverforestDark = Theme{
	Name:    "Everforest Dark",
	Keyword: "#e67e80", // red
	Param:   "#dbbc7f", // yellow
	Str:     "#a7c080", // green
	Bracket: "#83c092", // aqua
	Fg:      "#d3c6aa", // fg
	Dim:     "#3d484d", // bg2
	Comment: "#859289", // grey1
	Playing: "#a7c080", // green
	Paused:  "#e69875", // orange
	Error:   "#e67e80", // red
}

// Themes is the ordered cycle: t key steps through this slice.
var Themes = []Theme{
	ThemeDracula,
	ThemeNord,
	ThemeGruvboxDark,
	ThemeOneDark,
	ThemeEverforestDark,
}

// ─── Active style vars (set by setTheme) ─────────────────────────────────────

var (
	// Playback state
	playingStyle lipgloss.Style
	pausedStyle  lipgloss.Style
	loadingStyle lipgloss.Style
	errorStyle   lipgloss.Style

	// Syntax — preamble
	keywordStyle lipgloss.Style
	fnNameStyle  lipgloss.Style
	paramStyle   lipgloss.Style
	stringStyle  lipgloss.Style
	punctStyle   lipgloss.Style
	fgStyle      lipgloss.Style

	// UI tokens
	bracketStyle lipgloss.Style
	timeStyle    lipgloss.Style
	linkStyle    lipgloss.Style
	sepStyle     lipgloss.Style
	commentStyle lipgloss.Style
	dimStyle     lipgloss.Style

	// Episode list
	epNumStyle      lipgloss.Style
	epTitleStyle    lipgloss.Style
	epCurrentStyle  lipgloss.Style
	epSelectedStyle lipgloss.Style
	epDimStyle      lipgloss.Style

	// Center pane
	episodeTitleStyle lipgloss.Style
	trackArtistStyle  lipgloss.Style
	trackSepStyle     lipgloss.Style
	trackTitleStyle   lipgloss.Style
)

// setTheme updates all style vars from the given theme.
// Called once at startup (from initialModel) and on every theme switch.
func setTheme(t Theme) {
	c := func(hex string) lipgloss.Color { return lipgloss.Color(hex) }

	// Playback
	playingStyle = lipgloss.NewStyle().Foreground(c(t.Playing))
	pausedStyle  = lipgloss.NewStyle().Foreground(c(t.Paused))
	loadingStyle = lipgloss.NewStyle().Foreground(c(t.Comment))
	errorStyle   = lipgloss.NewStyle().Foreground(c(t.Error))

	// Syntax
	keywordStyle = lipgloss.NewStyle().Foreground(c(t.Keyword)).Italic(true)
	fnNameStyle  = lipgloss.NewStyle().Foreground(c(t.Fg))
	paramStyle   = lipgloss.NewStyle().Foreground(c(t.Param)).Italic(true)
	stringStyle  = lipgloss.NewStyle().Foreground(c(t.Str))
	punctStyle   = lipgloss.NewStyle().Foreground(c(t.Fg))
	fgStyle      = lipgloss.NewStyle().Foreground(c(t.Fg))

	// UI tokens
	bracketStyle = lipgloss.NewStyle().Foreground(c(t.Bracket))
	timeStyle    = lipgloss.NewStyle().Foreground(c(t.Fg))
	linkStyle    = lipgloss.NewStyle().Foreground(c(t.Bracket))
	sepStyle     = lipgloss.NewStyle().Foreground(c(t.Dim))
	commentStyle = lipgloss.NewStyle().Foreground(c(t.Comment))
	dimStyle     = lipgloss.NewStyle().Foreground(c(t.Dim))

	// Episode list
	epNumStyle      = lipgloss.NewStyle().Foreground(c(t.Bracket))
	epTitleStyle    = lipgloss.NewStyle().Foreground(c(t.Bracket)).Italic(true)
	epCurrentStyle  = lipgloss.NewStyle().Foreground(c(t.Bracket)).Bold(true)
	epSelectedStyle = lipgloss.NewStyle().Foreground(c(t.Fg)).Bold(true)
	epDimStyle      = lipgloss.NewStyle().Foreground(c(t.Comment)).Italic(true)

	// Center pane
	episodeTitleStyle = lipgloss.NewStyle().Foreground(c(t.Fg))
	trackArtistStyle  = lipgloss.NewStyle().Foreground(c(t.Fg))
	trackSepStyle     = lipgloss.NewStyle().Foreground(c(t.Comment))
	trackTitleStyle   = lipgloss.NewStyle().Foreground(c(t.Fg))
}

// ─── OSC 8 hyperlinks ────────────────────────────────────────────────────────

// termSupportsHyperlinks returns true when the terminal is known to support
// OSC 8 hyperlinks. Falls back gracefully in unsupported terminals.
func termSupportsHyperlinks() bool {
	switch os.Getenv("TERM_PROGRAM") {
	case "ghostty", "iTerm.app", "WezTerm", "wezterm", "Hyper", "tabby":
		return true
	}
	// Kitty uses TERM=xterm-kitty
	if os.Getenv("TERM") == "xterm-kitty" {
		return true
	}
	// VTE-based terminals (GNOME Terminal, Tilix, etc.)
	if os.Getenv("VTE_VERSION") != "" {
		return true
	}
	return false
}

// hyperlink renders label as a clickable OSC 8 link in supporting terminals.
// Falls back to a plain bracketed tok() in unsupported ones.
func hyperlink(label, url string) string {
	if !termSupportsHyperlinks() {
		return tok(label)
	}
	// OSC 8 ; params ; URI ST label OSC 8 ; ; ST
	return "\033]8;;" + url + "\033\\" +
		bracketStyle.Render("["+label+"]") +
		"\033]8;;\033\\"
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// tok renders a [bracketed] control token in the current bracket colour.
func tok(s string) string {
	return bracketStyle.Render("[" + s + "]")
}
