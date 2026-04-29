package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/fpigeonjr/music-for-coding-tui/internal/feed"
	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
)

// ─── Top-level View ──────────────────────────────────────────────────────────

func (m model) View() string {
	if m.width < minWidth || m.height < minHeight {
		return fmt.Sprintf(
			"\n  %s\n\n  Terminal too small (min %d×%d, got %d×%d). Please resize.\n",
			errorStyle.Render("[error]"), minWidth, minHeight, m.width, m.height,
		)
	}
	if m.err != nil {
		return fmt.Sprintf("\n  %s\n\n  %s\n",
			errorStyle.Render("[error]"), errorStyle.Render(m.err.Error()),
		)
	}

	if m.showHelp {
		return m.renderHelpOverlay()
	}

	left, center, right := m.paneWidths()
	h := m.height - 1

	leftPane := lipgloss.NewStyle().
		Width(left).Height(h).PaddingRight(2).
		Render(m.renderLeft(left - 2))

	centerPane := lipgloss.NewStyle().
		Width(center).Height(h).PaddingLeft(1).PaddingRight(1).
		Render(m.renderCenter(center-2, h))

	rightPane := lipgloss.NewStyle().
		Width(right).Height(h).PaddingLeft(2).
		Render(m.renderRight(right - 2))

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, centerPane, rightPane)

	if m.commandMode || m.commandResult {
		return lipgloss.JoinVertical(lipgloss.Top, mainContent, m.renderCommandBar())
	}

	return mainContent
}

// ─── Help overlay ────────────────────────────────────────────────────────────

func (m model) renderHelpOverlay() string {
	header := bracketStyle.Render("keybindings") +
		commentStyle.Render("  -  press esc or ? to close")

	sep := sepStyle.Render(strings.Repeat("─", 46))

	col := func(key, desc string) string {
		return bracketStyle.Render(fmt.Sprintf("%-12s", key)) +
			fgStyle.Render(desc)
	}

	playback := strings.Join([]string{
		commentStyle.Render("PLAYBACK"),
		col("space", "play / pause"),
		col("← / h", "seek back 30s"),
		col("→ / l", "seek forward 30s"),
		col("p / [", "previous episode"),
		col("n / ]", "next episode"),
		col("r", "random episode"),
		col("f", "toggle favourite"),
		col("-", "volume -10%"),
		col("=", "volume +10% (max 150%)"),
	}, "\n")

	navigation := strings.Join([]string{
		commentStyle.Render("NAVIGATION"),
		col("j / ↓", "scroll list down"),
		col("k / ↑", "scroll list up"),
		col("enter", "play selected episode"),
		"",
		commentStyle.Render("GENERAL"),
		col("t", "cycle theme"),
		col("?", "this help"),
		col("/", "command bar  (theme, jump, vol, fav, random, quit)"),
		col("q / ctrl+c", "quit"),
	}, "\n")

	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(32).Render(playback),
		lipgloss.NewStyle().Width(28).Render(navigation),
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.Bracket)).
		Padding(1, 2).
		Render(strings.Join([]string{header, sep, cols}, "\n"))

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center, box)
}

// ─── Left pane ───────────────────────────────────────────────────────────────

func (m model) renderLeft(width int) string {
	sep := sepStyle.Render(strings.Repeat("-", width))

	ep := m.currentEpisode()
	epLine := ""
	if ep.Number > 0 {
		epLine = dimStyle.Render("• ") +
			epNumStyle.Render(fmt.Sprintf("Episode %d:", ep.Number)) + " " +
			epTitleStyle.Render(truncate(ep.Title, width-14))
	}

	// Transport row
	// (removed - decorative only, use keyboard shortcuts below)

	// Time + volume row - volume shown in orange when boosted above 100%
	pos := player.FormatDuration(m.state.Position)
	volStyle := fgStyle
	if m.volume > 100 {
		volStyle = pausedStyle // orange = boost mode signal
	}
	timeVol := commentStyle.Render("pos:") + " " + timeStyle.Render(pos) +
		commentStyle.Render(" | ") +
		commentStyle.Render("vol:") + " " +
		volStyle.Render(fmt.Sprintf("%d%%", m.volume))

	// Links
	links := fmt.Sprintf("%s %s %s\n%s %s\n%s %s",
		hyperlink("about", "https://musicforprogramming.net/about"),
		hyperlink("credits", "https://musicforprogramming.net/credits"),
		hyperlink("rss.xml", "https://musicforprogramming.net/rss.xml"),
		hyperlink("patreon", "https://www.patreon.com/datassette"),
		hyperlink("podcasts.apple", "https://podcasts.apple.com/us/podcast/music-for-programming/id500565620"),
		hyperlink("folder.jpg", "https://musicforprogramming.net/img/folder.jpg"),
		hyperlink("invert", "https://musicforprogramming.net"),
	)

	// Stats
	stats := m.renderStats()

	// Help (dim - secondary info)
	help := strings.Join([]string{
		dimStyle.Render("space  play/pause"),
		dimStyle.Render("←/→    seek ±30s"),
		dimStyle.Render("p/n    prev/next"),
		dimStyle.Render("r      random"),
		dimStyle.Render("-/=    volume (0-150%)"),
		dimStyle.Render("f      favourite"),
		dimStyle.Render("t      cycle theme"),
		dimStyle.Render("j/k    browse list"),
		dimStyle.Render("enter  load selected"),
		dimStyle.Render("?      keybindings"),
		dimStyle.Render("q      quit"),
	}, "\n")

	// Theme flash - shown briefly after switching
	themeFlash := ""
	if m.themeMsg != "" {
		themeFlash = commentStyle.Render("theme: ") + bracketStyle.Render(m.themeMsg)
	}

	return strings.Join([]string{
		renderPreamble(), sep,
		epLine, timeVol, sep,
		stats, sep,
		links, sep,
		help, themeFlash,
	}, "\n")
}

// ─── Center pane ─────────────────────────────────────────────────────────────

func (m model) renderCenter(width, height int) string {
	ep := m.currentEpisode()

	if m.loading || len(m.episodes) == 0 {
		return loadingStyle.Render("[loading] ...")
	}

	// Large episode title
	titleText := fmt.Sprintf("Episode %d:\n%s", ep.Number, ep.Title)
	title := episodeTitleStyle.Width(width).Render(titleText)

	// Control row
	var stopTok string
	if m.state.Paused {
		stopTok = pausedStyle.Render("[play]")
	} else {
		stopTok = playingStyle.Render("[stop]")
	}
	pos := player.FormatDuration(m.state.Position)
	dur := player.FormatDuration(m.state.Duration)

	// Favourite indicator
	favTok := tok("favourite")
	if m.favourites[ep.Number] {
		favTok = bracketStyle.Render("[favourite ★]")
	}

	controls := fmt.Sprintf("%s %s\n%s %.0f MB\n%s",
		stopTok,
		timeStyle.Render(fmt.Sprintf("%s / %s", pos, dur)),
		tok("source"), float64(ep.Size)/1_000_000,
		favTok,
	)

	// title(2) + blank(1) + controls(3) + blank(1) = 7 fixed lines
	const fixedLines = 7
	maxTrackLines := height - fixedLines
	if maxTrackLines < 3 {
		maxTrackLines = 3
	}

	tracklist := m.renderTracklist(maxTrackLines)

	return strings.Join([]string{title, "", controls, "", tracklist}, "\n")
}

func (m model) renderTracklist(maxLines int) string {
	if m.tracksFetching {
		return loadingStyle.Render("fetching tracklist...")
	}
	if len(m.tracks) == 0 {
		return commentStyle.Render("no tracklist available")
	}
	lines := make([]string, len(m.tracks))
	for i, t := range m.tracks {
		if t.Artist != "" {
			lines[i] = trackArtistStyle.Render(t.Artist) +
				trackSepStyle.Render(" - ") +
				trackTitleStyle.Render(t.Title)
		} else {
			lines[i] = trackTitleStyle.Render(t.Title)
		}
	}
	// Episode link at the bottom of the tracklist
	ep := m.currentEpisode()
	if ep.Slug != "" {
		epURL := "https://musicforprogramming.net/" + ep.Slug
		lines = append(lines, "", hyperlink(epURL, epURL))
	}
	// Truncate to fit available height, showing how many tracks were clipped
	if maxLines > 0 && len(lines) > maxLines {
		clipped := len(lines) - (maxLines - 1)
		lines = lines[:maxLines-1]
		lines = append(lines, commentStyle.Render(fmt.Sprintf("  ↓ %d more", clipped)))
	}
	return strings.Join(lines, "\n")
}

// ─── Right pane ──────────────────────────────────────────────────────────────

func (m model) renderRight(width int) string {
	if len(m.episodes) == 0 {
		return loadingStyle.Render("loading...")
	}

	visible := m.rightPaneHeight()
	end := m.listOffset + visible
	if end > len(m.episodes) {
		end = len(m.episodes)
	}

	var sb strings.Builder
	for i := m.listOffset; i < end; i++ {
		ep := m.episodes[i]
		star := ""
		if m.favourites[ep.Number] {
			star = "★"
		}
		num := fmt.Sprintf("%2d: ", ep.Number)
		title := truncate(ep.Title+star, width-len(num)-2)

		var line string
		switch {
		case i == m.currentIdx && i == m.selectedIdx:
			// Playing + cursor: bright cyan bold, play marker
			line = epCurrentStyle.Render("▶ "+num) + epCurrentStyle.Render(title)
		case i == m.currentIdx:
			// Playing, cursor elsewhere: cyan bold with marker
			line = epCurrentStyle.Render("▶ "+num+title)
		case i == m.selectedIdx:
			// Cursor, not playing: bright foreground so it pops from cyan list
			line = epSelectedStyle.Render("  "+num+title)
		default:
			// All others: muted cyan italic (matches MFP site)
			line = epDimStyle.Render("  "+num) + epTitleStyle.Render(title)
		}
		sb.WriteString(line)
		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	var indicators []string
	if m.listOffset > 0 {
		indicators = append(indicators, commentStyle.Render("  ↑ more"))
	}
	if end < len(m.episodes) {
		indicators = append(indicators, commentStyle.Render("  ↓ more"))
	}
	if len(indicators) > 0 {
		sb.WriteByte('\n')
		sb.WriteString(strings.Join(indicators, " "))
	}

	return sb.String()
}

// ─── Stats ───────────────────────────────────────────────────────────────────

func (m model) renderStats() string {
	if len(m.episodes) == 0 {
		return commentStyle.Render("// loading...")
	}
	stats := feed.ComputeStats(m.episodes)
	h := stats.TotalSeconds / 3600
	mins := (stats.TotalSeconds % 3600) / 60
	secs := stats.TotalSeconds % 60
	return commentStyle.Render(fmt.Sprintf(
		"// %d episodes\n// %d hours\n// %d minutes\n// %d seconds",
		stats.Episodes, h, mins, secs,
	))
}

// ─── Preamble syntax highlight ───────────────────────────────────────────────

// renderPreamble produces the syntax-highlighted function musicFor(...) block.
func renderPreamble() string {
	kw := func(s string) string { return keywordStyle.Render(s) }
	fn := func(s string) string { return fnNameStyle.Render(s) }
	pm := func(s string) string { return paramStyle.Render(s) }
	st := func(s string) string { return stringStyle.Render(s) }
	pu := func(s string) string { return punctStyle.Render(s) }
	fg := func(s string) string { return fgStyle.Render(s) }

	lines := []string{
		kw("function") + " " + fn("musicFor") + pu("(") + pm("task") + " = " + st("'programming'") + pu(") {"),
		"  " + kw("return") + " " + pu("`") + st("A series of mixes"),
		"  " + st("intended for listening"),
		"  " + st("while ") + pu("${") + pm("task") + pu("}") + st(" to focus"),
		"  " + st("the brain and inspire"),
		"  " + st("the mind.") + pu("`") + fg("; }"),
	}
	return strings.Join(lines, "\n")
}

// ─── renderStatus (used by tests + error paths) ──────────────────────────────

func (m model) renderStatus() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("[error] %v", m.err))
	}
	if m.loading || !m.playerReady {
		return loadingStyle.Render("[starting] ...")
	}
	if !m.state.Loaded {
		return loadingStyle.Render("[loading] ...")
	}
	pos := player.FormatDuration(m.state.Position)
	dur := player.FormatDuration(m.state.Duration)
	elapsed := timeStyle.Render(fmt.Sprintf("%s / %s", pos, dur))
	if m.state.Paused {
		return fmt.Sprintf("%s  %s", pausedStyle.Render("[paused]"), elapsed)
	}
	return fmt.Sprintf("%s  %s", playingStyle.Render("[playing]"), elapsed)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// truncate clips s to maxLen runes, appending "…" if clipped.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-1]) + "…"
}

// ─── Command Bar ─────────────────────────────────────────────────────────────

func (m model) renderCommandBar() string {
	m.commandInput.Width = m.width - 4

	var content string
	if m.commandMode {
		// Typing mode: show the live textinput + autocomplete ghost text
		content = m.commandInput.View()
		if hint := m.getAutocompleteHint(); hint != "" {
			hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.CommandBarHint))
			content += hintStyle.Render(hint)
		}
	} else {
		// Result display mode: just show the prompt character
		content = commentStyle.Render(":")
	}

	if m.commandError != "" {
		content += "  " + errorStyle.Render(m.commandError)
	} else if m.commandOutput != "" {
		content += "  " + commentStyle.Render(m.commandOutput)
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color(m.theme.CommandBarBg)).
		Foreground(lipgloss.Color(m.theme.Fg)).
		Width(m.width).
		Padding(0, 1).
		Render(content)
}

// getAutocompleteHint returns inline ghost-text for the current command input.
func (m model) getAutocompleteHint() string {
	input := m.commandInput.Value()
	if input == "" {
		return "  theme · jump · vol · fav · random · help · quit"
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ""
	}
	command := parts[0]

	switch command {
	case "theme":
		if len(parts) <= 1 {
			return "  dracula · nord · gruvbox · onedark · everforest"
		}
		// Ghost-text prefix completion
		partial := strings.ToLower(strings.Join(parts[1:], " "))
		if strings.HasSuffix(input, " ") {
			partial += " "
		}
		for _, t := range Themes {
			name := strings.ToLower(t.Name)
			if strings.HasPrefix(name, partial) && name != partial {
				return name[len(partial):]
			}
		}
	case "vol", "volume":
		if len(parts) <= 1 {
			return "  <0-150>"
		}
	case "jump":
		if len(parts) <= 1 {
			return "  <episode-number or title>"
		}
	case "fav", "favourite", "favorite":
		if len(parts) <= 1 {
			return "  toggle favourite on current episode"
		}
	case "random":
		return "  play a random episode"
	case "help":
		return "  open keybindings overlay"
	case "quit", "exit":
		return "  quit mfp"
	}

	return ""
}
