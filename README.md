# music-for-coding-tui

A terminal UI for [musicforprogramming.net](https://musicforprogramming.net) — Go + Bubble Tea.

Streams MFP episodes in your terminal with playback controls, a scrollable episode
index, and a three-pane layout inspired by the original site's monospace aesthetic.

<img width="1128" height="542" alt="image" src="https://github.com/user-attachments/assets/089acf0c-615c-4c0c-9482-f9234d9af55d" />


## Status

**Phase 1 complete** — audio plumbing spike. mpv IPC client working, Episode 78
plays and responds to controls.

See the [project board](https://github.com/users/fpigeonjr/projects/4) for progress.

---

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.22+ | `brew install go` |
| mpv | any | `brew install mpv` |

---

## Quick start

```bash
git clone https://github.com/fpigeonjr/music-for-coding-tui.git
cd music-for-coding-tui

make run
```

---

## Controls

| Key | Action |
|-----|--------|
| `space` | Play / pause |
| `→` or `l` | Seek forward 30s |
| `←` or `h` | Seek back 30s |
| `q` / `Ctrl+C` | Quit |

---

## Development

```bash
make run      # run from source
make build    # compile binary → ./music-for-coding-tui
make test     # run all tests (unit + integration)
make lint     # go vet
make tidy     # go mod tidy
```

---

## Roadmap

| Phase | Goal | Status |
|-------|------|--------|
| 1 | Audio plumbing — mpv IPC, play/pause/seek, status line | ✅ Done |
| 2 | RSS + episode model — parse feed, prev/next navigation | 🔜 Next |
| 3 | Three-pane layout — left transport, center tracklist, right index | ⏳ |
| 4 | MFP aesthetic — colors, syntax-highlighted preamble, bracketed tokens | ⏳ |
| 5 | Niceties — favorites, random, volume, resume position | ⏳ |

---

## Testing

```bash
make test
# 10 unit tests  (cmd/mfp     — no mpv required)
#  7 integration (player pkg  — requires mpv)
```

See [docs/phase-1-smoketest.md](docs/phase-1-smoketest.md) for the full manual QA checklist.

---

## License

MIT
