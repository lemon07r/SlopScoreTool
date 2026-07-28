package main

import (
	"github.com/lemon07r/SlopScoreTool/slop"
)

// The brief format is a trimmed JSON view of the full report for
// context-conscious machine consumers. It drops method provenance,
// calibration detail, signals, and excerpts while keeping every value an
// agent needs to act on findings. The full JSON contract is unchanged.

type briefReport struct {
	SchemaVersion int               `json:"schema_version"`
	Analyzer      slop.AnalyzerInfo `json:"analyzer"`
	Compatibility string            `json:"compatibility"`
	Documents     []briefDocument   `json:"documents"`
	Corpus        *slop.CorpusInfo  `json:"corpus,omitempty"`
}

type briefDocument struct {
	Name       string           `json:"name"`
	Stats      slop.Stats       `json:"stats"`
	Warnings   []string         `json:"warnings,omitempty"`
	Score      *briefScore      `json:"score"`
	Findings   []briefFinding   `json:"findings"`
	Truncation *slop.Truncation `json:"truncation,omitempty"`
}

type briefScore struct {
	Value      float64          `json:"value"`
	Basis      string           `json:"basis"`
	Components []briefComponent `json:"components"`
}

type briefComponent struct {
	Name   string  `json:"name"`
	Count  int     `json:"count"`
	Rate   float64 `json:"rate"`
	Unit   string  `json:"unit"`
	Points float64 `json:"points"`
}

type briefLocation struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type briefFinding struct {
	ID       string         `json:"id"`
	Kind     string         `json:"kind"`
	Text     string         `json:"text"`
	Count    int            `json:"count"`
	Reason   string         `json:"reason"`
	Scored   bool           `json:"scored"`
	Location *briefLocation `json:"location,omitempty"`
}

type briefComparison struct {
	SchemaVersion int                   `json:"schema_version"`
	Analyzer      slop.AnalyzerInfo     `json:"analyzer"`
	Compatibility string                `json:"compatibility"`
	Before        string                `json:"before"`
	After         string                `json:"after"`
	Deltas        slop.ComparisonDeltas `json:"deltas"`
}

func briefReportFrom(report slop.Report) briefReport {
	out := briefReport{
		SchemaVersion: report.SchemaVersion,
		Analyzer:      report.Analyzer,
		Compatibility: report.Method.Compatibility,
		Corpus:        report.Corpus,
	}
	for _, doc := range report.Documents {
		out.Documents = append(out.Documents, briefDocumentFrom(doc))
	}
	return out
}

func briefDocumentFrom(doc slop.Document) briefDocument {
	out := briefDocument{
		Name:       doc.Name,
		Stats:      doc.Stats,
		Warnings:   doc.Warnings,
		Truncation: doc.Truncation,
	}
	if doc.Score != nil {
		score := &briefScore{Value: doc.Score.Value, Basis: doc.Score.Basis}
		for _, c := range doc.Score.Components {
			score.Components = append(score.Components, briefComponent{
				Name: c.Name, Count: c.Count, Rate: c.Rate, Unit: c.Unit, Points: c.Points,
			})
		}
		out.Score = score
	}
	for _, f := range doc.Findings {
		bf := briefFinding{
			ID: f.ID, Kind: f.Kind, Text: f.Text, Count: f.Count,
			Reason: f.Reason, Scored: f.Scored,
		}
		if len(f.Locations) > 0 {
			bf.Location = &briefLocation{Start: f.Locations[0].Start, End: f.Locations[0].End}
		}
		out.Findings = append(out.Findings, bf)
	}
	return out
}

func briefComparisonFrom(cmp slop.Comparison) briefComparison {
	return briefComparison{
		SchemaVersion: cmp.SchemaVersion,
		Analyzer:      cmp.Analyzer,
		Compatibility: cmp.Method.Compatibility,
		Before:        cmp.Before.Name,
		After:         cmp.After.Name,
		Deltas:        cmp.Deltas,
	}
}
