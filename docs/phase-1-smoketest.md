# Phase 1 Smoke Test Checklist

Manual QA checklist for the Phase 1 audio plumbing spike.
Run this before closing the Phase 1 milestone.

## Prerequisites

- [ ] macOS (Apple Silicon or Intel)
- [ ] Go 1.22+ installed (`go version`)
- [ ] mpv installed (`mpv --version`) — install with `brew install mpv`
- [ ] Active internet connection (Episode 78 streams from datashat.net)

---

## 1. Fresh clone

```bash
git clone https://github.com/fpigeonjr/music-for-coding-tui.git
cd music-for-coding-tui
```

Expected: no build artifacts present.

---

## 2. Build

```bash
make build
```

- [ ] Exits with code 0
- [ ] Binary `./music-for-coding-tui` created

---

## 3. Unit + integration tests

```bash
make test
```

- [ ] All 17 tests pass
- [ ] No tests fail or hang (timeout: 90s)
- [ ] `cmd/mfp` shows 10 unit tests (no mpv required)
- [ ] `internal/player` shows 7 tests (requires mpv)

---

## 4. Launch and audio

```bash
make run
```

- [ ] TUI opens in alt-screen (terminal clears)
- [ ] Title shows `Episode 78: Datassette` in cyan
- [ ] Status briefly shows `[starting] ...` then `[loading] ...`
- [ ] Status transitions to `[playing]  00:00 / 1:29:59` within ~5s
- [ ] Timer ticks forward in real time
- [ ] **Audio is audible** through speakers/headphones

---

## 5. Playback controls

With the TUI running:

### Play / Pause
- [ ] Press `space` → status changes to `[paused]` in gold, timer stops
- [ ] Press `space` again → status returns to `[playing]` in green, timer resumes

### Seek forward
- [ ] Press `→` (right arrow) → timer jumps forward ~30s
- [ ] Press `l` → same result

### Seek backward
- [ ] Press `←` (left arrow) → timer jumps back ~30s
- [ ] Press `h` → same result

---

## 6. Clean quit via `q`

```bash
# While TUI is running, press q
# Then in a new terminal or after quit:
pgrep mpv
```

- [ ] TUI exits cleanly back to the shell prompt
- [ ] `pgrep mpv` returns nothing (no orphan process)
- [ ] No error output printed to terminal

---

## 7. Clean quit via `Ctrl+C`

```bash
make run
# Press Ctrl+C
pgrep mpv
```

- [ ] TUI exits immediately
- [ ] `pgrep mpv` returns nothing
- [ ] No zombie process in `ps aux | grep mpv`

---

## 8. Socket cleanup

```bash
ls /tmp/mfp-mpv-*.sock 2>/dev/null || echo "clean"
```

- [ ] No leftover `.sock` files after quit

---

## Sign-off

| Item | Result | Notes |
|------|--------|-------|
| Build | ✅ / ❌ | |
| Tests 17/17 | ✅ / ❌ | |
| Audio plays | ✅ / ❌ | |
| Play/pause | ✅ / ❌ | |
| Seek ±30s | ✅ / ❌ | |
| Clean quit (`q`) | ✅ / ❌ | |
| Clean quit (`Ctrl+C`) | ✅ / ❌ | |
| No orphan mpv | ✅ / ❌ | |
| No leftover socket | ✅ / ❌ | |

**Phase 1 complete when all rows are ✅.**
