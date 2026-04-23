package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/fpigeonjr/music-for-coding-tui/internal/feed"
	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
)

// ─── Preamble ────────────────────────────────────────────────────────────────

const preamble = "function musicFor(task = 'programming') {\n  return `A series of mixes\n  intended for listening\n  while ${task} to focus\n  the brain and inspire\n  the mind.`;\n}"

// ─── Top-level View ──────────────────────────────────────────────────────────

func (m model) View() string {
	if m.width < minWidth || m.height < minHeight {
		return fmt.Sprintf(
			"\n  Terminal too small (min %d×%d, got %d×%d)\n  Please resize and try again.\n",
			minWidth, minHeight, m.width, m.height,
		)
	}

	if m.err != nil {
		return fmt.Sprintf("\n  %s\n\n  %s\n",
			errorStyle.Render("[error]"),
			errorStyle.Render(m.err.Error()),
		)
	}

	left, center, right := m.paneWidths()
	h := m.height - 1 // leave one line for terminal cursor

	leftPane := lipgloss.NewStyle().
		Width(left).
		Height(h).
		PaddingRight(2).
		Render(m.renderLeft(left - 2))

	centerPane := lipgloss.NewStyle().
		Width(center).
		Height(h).
		PaddingLeft(1).
		PaddingRight(1).
		Render(m.renderCenter(center - 2))

	rightPane := lipgloss.NewStyle().
		Width(right).
		Height(h).
		PaddingLeft(2).
		Render(m.renderRight(right - 2))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, centerPane, rightPane)
}

// ─── Left pane ───────────────────────────────────────────────────────────────

func (m model) renderLeft(width int) string {
	sep := dimStyle.Render(strings.Repeat("-", width))

	// Preamble
	pre := dimStyle.Render(preamble)

	// Current episode info (truncated to width)
	ep := m.currentEpisode()
	epLine := ""
	if ep.Number > 0 {
		epLine = truncate(fmt.Sprintf("• Episode %d: %s", ep.Number, ep.Title), width)
		epLine = dimStyle.Render(epLine)
	}

	// Transport controls
	transport := dimStyle.Render("[prev] [-30] [stop] [+30] [next]")

	// Time + volume row
	pos := player.FormatDuration(m.state.Position)
	timeVol := dimStyle.Render(fmt.Sprintf("%s [v-] 100%% [v+] [random]", pos))

	// Links
	links := dimStyle.Render("[about] [credits] [rss.xml]\n[patreon] [podcasts.apple]\n[folder.jpg] [invert]")

	// Stats
	stats := m.renderStats()

	// Help
	help := dimStyle.Render("space play/pause\n←/→  seek ±30s\np/n  prev/next\nj/k  browse list\nenter load selected\nq    quit")

	return strings.Join([]string{
		pre, sep, epLine, transport, timeVol, sep, stats, sep, links, sep, help,
	}, "\n")
}

// ─── Center pane ─────────────────────────────────────────────────────────────

func (m model) renderCenter(width int) string {
	ep := m.currentEpisode()

	if m.loading || len(m.episodes) == 0 {
		return loadingStyle.Render("[loading] ...")
	}

	// Large episode title (may wrap)
	titleText := fmt.Sprintf("Episode %d:\n%s", ep.Number, ep.Title)
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EEEEEE")).
		Width(width).
		Render(titleText)

	// Control row
	dur := player.FormatDuration(m.state.Duration)
	sizeMB := float64(ep.Size) / 1_000_000
	controls := m.renderCenterControls(dur, sizeMB)

	// Tracklist
	tracklist := m.renderTracklist()

	return strings.Join([]string{title, "", controls, "", tracklist}, "\n")
}

func (m model) renderCenterControls(dur string, sizeMB float64) string {
	var stop string
	if m.state.Paused {
		stop = pausedStyle.Render("[play]")
	} else {
		stop = playingStyle.Render("[stop]")
	}

	size := dimStyle.Render(fmt.Sprintf("[source] %.0f MB", sizeMB))
	fav := dimStyle.Render("[favourite]")

	pos := player.FormatDuration(m.state.Position)
	elapsed := timeStyle.Render(fmt.Sprintf("%s / %s", pos, dur))

	return fmt.Sprintf("%s %s\n%s\n%s", stop, elapsed, size, fav)
}

func (m model) renderTracklist() string {
	if m.tracksFetching {
		return loadingStyle.Render("fetching tracklist...")
	}
	if len(m.tracks) == 0 {
		return dimStyle.Render("no tracklist available")
	}

	lines := make([]string, len(m.tracks))
	for i, t := range m.tracks {
		if t.Artist != "" {
			lines[i] = dimStyle.Render(fmt.Sprintf("%s - %s", t.Artist, t.Title))
		} else {
			lines[i] = dimStyle.Render(t.Title)
		}
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
		line := fmt.Sprintf("%2d: %s", ep.Number, ep.Title)
		line = truncate(line, width)

		switch {
		case i == m.currentIdx && i == m.selectedIdx:
			// Playing AND cursor here
			sb.WriteString(selectedStyle.Render("▶ " + line))
		case i == m.currentIdx:
			// Playing but cursor elsewhere
			sb.WriteString(currentStyle.Render("▶ " + line))
		case i == m.selectedIdx:
			// Cursor, not playing
			sb.WriteString(selectedStyle.Render("  " + line))
		default:
			sb.WriteString(dimStyle.Render("  " + line))
		}

		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	// Scroll indicators
	var indicators []string
	if m.listOffset > 0 {
		indicators = append(indicators, dimStyle.Render("  ↑ more"))
	}
	if end < len(m.episodes) {
		indicators = append(indicators, dimStyle.Render("  ↓ more"))
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
		return dimStyle.Render("// loading...")
	}
	stats := feed.ComputeStats(m.episodes)
	h := stats.TotalSeconds / 3600
	mins := (stats.TotalSeconds % 3600) / 60
	secs := stats.TotalSeconds % 60

	return dimStyle.Render(fmt.Sprintf(
		"// %d episodes\n// %d hours\n// %d minutes\n// %d seconds",
		stats.Episodes, h, mins, secs,
	))
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// truncate clips s to maxLen runes, adding "…" if clipped.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}

// renderStatus is used by tests and the loading/error early-return paths.
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
