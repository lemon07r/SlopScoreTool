// Command datagen refreshes the analyzer's embedded data from the pinned
// upstream EQ-Bench slop-score revision. It fetches the source lists and
// leaderboard, validates their shape, derives the compact calibration
// artifact, and records provenance in a manifest. Runtime analysis never
// needs the network; this tool is for explicit data refreshes only.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	upstreamRepo   = "https://github.com/sam-paech/slop-score"
	upstreamCommit = "289264ab2a0df1358ef81ced3c259610ea709514"
	rangeBuffer    = 0.1
)

var listFiles = []string{
	"slop_list.json",
	"slop_list_bigrams.json",
	"slop_list_trigrams.json",
}

const leaderboardFile = "leaderboard_results.json"

type normRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type calibration struct {
	Version int                  `json:"version"`
	Ranges  map[string]normRange `json:"ranges"`
}

type fileRecord struct {
	SHA256  string `json:"sha256"`
	Bytes   int    `json:"bytes"`
	Entries int    `json:"entries,omitempty"`
}

type manifest struct {
	Version int `json:"version"`
	Source  struct {
		Repo    string `json:"repo"`
		Commit  string `json:"commit"`
		License string `json:"license"`
	} `json:"source"`
	Files       map[string]fileRecord `json:"files"`
	Leaderboard struct {
		SHA256                string `json:"sha256"`
		Bytes                 int    `json:"bytes"`
		Results               int    `json:"results"`
		IncludesHumanBaseline bool   `json:"includes_human_baseline"`
	} `json:"leaderboard"`
	Derived struct {
		RangeBuffer float64 `json:"range_buffer"`
		RangesFrom  string  `json:"ranges_from"`
	} `json:"derived"`
}

func sourceURL(name string) string {
	return upstreamRepo + "/raw/" + upstreamCommit + "/data/" + name
}

func fetch(client *http.Client, name string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, sourceURL(name), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", name, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: HTTP %d", name, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	return body, nil
}

// listEntries validates the [["phrase"], ...] shape shared by all slop lists.
func listEntries(body []byte) (int, error) {
	var entries [][]string
	dec := json.NewDecoder(bytes.NewReader(body))
	if err := dec.Decode(&entries); err != nil {
		return 0, fmt.Errorf("parse list: %w", err)
	}
	if len(entries) == 0 {
		return 0, errors.New("list is empty")
	}
	for i, e := range entries {
		if len(e) == 0 || e[0] == "" {
			return 0, fmt.Errorf("entry %d has no phrase", i)
		}
	}
	return len(entries), nil
}

type leaderboardMetrics struct {
	SlopWords    float64 `json:"slop_list_matches_per_1k_words"`
	SlopTrigrams float64 `json:"slop_trigram_matches_per_1k_words"`
	Contrast     float64 `json:"not_x_but_y_per_1k_chars"`
}

type leaderboard struct {
	Results []struct {
		Model   string             `json:"model"`
		Metrics leaderboardMetrics `json:"metrics"`
	} `json:"results"`
}

func parseLeaderboard(body []byte) (*leaderboard, error) {
	var lb leaderboard
	if err := json.Unmarshal(body, &lb); err != nil {
		return nil, fmt.Errorf("parse leaderboard: %w", err)
	}
	if len(lb.Results) < 2 {
		return nil, fmt.Errorf("leaderboard has %d results, want at least 2", len(lb.Results))
	}
	return &lb, nil
}

// paddedRange mirrors the upstream normalization: observed min/max expanded
// by rangeBuffer of the span on each side.
func paddedRange(values []float64) normRange {
	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	span := max - min
	return normRange{Min: min - span*rangeBuffer, Max: max + span*rangeBuffer}
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func writeJSON(dir, name string, v any) error {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	buf = append(buf, '\n')
	return os.WriteFile(filepath.Join(dir, name), buf, 0o600)
}

func run(outDir string) error {
	client := &http.Client{Timeout: 60 * time.Second}

	m := manifest{Version: 1, Files: map[string]fileRecord{}}
	m.Source.Repo = upstreamRepo
	m.Source.Commit = upstreamCommit
	m.Source.License = "MIT"
	m.Derived.RangeBuffer = rangeBuffer
	m.Derived.RangesFrom = leaderboardFile

	for _, name := range listFiles {
		body, err := fetch(client, name)
		if err != nil {
			return err
		}
		entries, err := listEntries(body)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(outDir, name), body, 0o600); err != nil {
			return err
		}
		m.Files[name] = fileRecord{SHA256: sha256Hex(body), Bytes: len(body), Entries: entries}
	}

	body, err := fetch(client, leaderboardFile)
	if err != nil {
		return err
	}
	lb, err := parseLeaderboard(body)
	if err != nil {
		return err
	}

	var words, trigrams, contrast []float64
	human := false
	for _, r := range lb.Results {
		words = append(words, r.Metrics.SlopWords)
		trigrams = append(trigrams, r.Metrics.SlopTrigrams)
		contrast = append(contrast, r.Metrics.Contrast)
		if r.Model == "human-baseline" {
			human = true
		}
	}
	m.Leaderboard.SHA256 = sha256Hex(body)
	m.Leaderboard.Bytes = len(body)
	m.Leaderboard.Results = len(lb.Results)
	m.Leaderboard.IncludesHumanBaseline = human

	cal := calibration{
		Version: 1,
		Ranges: map[string]normRange{
			"slop_words":        paddedRange(words),
			"slop_trigrams":     paddedRange(trigrams),
			"contrast_patterns": paddedRange(contrast),
		},
	}

	if err := writeJSON(outDir, "calibration.json", cal); err != nil {
		return err
	}
	if err := writeJSON(outDir, "manifest.json", m); err != nil {
		return err
	}

	fmt.Printf("datagen: %d lists, leaderboard %d results (human baseline: %t)\n",
		len(m.Files), m.Leaderboard.Results, human)
	for k, r := range cal.Ranges {
		fmt.Printf("  %-18s [%.4f, %.4f]\n", k, r.Min, r.Max)
	}
	return nil
}

func main() {
	outDir := "slop/data"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		fmt.Fprintln(os.Stderr, "datagen:", err)
		os.Exit(1)
	}
	if err := run(outDir); err != nil {
		fmt.Fprintln(os.Stderr, "datagen:", err)
		os.Exit(1)
	}
}
