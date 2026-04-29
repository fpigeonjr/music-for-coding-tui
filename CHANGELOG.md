# Changelog

All notable changes to this project will be documented in this file.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-04-28

### Added
- `/` opens a vim-style command bar at the bottom of the screen
- Commands: `theme`, `jump`, `vol`, `fav`, `random`, `help`, `quit`/`exit`
- `theme` supports short aliases (`gruvbox`, `onedark`, `one`, `everforest`) and hyphen/underscore normalisation (`gruvbox-dark`, `one_dark`)
- `jump <n>` jumps to episode by number; `jump <query>` fuzzy-matches by title
- `vol <0-150>` sets volume directly
- `fav` toggles favourite on the current episode (also: `favourite`, `favorite`)
- `random` picks a random episode with confirmation in the bar
- Autocomplete ghost-text hints for every command and alias
- `CommandBarHint` colour added to all themes — tuned for 3–4.2:1 contrast against the bar background (previously hints were near-invisible on all themes using the `Dim` colour)
- `CommandBarBg` added to `ThemeNord` (was missing)
- 67 new tests covering command dispatch, all handlers, hints, bar rendering, theme contrast, and tracklist truncation

### Fixed
- Command bar feedback (output/errors) was invisible — bar closed before the result message arrived; fixed with a separate `commandResult` display state that keeps the bar visible until the auto-clear timer fires
- Centre pane overflow on episodes with long tracklists — content now truncates to fit terminal height with a `↓ N more` indicator (matches right-pane scroll pattern)
- `truncate` helper was appending `...` (3 bytes, exceeding `maxLen`) — reverted to `…` (1 rune)
- Removed dead `commandTabIndex` / `commandCompletions` model fields
- Removed dev-only `gh-issue` command from user-facing dispatch
- Fixed broken indentation in `fetchTracklistCmd` and all downstream helpers

## [0.3.0] - 2026-04-24

### Added
- 5 colour themes: Dracula (default), Nord, Gruvbox Dark, One Dark, Everforest Dark
- `t` key cycles themes — selection persists to `~/.config/music-for-coding/theme.json`
- `?` key opens a full keybindings overlay (two-column, rounded border) — playback never interrupted
- OSC 8 clickable hyperlinks in left pane (`[about]`, `[credits]`, `[rss.xml]`, `[patreon]`, `[podcasts.apple]`, `[folder.jpg]`) and center pane episode URL
- Volume boost indicator — `vol:` turns orange above 100% to signal boost mode
- `pos: xx:xx | vol: xxx%` labels in left pane for clarity
- Last-played episode now restored on relaunch (`last-episode.json`)
- All session state fully persisted: episode, position, volume, theme, favourites

### Changed
- Removed non-functional `[prev] [-30] [stop] [+30] [next]` transport tokens — decorative only, replaced by key reference
- Volume key reference updated to show `0–150%` range

### Fixed
- Last-played episode was lost on quit — now saved to `~/.config/music-for-coding/last-episode.json`
- Config directory now uses `~/.config` (XDG) consistently on macOS instead of `~/Library/Application Support`

### Known issues
- OSC 8 hyperlinks not clickable in Ghostty/tmux via `Cmd+click` — upstream bug [ghostty#11907](https://github.com/ghostty-org/ghostty/issues/11907). Works in iTerm2, kitty, WezTerm.

## [0.2.0] - 2026-04-23

### Added
- Homebrew tap: `brew tap fpigeonjr/homebrew-tap && brew install mfp`
- goreleaser: pre-built binaries for darwin/arm64, darwin/amd64, linux/arm64, linux/amd64
- GitHub Actions release workflow: auto-builds + publishes on every `v*` tag
- Homebrew formula auto-updated by goreleaser on release (no manual SHA update needed)
- Volume persisted across sessions (`~/.config/music-for-coding/volume.json`)
- `mfp --version` / `mfp -v` flag

### Fixed
- Config dir uses `~/.config` (XDG) instead of `~/Library/Application Support` on macOS
- Homebrew tap install path no longer documented before the tap existed

## [0.1.0] - 2026-04-23

First public release. All five core phases complete.

### Added

**Audio playback (Phase 1)**
- `mpv` subprocess driven via JSON IPC socket
- Play, pause, seek ±30s, clean shutdown with no orphan processes
- `FormatDuration` helper for `mm:ss` / `h:mm:ss` display

**RSS + episode model (Phase 2)**
- Fetches and parses `musicforprogramming.net/rss.xml`
- RSS cached to `~/.cache/music-for-coding/rss.xml` (1h TTL, network-failure fallback)
- Tracklists fetched lazily from individual episode HTML pages (`__SAPPER__` JSON)
- `p` / `n` keys cycle through all 78 episodes
- `ComputeStats` aggregates total episodes and runtime

**Three-pane layout (Phase 3)**
- Left pane: `function musicFor()` preamble, transport controls, stats, key reference
- Center pane: large episode title, `[stop]` / `[source]` / `[favourite]` controls, tracklist
- Right pane: scrollable episode index with `▶` playing marker and cursor highlight
- `j` / `k` browse list without interrupting playback; `enter` loads selected episode
- Graceful reflow on terminal resize; minimum 80×20 size guard

**MFP aesthetic (Phase 4)**
- Dracula-inspired palette: pink keywords, orange params, yellow strings, cyan `[tokens]`
- Token-by-token syntax highlighting of the `function musicFor()` preamble
- Episode list: muted cyan italic for all, bold cyan `▶` for current, bold white for cursor
- Central `styles.go` registry — no inline colours anywhere

**Niceties (Phase 5)**
- `f` — toggle ★ favourite on current episode, persisted to `~/.config/music-for-coding/favourites.json`
- `r` — random episode (never repeats current)
- `-` / `=` — volume down/up ±10% (clamped 0–150)
- Resume position: saved on every tick, restored automatically on relaunch
- Positions persisted to `~/.config/music-for-coding/positions.json`

### Prerequisites
- Go 1.22+
- `mpv` (`brew install mpv`)

[0.4.0]: https://github.com/fpigeonjr/music-for-coding-tui/releases/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/fpigeonjr/music-for-coding-tui/releases/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/fpigeonjr/music-for-coding-tui/releases/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/fpigeonjr/music-for-coding-tui/releases/tag/v0.1.0
