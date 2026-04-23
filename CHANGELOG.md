# Changelog

All notable changes to this project will be documented in this file.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.2.0]: https://github.com/fpigeonjr/music-for-coding-tui/releases/tag/v0.2.0
[0.1.0]: https://github.com/fpigeonjr/music-for-coding-tui/releases/tag/v0.1.0
