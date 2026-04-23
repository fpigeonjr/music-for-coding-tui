package store

import (
	"os"
	"testing"
)

func TestFavourites_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	favs := map[int]bool{78: true, 42: true, 1: true}
	if err := SaveFavourites(favs); err != nil {
		t.Fatalf("SaveFavourites: %v", err)
	}
	loaded, err := LoadFavourites()
	if err != nil {
		t.Fatalf("LoadFavourites: %v", err)
	}
	for _, n := range []int{78, 42, 1} {
		if !loaded[n] {
			t.Errorf("expected episode %d to be a favourite", n)
		}
	}
}

func TestFavourites_EmptyOnMissingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	favs, err := LoadFavourites()
	if err != nil {
		t.Fatalf("LoadFavourites on missing file: %v", err)
	}
	if len(favs) != 0 {
		t.Errorf("expected empty map, got %v", favs)
	}
}

func TestFavourites_CorruptFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir, _ := configDir()
	_ = os.MkdirAll(dir, 0o755)
	path, _ := favouritesPath()
	_ = os.WriteFile(path, []byte("not json {{{{"), 0o644)

	favs, err := LoadFavourites()
	if err != nil {
		t.Fatalf("expected graceful degradation on corrupt file, got: %v", err)
	}
	if len(favs) != 0 {
		t.Errorf("expected empty map on corrupt file, got %v", favs)
	}
}

func TestPositions_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := SavePosition(78, 273.5); err != nil {
		t.Fatalf("SavePosition: %v", err)
	}
	pos, err := LoadPositions()
	if err != nil {
		t.Fatalf("LoadPositions: %v", err)
	}
	if pos[78] != 273.5 {
		t.Errorf("pos[78] = %v, want 273.5", pos[78])
	}
}

func TestPositions_SkipsNearStart(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := SavePosition(78, 3.0); err != nil {
		t.Fatalf("SavePosition: %v", err)
	}
	pos, err := LoadPositions()
	if err != nil {
		t.Fatalf("LoadPositions: %v", err)
	}
	if _, ok := pos[78]; ok {
		t.Error("expected position < 5s to not be persisted")
	}
}

func TestVolume_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := SaveVolume(80); err != nil {
		t.Fatalf("SaveVolume: %v", err)
	}
	vol, err := LoadVolume()
	if err != nil {
		t.Fatalf("LoadVolume: %v", err)
	}
	if vol != 80 {
		t.Errorf("vol = %d, want 80", vol)
	}
}

func TestVolume_DefaultWhenMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	vol, err := LoadVolume()
	if err != nil {
		t.Fatalf("LoadVolume on missing file: %v", err)
	}
	if vol != DefaultVolume {
		t.Errorf("vol = %d, want %d (default)", vol, DefaultVolume)
	}
}

func TestVolume_OutOfRangeFallsBack(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir, _ := configDir()
	_ = os.MkdirAll(dir, 0o755)
	path, _ := volumePath()
	_ = os.WriteFile(path, []byte("999"), 0o644)

	vol, _ := LoadVolume()
	if vol != DefaultVolume {
		t.Errorf("out-of-range vol = %d, want default %d", vol, DefaultVolume)
	}
}
