// Command eval runs the analyzer over the behavior corpus (testdata/corpus)
// and the regression fixtures (testdata/fixtures), aggregating scores and
// diagnostic firing rates per label. It is the evidence base for keeping,
// cutting, or adjusting the non-scored diagnostics, and its output feeds
// docs/methodology.md. Run via `make eval`.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slop-score/slop"
	"sort"
	"strings"
)

var labels = map[string]string{
	"human-essay":        "human-like",
	"human-review":       "human-like",
	"model-like-fiction": "model-like",
	"model-like-essay":   "model-like",
	"technical-docs":     "technical",
	"dialogue-heavy":     "dialogue-heavy",
	"rhetorical":         "rhetorical",
	"short-note":         "short",
	"pair-before":        "revision-before",
	"pair-after":         "revision-after",
	// Regression fixtures inherited from the original Go seed: provenance
	// known, human/AI status unverified, so they form their own group
	// rather than a labeled claim.
	"sample1": "legacy-fixture",
	"sample2": "legacy-fixture",
	"sample3": "legacy-fixture",
	"sample4": "legacy-fixture",
	"sample5": "legacy-fixture",
}

type docResult struct {
	Name       string         `json:"name"`
	Label      string         `json:"label"`
	Score      *float64       `json:"score"`
	Words      int            `json:"words"`
	SentenceCV float64        `json:"sentence_length_cv"`
	Kinds      map[string]int `json:"findings_by_kind"`
}

type labelAgg struct {
	Label     string         `json:"label"`
	Docs      int            `json:"docs"`
	MeanScore float64        `json:"mean_score"`
	Scored    int            `json:"scored"`
	MeanCV    float64        `json:"mean_sentence_length_cv"`
	Kinds     map[string]int `json:"total_findings_by_kind"`
}

func main() {
	jsonOut := flag.Bool("json", false, "emit JSON instead of a table")
	flag.Parse()

	a, err := slop.NewAnalyzer()
	if err != nil {
		fmt.Fprintln(os.Stderr, "eval:", err)
		os.Exit(1)
	}

	var results []docResult
	for _, dir := range []string{"testdata/corpus", "testdata/fixtures"} {
		files, err := filepath.Glob(filepath.Join(dir, "*.txt"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "eval:", err)
			os.Exit(1)
		}
		sort.Strings(files)
		for _, f := range files {
			name := strings.TrimSuffix(filepath.Base(f), ".txt")
			label, ok := labels[name]
			if !ok {
				continue
			}
			data, err := os.ReadFile(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, "eval:", err)
				os.Exit(1)
			}
			doc, err := a.Analyze(name, string(data), slop.Options{Full: true})
			if err != nil {
				fmt.Fprintln(os.Stderr, "eval:", err)
				os.Exit(1)
			}
			r := docResult{Name: name, Label: label, Words: doc.Stats.Words, Kinds: map[string]int{}}
			if doc.Score != nil {
				v := doc.Score.Value
				r.Score = &v
			}
			for _, s := range doc.Signals {
				if s.Name == "sentence_length_cv" {
					r.SentenceCV = s.Value
				}
			}
			for _, fd := range doc.Findings {
				r.Kinds[fd.Kind]++
			}
			results = append(results, r)
		}
	}

	aggs := map[string]*labelAgg{}
	var order []string
	for _, r := range results {
		agg, ok := aggs[r.Label]
		if !ok {
			agg = &labelAgg{Label: r.Label, Kinds: map[string]int{}}
			aggs[r.Label] = agg
			order = append(order, r.Label)
		}
		agg.Docs++
		if r.Score != nil {
			agg.MeanScore += *r.Score
			agg.Scored++
		}
		agg.MeanCV += r.SentenceCV
		for k, n := range r.Kinds {
			agg.Kinds[k] += n
		}
	}
	aggsList := make([]*labelAgg, 0, len(aggs))
	for _, l := range order {
		agg := aggs[l]
		if agg.Scored > 0 {
			agg.MeanScore /= float64(agg.Scored)
		}
		agg.MeanCV /= float64(agg.Docs)
		aggsList = append(aggsList, agg)
	}

	if *jsonOut {
		out := map[string]any{"documents": results, "by_label": aggsList}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintln(os.Stderr, "eval:", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("%-18s %-16s %7s %6s %5s | %s\n", "document", "label", "score", "cv", "words", "findings by kind")
	fmt.Println(strings.Repeat("-", 100))
	for _, r := range results {
		score := "  -"
		if r.Score != nil {
			score = fmt.Sprintf("%6.1f", *r.Score)
		}
		fmt.Printf("%-18s %-16s %7s %6.2f %5d | %s\n", r.Name, r.Label, score, r.SentenceCV, r.Words, kindsSummary(r.Kinds))
	}
	fmt.Println()
	fmt.Printf("%-16s %5s %7s %6s | %s\n", "label", "docs", "score", "cv", "total findings by kind")
	fmt.Println(strings.Repeat("-", 100))
	for _, agg := range aggsList {
		fmt.Printf("%-16s %5d %7.1f %6.2f | %s\n", agg.Label, agg.Docs, agg.MeanScore, agg.MeanCV, kindsSummary(agg.Kinds))
	}
}

func kindsSummary(kinds map[string]int) string {
	if len(kinds) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(kinds))
	for k := range kinds {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s:%d", k, kinds[k])
	}
	return strings.Join(parts, " ")
}
