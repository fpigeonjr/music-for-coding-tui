package main

import "github.com/charmbracelet/lipgloss"

// ─── MFP Colour Palette (Dracula-inspired) ───────────────────────────────────
//
// Derived from the musicforprogramming.net source screenshot:
//   background  #1a1a1a  near-black
//   foreground  #f8f8f2  near-white
//   comment     #6272a4  muted blue-gray  (// stats, dim text)
//   cyan        #8be9fd  bright cyan      ([tokens], episode numbers)
//   green       #50fa7b  bright green     (playing indicator)
//   orange      #ffb86c  soft orange      (params, time, paused)
//   pink        #ff79c6  hot pink         (keywords: function, return)
//   red         #ff5555  red              (errors)
//   yellow      #f1fa8c  pale yellow      (strings, template literals)
//   dim         #44475a  dark gray        (separators, inactive)

const (
	clrFg      = "#f8f8f2"
	clrComment = "#6272a4"
	clrCyan    = "#8be9fd"
	clrGreen   = "#50fa7b"
	clrOrange  = "#ffb86c"
	clrPink    = "#ff79c6"
	clrRed     = "#ff5555"
	clrYellow  = "#f1fa8c"
	clrDim     = "#44475a"
)

// ─── Playback state ──────────────────────────────────────────────────────────

var (
	playingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrGreen))
	pausedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(clrOrange))
	loadingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrComment))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(clrRed))
)

// ─── Syntax — preamble ───────────────────────────────────────────────────────

var (
	keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrPink)).Italic(true)
	fnNameStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFg))
	paramStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(clrOrange)).Italic(true)
	stringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(clrYellow))
	punctStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFg))
	fgStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFg))
)

// ─── UI tokens ───────────────────────────────────────────────────────────────

var (
	bracketStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrCyan))
	timeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFg))
	linkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(clrCyan))
	sepStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(clrDim))
	commentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrComment))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(clrDim))
)

// ─── Episode list ─────────────────────────────────────────────────────────────

var (
	// epNumStyle: right-pane episode numbers  "78:"
	epNumStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrCyan))
	// epTitleStyle: right-pane artist names (italic cyan, like the site)
	epTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrCyan)).Italic(true)
	// epCurrentStyle: the episode currently playing (brighter, bold)
	epCurrentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrCyan)).Bold(true)
	// epSelectedStyle: cursor position when not the playing episode
	epSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFg)).Bold(true)
	// epDimStyle: all other episodes
	epDimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrComment)).Italic(true)
)

// ─── Center pane ─────────────────────────────────────────────────────────────

var (
	// episodeTitleStyle: large center-pane title
	episodeTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFg))
	// trackStyle: tracklist lines  "Artist - Title"
	trackArtistStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFg))
	trackSepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(clrComment))
	trackTitleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFg))
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

// tok renders a [bracketed] control token in cyan.
func tok(s string) string {
	return bracketStyle.Render("[" + s + "]")
}
