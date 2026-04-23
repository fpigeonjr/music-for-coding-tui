# music-for-coding-tui

A terminal UI for [musicforprogramming.net](https://musicforprogramming.net) — Go + Bubble Tea.

A phased exploration of a TUI that matches the Music For Programming website's
three-pane monospace aesthetic, with actual audio playback via `mpv`.

## Status

Phase 1 — audio plumbing spike. See [milestones](../../milestones) for progress.

## Prerequisites

- Go 1.22+
- [`mpv`](https://mpv.io) (`brew install mpv`)

## Plan

| Phase | Goal |
|-------|------|
| 1 | Audio plumbing spike — drive `mpv` from Bubble Tea, play one hard-coded MFP episode |
| 2 | RSS + episode model — parse `musicforprogramming.net/rss.xml`, cycle episodes |
| 3 | Three-pane layout (unstyled) — left transport, center tracklist, right index |
| 4 | MFP aesthetic pass — colors, typography, bracketed control tokens |
| 5 | Niceties — favorites, random, volume, resume position |

## License

MIT
