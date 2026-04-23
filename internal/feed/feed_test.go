package feed

import (
	"strings"
	"testing"
)

// ─── parseEpisodeTitle ────────────────────────────────────────────────────────

func TestParseEpisodeTitle(t *testing.T) {
	tests := []struct {
		input     string
		wantNum   int
		wantTitle string
		wantErr   bool
	}{
		{"Episode 78: Datassette", 78, "Datassette", false},
		{"Episode 1: Datassette", 1, "Datassette", false},
		{"Episode 55: 20 Jazz Funk Greats", 55, "20 Jazz Funk Greats", false},
		{"Episode 14: Tahlhoff Garten + Untitled", 14, "Tahlhoff Garten + Untitled", false},
		{"bad title", 0, "", true},
	}
	for _, tt := range tests {
		num, title, err := parseEpisodeTitle(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseEpisodeTitle(%q): expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseEpisodeTitle(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if num != tt.wantNum {
			t.Errorf("parseEpisodeTitle(%q) num = %d, want %d", tt.input, num, tt.wantNum)
		}
		if title != tt.wantTitle {
			t.Errorf("parseEpisodeTitle(%q) title = %q, want %q", tt.input, title, tt.wantTitle)
		}
	}
}

// ─── slugFromLink ─────────────────────────────────────────────────────────────

func TestSlugFromLink(t *testing.T) {
	tests := []struct {
		link string
		want string
	}{
		{"http://musicforprogramming.net/seventyeight", "seventyeight"},
		{"http://musicforprogramming.net/seventyeight/", "seventyeight"},
		{"http://musicforprogramming.net/one", "one"},
	}
	for _, tt := range tests {
		got := slugFromLink(tt.link)
		if got != tt.want {
			t.Errorf("slugFromLink(%q) = %q, want %q", tt.link, got, tt.want)
		}
	}
}

// ─── parseSingleTrack ────────────────────────────────────────────────────────

func TestParseSingleTrack(t *testing.T) {
	tests := []struct {
		line       string
		wantArtist string
		wantTitle  string
	}{
		{"David Borden - Enfield In Winter", "David Borden", "Enfield In Winter"},
		{"Datassette - rain_wind_canvas_20250601", "Datassette", "rain_wind_canvas_20250601"},
		{"zakè [ft. Slow Blink] - Caelum (No. 2)", "zakè [ft. Slow Blink]", "Caelum (No. 2)"},
		{"BBC Sound Effects - Trundling Wagon Interior", "BBC Sound Effects", "Trundling Wagon Interior"},
		// Title with dash should keep everything after the FIRST separator
		{"Artist - Title - With - Dashes", "Artist", "Title - With - Dashes"},
		// No separator → whole line becomes title
		{"NoSeparatorLine", "", "NoSeparatorLine"},
	}
	for _, tt := range tests {
		got := parseSingleTrack(tt.line)
		if got.Artist != tt.wantArtist {
			t.Errorf("parseSingleTrack(%q).Artist = %q, want %q", tt.line, got.Artist, tt.wantArtist)
		}
		if got.Title != tt.wantTitle {
			t.Errorf("parseSingleTrack(%q).Title = %q, want %q", tt.line, got.Title, tt.wantTitle)
		}
	}
}

// ─── parseTrackLines ─────────────────────────────────────────────────────────

func TestParseTrackLines(t *testing.T) {
	// Simulate the decoded tracklist string from Episode 78
	raw := "\n\t\t\tDavid Borden - Enfield In Winter<br>\n\t\t\tDatassette - rain_wind_canvas_20250601<br>\n\t\t\tzakè [ft. Slow Blink] - Caelum (No. 2)<br>\n\t\t"
	tracks := parseTrackLines(raw)
	if len(tracks) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(tracks))
	}
	if tracks[0].Artist != "David Borden" {
		t.Errorf("track 0 artist = %q, want %q", tracks[0].Artist, "David Borden")
	}
	if tracks[1].Title != "rain_wind_canvas_20250601" {
		t.Errorf("track 1 title = %q, want %q", tracks[1].Title, "rain_wind_canvas_20250601")
	}
	if tracks[2].Artist != "zakè [ft. Slow Blink]" {
		t.Errorf("track 2 artist = %q, want %q", tracks[2].Artist, "zakè [ft. Slow Blink]")
	}
}

// ─── parseTracklist ──────────────────────────────────────────────────────────

func TestParseTracklist_FromSapperHTML(t *testing.T) {
	// Minimal excerpt of the actual __SAPPER__ blob from Episode 78
	html := `__SAPPER__={preloaded:[void 0,{entry:{slug:"seventyeight",tracklist:"\n\t\t\tDavid Borden - Enfield In Winter\u003Cbr\u003E\n\t\t\tDatassette - rain_wind_canvas_20250601\u003Cbr\u003E\n\t\t\tElias. - Acquatic\u003Cbr\u003E\n\t\t"}}]};`

	tracks, err := parseTracklist(html)
	if err != nil {
		t.Fatalf("parseTracklist error: %v", err)
	}
	if len(tracks) != 3 {
		t.Fatalf("expected 3 tracks, got %d: %+v", len(tracks), tracks)
	}
	if tracks[0].Artist != "David Borden" || tracks[0].Title != "Enfield In Winter" {
		t.Errorf("track 0 = %+v, want {David Borden Enfield In Winter}", tracks[0])
	}
	if tracks[2].Artist != "Elias." {
		t.Errorf("track 2 artist = %q, want %q", tracks[2].Artist, "Elias.")
	}
}

// ─── parseDurationSeconds ─────────────────────────────────────────────────────

func TestParseDurationSeconds(t *testing.T) {
	tests := []struct {
		d    string
		want int
	}{
		{"1:30:00", 5400},
		{"2:00:00", 7200},
		{"2:14:02", 8042},
		{"30:00", 1800},
		{"1:00", 60},
	}
	for _, tt := range tests {
		got := parseDurationSeconds(tt.d)
		if got != tt.want {
			t.Errorf("parseDurationSeconds(%q) = %d, want %d", tt.d, got, tt.want)
		}
	}
}

// ─── ComputeStats ────────────────────────────────────────────────────────────

func TestComputeStats(t *testing.T) {
	eps := []Episode{
		{Duration: "1:30:00"}, // 5400s
		{Duration: "2:00:00"}, // 7200s
		{Duration: "30:00"},   // 1800s
	}
	stats := ComputeStats(eps)
	if stats.Episodes != 3 {
		t.Errorf("Episodes = %d, want 3", stats.Episodes)
	}
	if stats.TotalSeconds != 14400 {
		t.Errorf("TotalSeconds = %d, want 14400", stats.TotalSeconds)
	}
}

// ─── parseRSS ────────────────────────────────────────────────────────────────

func TestParseRSS_MinimalFeed(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rss xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" version="2.0">
  <channel>
    <item>
      <title>Episode 78: Datassette</title>
      <link>http://musicforprogramming.net/seventyeight</link>
      <pubDate>Mon, 20 Apr 2026 14:18:00 GMT</pubDate>
      <itunes:duration>1:30:00</itunes:duration>
      <enclosure url="https://datashat.net/music_for_programming_78-datassette.mp3" length="158925518" type="audio/mpeg"/>
    </item>
    <item>
      <title>Episode 77: Phonaut</title>
      <link>http://musicforprogramming.net/seventyseven</link>
      <pubDate>Wed, 14 Jan 2026 19:56:00 GMT</pubDate>
      <itunes:duration>2:00:00</itunes:duration>
      <enclosure url="https://datashat.net/music_for_programming_77-phonaut.mp3" length="221651679" type="audio/mpeg"/>
    </item>
  </channel>
</rss>`
	eps, err := parseRSS([]byte(xml))
	if err != nil {
		t.Fatalf("parseRSS error: %v", err)
	}
	if len(eps) != 2 {
		t.Fatalf("expected 2 episodes, got %d", len(eps))
	}
	ep := eps[0]
	if ep.Number != 78 {
		t.Errorf("Number = %d, want 78", ep.Number)
	}
	if ep.Title != "Datassette" {
		t.Errorf("Title = %q, want Datassette", ep.Title)
	}
	if ep.Slug != "seventyeight" {
		t.Errorf("Slug = %q, want seventyeight", ep.Slug)
	}
	if ep.Duration != "1:30:00" {
		t.Errorf("Duration = %q, want 1:30:00", ep.Duration)
	}
	if ep.Size != 158925518 {
		t.Errorf("Size = %d, want 158925518", ep.Size)
	}
	if !strings.Contains(ep.URL, "78-datassette.mp3") {
		t.Errorf("URL = %q, expected to contain 78-datassette.mp3", ep.URL)
	}
}

// ─── Integration test (requires network) ─────────────────────────────────────

func TestFetch_LiveRSS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live RSS fetch in short mode (-short)")
	}
	eps, err := Fetch()
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if len(eps) < 10 {
		t.Errorf("expected at least 10 episodes, got %d", len(eps))
	}
	// Spot-check episode 78
	for _, ep := range eps {
		if ep.Number == 78 {
			if ep.Title != "Datassette" {
				t.Errorf("ep 78 title = %q, want Datassette", ep.Title)
			}
			if ep.URL == "" {
				t.Error("ep 78 URL is empty")
			}
			return
		}
	}
	t.Error("episode 78 not found in live feed")
}

func TestFetchTracklist_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live tracklist fetch in short mode (-short)")
	}
	tracks, err := FetchTracklist("seventyeight")
	if err != nil {
		t.Fatalf("FetchTracklist error: %v", err)
	}
	if len(tracks) < 5 {
		t.Errorf("expected at least 5 tracks, got %d", len(tracks))
	}
	if tracks[0].Artist == "" {
		t.Error("first track has empty artist")
	}
}
