package slop

import "sort"

// Comparison pairs two document analyses with per-signal and per-finding
// deltas. It never declares a winner: a lower score is not a better revision.
type Comparison struct {
	SchemaVersion int              `json:"schema_version"`
	Analyzer      AnalyzerInfo     `json:"analyzer"`
	Method        MethodInfo       `json:"method"`
	Before        Document         `json:"before"`
	After         Document         `json:"after"`
	Deltas        ComparisonDeltas `json:"deltas"`
}

// ComparisonDeltas summarizes what changed between two drafts.
type ComparisonDeltas struct {
	Score      *ScoreDelta      `json:"score,omitempty"`
	Components []ComponentDelta `json:"components,omitempty"`
	Findings   []FindingDelta   `json:"findings"`
	Truncation *DeltaTruncation `json:"truncation,omitempty"`
}

// ScoreDelta is the composite change, labeled by basis.
type ScoreDelta struct {
	Basis  string  `json:"basis"`
	Before float64 `json:"before"`
	After  float64 `json:"after"`
	Delta  float64 `json:"delta"`
}

// ComponentDelta is the raw-rate change of one scored component.
type ComponentDelta struct {
	Name   string  `json:"name"`
	Unit   string  `json:"unit"`
	Before float64 `json:"before"`
	After  float64 `json:"after"`
	Delta  float64 `json:"delta"`
}

// FindingDelta is the count change of one finding present in either draft.
type FindingDelta struct {
	Kind   string `json:"kind"`
	Text   string `json:"text"`
	Before int    `json:"before"`
	After  int    `json:"after"`
	Status string `json:"status"` // "added", "removed", "increased", "decreased", "unchanged"
}

// DeltaTruncation records omission from the findings delta list.
type DeltaTruncation struct {
	Shown int `json:"shown"`
	Total int `json:"total"`
}

// Compare builds a Comparison from two analyzed documents sharing the same
// analyzer method metadata.
func (a *Analyzer) Compare(before, after Document, maxFindings int) Comparison {
	cmp := Comparison{
		SchemaVersion: SchemaVersion,
		Analyzer:      AnalyzerInfo{Name: AnalyzerName, Version: Version},
		Method:        a.methodInfo(),
		Before:        before,
		After:         after,
	}

	if before.Score != nil && after.Score != nil {
		cmp.Deltas.Score = &ScoreDelta{
			Basis:  before.Score.Basis,
			Before: before.Score.Value,
			After:  after.Score.Value,
			Delta:  after.Score.Value - before.Score.Value,
		}
		beforeRates := map[string]Component{}
		for _, c := range before.Score.Components {
			beforeRates[c.Name] = c
		}
		for _, c := range after.Score.Components {
			b := beforeRates[c.Name]
			cmp.Deltas.Components = append(cmp.Deltas.Components, ComponentDelta{
				Name: c.Name, Unit: c.Unit, Before: b.Rate, After: c.Rate, Delta: c.Rate - b.Rate,
			})
		}
	}

	type key struct{ kind, text string }
	counts := map[key][2]int{}
	var kinds []string
	seen := map[string]bool{}
	markKind := func(k string) {
		if !seen[k] {
			seen[k] = true
			kinds = append(kinds, k)
		}
	}
	for _, f := range before.Findings {
		counts[key{f.Kind, f.Text}] = [2]int{f.Count, 0}
		markKind(f.Kind)
	}
	for _, f := range after.Findings {
		k := key{f.Kind, f.Text}
		pair := counts[k]
		pair[1] = f.Count
		counts[k] = pair
		markKind(f.Kind)
	}

	var deltas []FindingDelta
	for k, pair := range counts {
		status := "unchanged"
		switch {
		case pair[0] == 0 && pair[1] > 0:
			status = "added"
		case pair[0] > 0 && pair[1] == 0:
			status = "removed"
		case pair[1] > pair[0]:
			status = "increased"
		case pair[1] < pair[0]:
			status = "decreased"
		}
		deltas = append(deltas, FindingDelta{
			Kind: k.kind, Text: k.text, Before: pair[0], After: pair[1], Status: status,
		})
	}
	sort.Slice(deltas, func(i, j int) bool {
		a, b := deltas[i], deltas[j]
		if kindOrder[a.Kind] != kindOrder[b.Kind] {
			return kindOrder[a.Kind] < kindOrder[b.Kind]
		}
		da, db := abs(a.After-a.Before), abs(b.After-b.Before)
		if da != db {
			return da > db
		}
		return a.Text < b.Text
	})

	total := len(deltas)
	if maxFindings > 0 && total > maxFindings {
		deltas = deltas[:maxFindings]
		cmp.Deltas.Truncation = &DeltaTruncation{Shown: len(deltas), Total: total}
	}
	cmp.Deltas.Findings = deltas
	return cmp
}

func abs(v int) int {
	if v < 0 {
		return v
	}
	return v
}
