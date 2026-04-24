# mfp

> **Unofficial client** — not affiliated with or endorsed by [musicforprogramming.net](https://musicforprogramming.net) or Datassette, but built with his blessing. 👍 All audio content belongs to its respective artists. This tool streams directly from MFP's own servers — no content is hosted or redistributed.

A terminal UI for [musicforprogramming.net](https://musicforprogramming.net) — Go + Bubble Tea.

Streams all 78 MFP episodes in your terminal with a three-pane layout, full
tracklist display, and a Dracula-inspired colour palette that mirrors the original site.

![Phase 4 screenshot](docs/screenshot-phase4.png)

---

## Install

### Homebrew (recommended — no Go required)

```bash
brew tap fpigeonjr/homebrew-tap
brew install mfp
```

> `mpv` is installed automatically as a dependency.

### go install (requires Go 1.22+)

```bash
go install github.com/fpigeonjr/music-for-coding-tui/cmd/mfp@latest
```

> Requires `mpv` separately: `brew install mpv`

### Build from source

```bash
git clone https://github.com/fpigeonjr/music-for-coding-tui.git
cd music-for-coding-tui
make install
```

---

## Usage

```bash
mfp              # launch
mfp --version    # print version
```

---

## Controls

| Key | Action |
|-----|--------|
| `space` | Play / pause |
| `→` or `l` | Seek forward 30s |
| `←` or `h` | Seek back 30s |
| `n` / `]` | Next (older) episode |
| `p` / `[` | Previous (newer) episode |
| `j` / `↓` | Scroll episode list down |
| `k` / `↑` | Scroll episode list up |
| `enter` | Play selected episode |
| `r` | Random episode |
| `f` | Toggle ★ favourite |
| `-` / `=` | Volume down / up |
| `q` / `Ctrl+C` | Quit |

---

## Persistent state

All state is saved to `~/.config/music-for-coding/`:

| File | Contents |
|------|----------|
| `positions.json` | Resume position per episode |
| `favourites.json` | Starred episode numbers |
| `volume.json` | Last used volume level |

---

## Layout

```
┌─ Left ──────────────┬─ Center ──────────────────┬─ Right ──────────────┐
│ function musicFor(  │ Episode 78:               │ ▶ 78: Datassette     │
│   task='programming'│ Datassette                │   77: Phonaut        │
│ ) { return `...` }  │                           │   76: Material Object│
│ ─────────────────── │ [stop] 01:46 / 1:29:59    │   75: Datassette     │
│ • Episode 78: ...   │ [source] 159 MB           │   74: NCW            │
│ [prev][-30][stop]   │ [favourite]               │   ...                │
│ [+30][next]         │                           │                      │
│ ─────────────────── │ David Borden - Enfield... │                      │
│ // 78 episodes      │ Datassette - rain_wind... │                      │
│ // 92 hours         │ ...                       │                      │
│ ─────────────────── │                           │                      │
│ [about][credits]... │                           │                      │
└─────────────────────┴───────────────────────────┴──────────────────────┘
```

---

## Development

```bash
make run        # run from source
make build      # compile → ./music-for-coding-tui
make install    # install mfp → $GOPATH/bin
make test       # unit tests (no network, no mpv required)
make test-full  # all tests including live RSS + mpv integration
make lint       # go vet
make tidy       # go mod tidy
```

---

## Releasing

Tagging a version triggers the full release pipeline automatically:

```bash
git tag v0.x.0
git push origin v0.x.0
```

GitHub Actions will:
1. Run all tests
2. Build binaries for darwin/arm64, darwin/amd64, linux/arm64, linux/amd64
3. Publish a GitHub Release with all artifacts
4. Update `Formula/mfp.rb` in `fpigeonjr/homebrew-tap` automatically

---

## Roadmap

| Phase | Goal | Status |
|-------|------|--------|
| 1 | Audio plumbing — mpv IPC, play/pause/seek, status line | ✅ Done |
| 2 | RSS + episode model — parse feed, prev/next navigation | ✅ Done |
| 3 | Three-pane layout — left transport, center tracklist, right index | ✅ Done |
| 4 | MFP aesthetic — Dracula palette, syntax-highlighted preamble, cyan `[tokens]` | ✅ Done |
| 5 | Niceties — favorites, random, volume, resume position | ✅ Done |
| 6 | Distribution — `go install` + `mfp --version`, v0.1.0 tagged | ✅ Done |
| 7 | Homebrew tap — `brew tap fpigeonjr/homebrew-tap && brew install mfp` | ✅ Done |
| 8 | goreleaser — pre-built arm64/amd64 binaries, automated tap updates | ✅ Done |
| 9 | homebrew-core — `brew install mfp` with no tap | 🔜 Next |

---

## Testing

```bash
make test       # 49 unit tests — no mpv or network required
make test-full  # + live RSS + tracklist + mpv integration tests
```

See [docs/phase-1-smoketest.md](docs/phase-1-smoketest.md) for the manual QA checklist.

---

## Architecture

```
cmd/mfp/
  main.go    — entry point, --version flag
  model.go   — model struct, messages, pane geometry, scroll logic
  update.go  — Init(), Update(), all tea.Cmd functions
  view.go    — View(), all render* helpers, preamble syntax highlight
  styles.go  — Dracula colour palette + Lip Gloss style registry

internal/
  player/    — mpv IPC client (spawn, load, pause, seek, volume, get_state)
  feed/      — RSS fetch + parse, tracklist scraping, 1h disk cache
  store/     — persistent state (favourites, positions, volume)
```

---

## License

MIT
