// Package feed fetches and parses the Music For Programming episode catalogue.
//
// Episode list: fetched from the RSS feed (cached to disk).
// Tracklists:   fetched lazily from individual episode HTML pages
//               (embedded in the __SAPPER__ JSON blob).
package feed

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ─── Constants ───────────────────────────────────────────────────────────────

const (
	rssURL      = "https://musicforprogramming.net/rss.xml"
	baseURL     = "https://musicforprogramming.net"
	cacheTTL    = 1 * time.Hour
	httpTimeout = 15 * time.Second
)

// ─── Data types ──────────────────────────────────────────────────────────────

// Track is one entry in an episode's tracklist.
type Track struct {
	Artist string
	Title  string
}

// Episode is a single MFP episode as parsed from the RSS feed.
// Tracks is populated lazily via FetchTracklist.
type Episode struct {
	Number   int
	Slug     string  // e.g. "seventyeight" — derived from the <link> field
	Title    string  // e.g. "Datassette"
	URL      string  // mp3 URL
	Duration string  // e.g. "1:30:00"
	Size     int64   // bytes
	PubDate  time.Time
	Tracks   []Track // nil until FetchTracklist is called
}

// Stats is aggregate data computed from the full episode list.
type Stats struct {
	Episodes     int
	TotalSeconds int
}

// ─── RSS XML structs ─────────────────────────────────────────────────────────

type rssRoot struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title     string     `xml:"title"`
	Link      string     `xml:"link"`
	PubDate   string     `xml:"pubDate"`
	Duration  string     `xml:"duration"`
	Enclosure rssEncl    `xml:"enclosure"`
}

type rssEncl struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
}

// ─── Fetch ───────────────────────────────────────────────────────────────────

// Fetch returns the full episode list, using a disk cache when fresh.
// Episodes are returned newest-first (same order as the RSS feed).
func Fetch() ([]Episode, error) {
	raw, err := fetchRSS()
	if err != nil {
		return nil, err
	}
	return parseRSS(raw)
}

// fetchRSS returns raw RSS XML, preferring a fresh disk cache.
func fetchRSS() ([]byte, error) {
	cachePath, err := rssCachePath()
	if err != nil {
		return nil, err
	}

	// Use cache if it exists and is younger than cacheTTL.
	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < cacheTTL {
			if data, err := os.ReadFile(cachePath); err == nil {
				return data, nil
			}
		}
	}

	// Fetch from network.
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(rssURL)
	if err != nil {
		// Fall back to stale cache on network error.
		if data, readErr := os.ReadFile(cachePath); readErr == nil {
			return data, nil
		}
		return nil, fmt.Errorf("fetching RSS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if data, readErr := os.ReadFile(cachePath); readErr == nil {
			return data, nil
		}
		return nil, fmt.Errorf("RSS HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading RSS body: %w", err)
	}

	// Persist to cache (best-effort).
	_ = os.MkdirAll(filepath.Dir(cachePath), 0o755)
	_ = os.WriteFile(cachePath, data, 0o644)

	return data, nil
}

// rssCachePath returns the platform cache directory for the RSS file.
func rssCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("finding cache dir: %w", err)
	}
	return filepath.Join(cacheDir, "music-for-coding", "rss.xml"), nil
}

// ─── Parse RSS ───────────────────────────────────────────────────────────────

// parseRSS converts raw RSS XML into a slice of Episodes.
func parseRSS(data []byte) ([]Episode, error) {
	var root rssRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parsing RSS XML: %w", err)
	}

	episodes := make([]Episode, 0, len(root.Channel.Items))
	for _, item := range root.Channel.Items {
		ep, err := itemToEpisode(item)
		if err != nil {
			continue // skip malformed items
		}
		episodes = append(episodes, ep)
	}
	return episodes, nil
}

// itemToEpisode converts a raw RSS item to an Episode.
func itemToEpisode(item rssItem) (Episode, error) {
	// Title is "Episode 78: Datassette" → number=78, title="Datassette"
	num, artistTitle, err := parseEpisodeTitle(item.Title)
	if err != nil {
		return Episode{}, err
	}

	// Slug from link: "http://musicforprogramming.net/seventyeight" → "seventyeight"
	slug := slugFromLink(item.Link)

	pubDate, _ := time.Parse(time.RFC1123, item.PubDate)

	return Episode{
		Number:   num,
		Slug:     slug,
		Title:    artistTitle,
		URL:      item.Enclosure.URL,
		Duration: item.Duration,
		Size:     item.Enclosure.Length,
		PubDate:  pubDate,
	}, nil
}

// parseEpisodeTitle splits "Episode 78: Datassette" into (78, "Datassette").
func parseEpisodeTitle(raw string) (int, string, error) {
	// Format: "Episode NN: Artist"
	raw = strings.TrimPrefix(raw, "Episode ")
	idx := strings.Index(raw, ": ")
	if idx < 0 {
		return 0, "", fmt.Errorf("unexpected title format: %q", raw)
	}
	num, err := strconv.Atoi(strings.TrimSpace(raw[:idx]))
	if err != nil {
		return 0, "", fmt.Errorf("parsing episode number from %q: %w", raw, err)
	}
	return num, strings.TrimSpace(raw[idx+2:]), nil
}

// slugFromLink extracts the episode slug from a full URL.
func slugFromLink(link string) string {
	parts := strings.Split(strings.TrimRight(link, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// ─── Tracklist ───────────────────────────────────────────────────────────────

// tracklistRe matches the tracklist value inside the Sapper JS blob.
// Sapper uses unquoted object keys: tracklist:"..."
var tracklistRe = regexp.MustCompile(`tracklist:"((?:[^"\\]|\\.)*)"`)

// FetchTracklist fetches an episode's HTML page and extracts the tracklist
// from the embedded __SAPPER__ JSON blob.
func FetchTracklist(slug string) ([]Track, error) {
	url := baseURL + "/" + slug + "/"
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching episode page %s: %w", slug, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading episode page: %w", err)
	}

	return parseTracklist(string(body))
}

// parseTracklist extracts and parses the tracklist from raw episode HTML.
func parseTracklist(html string) ([]Track, error) {
	matches := tracklistRe.FindStringSubmatch(html)
	if len(matches) < 2 {
		return nil, fmt.Errorf("tracklist not found in page HTML")
	}

	// Decode the JSON string escape sequences by wrapping in quotes
	// and using json.Unmarshal via a synthetic struct.
	raw := unescapeJSON(`"` + matches[1] + `"`)

	return parseTrackLines(raw), nil
}

// unescapeJSON decodes a JSON string literal (including \uXXXX sequences).
func unescapeJSON(jsonStr string) string {
	var s string
	// encoding/json will handle \n, \t, \uXXXX etc.
	if err := unmarshalString([]byte(jsonStr), &s); err != nil {
		// Fallback: return as-is
		return jsonStr
	}
	return s
}

// parseTrackLines splits an HTML tracklist string into Track values.
// Lines are separated by <br> tags; each line is "Artist - Title".
func parseTrackLines(raw string) []Track {
	// Normalize: replace <br> variants with newline
	raw = brRe.ReplaceAllString(raw, "\n")
	// Strip remaining HTML tags
	raw = htmlTagRe.ReplaceAllString(raw, "")

	var tracks []Track
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		track := parseSingleTrack(line)
		tracks = append(tracks, track)
	}
	return tracks
}

var (
	brRe      = regexp.MustCompile(`(?i)<br\s*/?>`)
	htmlTagRe = regexp.MustCompile(`<[^>]+>`)
	// Split on " - " but prefer the first occurrence to handle titles with dashes.
	dashRe = regexp.MustCompile(`\s+[-\x{2013}\x{2014}]\s+`)
)

// parseSingleTrack splits "Artist - Title" into a Track.
// Uses the first dash-surrounded separator to handle titles containing dashes.
func parseSingleTrack(line string) Track {
	loc := dashRe.FindStringIndex(line)
	if loc == nil {
		return Track{Title: line}
	}
	return Track{
		Artist: strings.TrimSpace(line[:loc[0]]),
		Title:  strings.TrimSpace(line[loc[1]:]),
	}
}

// ─── Stats ───────────────────────────────────────────────────────────────────

// ComputeStats returns aggregate stats from a slice of episodes.
func ComputeStats(episodes []Episode) Stats {
	total := 0
	for _, ep := range episodes {
		total += parseDurationSeconds(ep.Duration)
	}
	return Stats{
		Episodes:     len(episodes),
		TotalSeconds: total,
	}
}

// parseDurationSeconds converts "h:mm:ss" or "mm:ss" to total seconds.
func parseDurationSeconds(d string) int {
	parts := strings.Split(d, ":")
	var h, m, s int
	switch len(parts) {
	case 3:
		h, _ = strconv.Atoi(parts[0])
		m, _ = strconv.Atoi(parts[1])
		s, _ = strconv.Atoi(parts[2])
	case 2:
		m, _ = strconv.Atoi(parts[0])
		s, _ = strconv.Atoi(parts[1])
	}
	return h*3600 + m*60 + s
}
