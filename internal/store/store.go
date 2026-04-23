// Package store persists user state to ~/.config/music-for-coding/.
//
// Currently stores:
//   - favourites: a set of episode numbers the user has starred
//   - positions:  last playback position per episode (for resume)
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ─── Paths ───────────────────────────────────────────────────────────────────

func configDir() (string, error) {
	// Prefer XDG_CONFIG_HOME if set, otherwise use ~/.config.
	// We don't use os.UserConfigDir() because on macOS it returns
	// ~/Library/Application Support which diverges from the XDG convention
	// used throughout this project's dotfiles.
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	dir := filepath.Join(base, "music-for-coding")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func favouritesPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "favourites.json"), nil
}

func positionsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "positions.json"), nil
}

func volumePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "volume.json"), nil
}

// ─── Favourites ──────────────────────────────────────────────────────────────

// LoadFavourites returns the set of starred episode numbers.
// Returns an empty map (not an error) if the file doesn't exist yet.
func LoadFavourites() (map[int]bool, error) {
	path, err := favouritesPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[int]bool), nil
	}
	if err != nil {
		return nil, err
	}
	var nums []int
	if err := json.Unmarshal(data, &nums); err != nil {
		return make(map[int]bool), nil // corrupt file → start fresh
	}
	favs := make(map[int]bool, len(nums))
	for _, n := range nums {
		favs[n] = true
	}
	return favs, nil
}

// SaveFavourites persists the set of starred episode numbers.
func SaveFavourites(favs map[int]bool) error {
	path, err := favouritesPath()
	if err != nil {
		return err
	}
	nums := make([]int, 0, len(favs))
	for n := range favs {
		nums = append(nums, n)
	}
	data, err := json.MarshalIndent(nums, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ─── Resume positions ─────────────────────────────────────────────────────────

// Positions maps episode number → last known position in seconds.
type Positions map[int]float64

// LoadPositions returns saved playback positions, or an empty map.
func LoadPositions() (Positions, error) {
	path, err := positionsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(Positions), nil
	}
	if err != nil {
		return nil, err
	}
	var pos Positions
	if err := json.Unmarshal(data, &pos); err != nil {
		return make(Positions), nil
	}
	return pos, nil
}

// SavePosition upserts the playback position for one episode.
func SavePosition(episodeNum int, seconds float64) error {
	pos, err := LoadPositions()
	if err != nil {
		pos = make(Positions)
	}
	if seconds < 5 {
		delete(pos, episodeNum) // don't bother saving near the start
		return savePositions(pos)
	}
	pos[episodeNum] = seconds
	return savePositions(pos)
}

func savePositions(pos Positions) error {
	path, err := positionsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(pos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ─── Volume ───────────────────────────────────────────────────────────────────

const DefaultVolume = 100

// LoadVolume returns the saved volume (0-150), or DefaultVolume if not set.
func LoadVolume() (int, error) {
	path, err := volumePath()
	if err != nil {
		return DefaultVolume, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultVolume, nil
	}
	if err != nil {
		return DefaultVolume, err
	}
	var vol int
	if err := json.Unmarshal(data, &vol); err != nil {
		return DefaultVolume, nil
	}
	if vol < 0 || vol > 150 {
		return DefaultVolume, nil
	}
	return vol, nil
}

// SaveVolume persists the current volume level.
func SaveVolume(vol int) error {
	path, err := volumePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(vol)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

