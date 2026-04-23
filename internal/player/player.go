// Package player drives an mpv subprocess over its JSON IPC socket.
//
// Usage:
//
//	p, err := player.New()
//	defer p.Close()
//	p.Load("https://example.com/file.mp3")
//	state, _ := p.GetState()
//	p.TogglePause()
//	p.Seek(-30)
package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// ─── Constants ───────────────────────────────────────────────────────────────

const (
	socketRetries    = 60               // 60 × 100 ms = 6 s max wait
	socketRetryDelay = 100 * time.Millisecond
	commandTimeout   = 5 * time.Second
)

// ─── Types ───────────────────────────────────────────────────────────────────

// State is a snapshot of current playback.
type State struct {
	Loaded   bool
	Paused   bool
	Position float64 // seconds
	Duration float64 // seconds
}

// ipcRequest is the JSON shape mpv expects.
type ipcRequest struct {
	Command   []interface{} `json:"command"`
	RequestID uint64        `json:"request_id"`
}

// ipcResponse covers both command responses and async events from mpv.
type ipcResponse struct {
	Error     string          `json:"error"`
	Data      json.RawMessage `json:"data"`
	RequestID uint64          `json:"request_id"`
	Event     string          `json:"event"`
}

// ─── Player ──────────────────────────────────────────────────────────────────

// Player owns an mpv process and the IPC connection to it.
type Player struct {
	sockPath string
	cmd      *exec.Cmd
	conn     net.Conn

	mu       sync.Mutex
	reqID    uint64
	pending  map[uint64]chan ipcResponse

	readDone chan struct{}
}

// New spawns mpv in headless idle mode and opens its IPC socket.
// The caller must call Close() when done.
func New() (*Player, error) {
	sockPath := filepath.Join(os.TempDir(), fmt.Sprintf("mfp-mpv-%d.sock", os.Getpid()))
	_ = os.Remove(sockPath) // clean up any leftover from a previous crash

	cmd := exec.Command("mpv",
		"--idle=yes",
		"--no-video",
		"--input-ipc-server="+sockPath,
		"--really-quiet",
	)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting mpv: %w", err)
	}

	// mpv creates the socket asynchronously — poll until it appears.
	var conn net.Conn
	var connErr error
	for i := 0; i < socketRetries; i++ {
		time.Sleep(socketRetryDelay)
		conn, connErr = net.Dial("unix", sockPath)
		if connErr == nil {
			break
		}
	}
	if connErr != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("connecting to mpv IPC socket after %d attempts: %w",
			socketRetries, connErr)
	}

	p := &Player{
		sockPath: sockPath,
		cmd:      cmd,
		conn:     conn,
		pending:  make(map[uint64]chan ipcResponse),
		readDone: make(chan struct{}),
	}
	go p.readLoop()
	return p, nil
}

// ─── IPC internals ───────────────────────────────────────────────────────────

// readLoop reads newline-delimited JSON from the socket and dispatches each
// response to the goroutine that issued the matching request.
func (p *Player) readLoop() {
	defer close(p.readDone)
	scanner := bufio.NewScanner(p.conn)
	for scanner.Scan() {
		var resp ipcResponse
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			continue
		}
		// Async events (property changes, end-file, etc.) are ignored for now.
		// Phase 2 will use observe_property to react to them.
		if resp.Event != "" {
			continue
		}
		p.mu.Lock()
		ch, ok := p.pending[resp.RequestID]
		if ok {
			delete(p.pending, resp.RequestID)
		}
		p.mu.Unlock()
		if ok {
			ch <- resp
		}
	}
}

// send serialises a command, writes it to the socket, and waits for the
// matching response (matched by request_id).
func (p *Player) send(args ...interface{}) (json.RawMessage, error) {
	p.mu.Lock()
	p.reqID++
	id := p.reqID
	ch := make(chan ipcResponse, 1)
	p.pending[id] = ch

	data, err := json.Marshal(ipcRequest{Command: args, RequestID: id})
	if err != nil {
		delete(p.pending, id)
		p.mu.Unlock()
		return nil, fmt.Errorf("marshaling command: %w", err)
	}
	data = append(data, '\n')
	_, writeErr := p.conn.Write(data)
	if writeErr != nil {
		delete(p.pending, id)
	}
	p.mu.Unlock()

	if writeErr != nil {
		return nil, fmt.Errorf("writing to mpv socket: %w", writeErr)
	}

	select {
	case resp := <-ch:
		if resp.Error != "" && resp.Error != "success" {
			return nil, fmt.Errorf("mpv: %s", resp.Error)
		}
		return resp.Data, nil
	case <-time.After(commandTimeout):
		p.mu.Lock()
		delete(p.pending, id)
		p.mu.Unlock()
		return nil, fmt.Errorf("mpv command timed out after %s", commandTimeout)
	}
}

// ─── Public API ──────────────────────────────────────────────────────────────

// Load starts playback of url immediately, replacing any current track.
func (p *Player) Load(url string) error {
	_, err := p.send("loadfile", url, "replace")
	return err
}

// TogglePause flips between playing and paused.
func (p *Player) TogglePause() error {
	_, err := p.send("cycle", "pause")
	return err
}

// Seek moves the playback position by delta seconds.
// Positive values skip forward; negative values rewind.
func (p *Player) Seek(delta float64) error {
	_, err := p.send("seek", delta, "relative")
	return err
}

// GetState polls mpv for the current position, duration, and pause flag.
// Returns a zero State (Loaded=false) if nothing has been loaded yet.
func (p *Player) GetState() (State, error) {
	posData, err := p.send("get_property", "time-pos")
	if err != nil {
		// Property not available = nothing loaded yet
		return State{}, nil
	}
	durData, err := p.send("get_property", "duration")
	if err != nil {
		return State{}, nil
	}
	pauseData, err := p.send("get_property", "pause")
	if err != nil {
		return State{}, nil
	}

	var pos, dur float64
	var paused bool
	_ = json.Unmarshal(posData, &pos)
	_ = json.Unmarshal(durData, &dur)
	_ = json.Unmarshal(pauseData, &paused)

	return State{
		Loaded:   true,
		Paused:   paused,
		Position: pos,
		Duration: dur,
	}, nil
}

// SeekAbsolute seeks to an absolute position in seconds.
func (p *Player) SeekAbsolute(secs float64) error {
	_, err := p.send("seek", secs, "absolute")
	return err
}

// SetVolume sets playback volume (0–150).
func (p *Player) SetVolume(vol int) error {
	_, err := p.send("set_property", "volume", vol)
	return err
}

// Close shuts down mpv gracefully and removes the socket file.
// It is safe to call Close multiple times.
func (p *Player) Close() error {
	_ = p.conn.Close() // causes readLoop to exit via scanner.Scan()
	<-p.readDone       // wait for clean shutdown

	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
	}
	_ = os.Remove(p.sockPath)
	return nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// FormatDuration formats a float64 number of seconds as mm:ss or h:mm:ss.
// Negative values are clamped to zero.
func FormatDuration(secs float64) string {
	if secs < 0 {
		secs = 0
	}
	total := int(secs)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
