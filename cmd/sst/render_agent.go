package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/lemon07r/SlopScoreTool/slop"
)

// The agent format is a compact, token-efficient report for LLM consumers:
// no provenance block, no calibration detail, no excerpts. Byte offsets let
// the agent re-read the source when it needs context.

func renderAgentHeader(w io.Writer) {
	fmt.Fprintf(w, "sst %s · %s · evidence for editorial judgment, not an AI detector\n",
		slop.Version, slop.CompatibilityStage1Fallback)
}

func renderAgentReport(w io.Writer, report slop.Report) error {
	renderAgentHeader(w)
	for _, doc := range report.Documents {
		renderAgentDocument(w, doc)
	}
	if report.Corpus != nil {
		renderAgentCorpus(w, report.Corpus)
	}
	fmt.Fprintln(w, "guide: repetition and concentration outrank single hits; keep intentional, quoted, or domain-necessary language; judge the prose, not the score")
	return nil
}

func renderAgentDocument(w io.Writer, doc slop.Document) {
	fmt.Fprintf(w, "== %s ==\n", doc.Name)
	if doc.Score != nil {
		var parts []string
		for _, c := range doc.Score.Components {
			parts = append(parts, fmt.Sprintf("%s %.1f", c.Name, c.Points))
		}
		fmt.Fprintf(w, "score: %.1f/100 = %s · orientation only\n", doc.Score.Value, strings.Join(parts, " + "))
	} else {
		fmt.Fprintln(w, "score: n/a (no alphabetic tokens)")
	}
	for _, warn := range doc.Warnings {
		fmt.Fprintf(w, "warning: %s\n", warn)
	}
	fmt.Fprintf(w, "stats: %d words · %d chars · %d sentences · %d paragraphs · %d unique words\n",
		doc.Stats.Words, doc.Stats.Chars, doc.Stats.Sentences, doc.Stats.Paragraphs, doc.Stats.UniqueWords)

	if len(doc.Findings) == 0 {
		fmt.Fprintln(w, "findings: none")
	} else {
		groupCounts := map[string]int{}
		for _, f := range doc.Findings {
			groupCounts[f.Kind]++
		}
		fmt.Fprintf(w, "findings (%d shown):\n", len(doc.Findings))
		lastKind := ""
		for _, f := range doc.Findings {
			if f.Kind != lastKind {
				fmt.Fprintf(w, "%s (%d):\n", groupTitle(f.Kind), groupCounts[f.Kind])
				lastKind = f.Kind
			}
			scored := ""
			if !f.Scored {
				scored = " [not scored]"
			}
			loc := ""
			if len(f.Locations) > 0 {
				loc = fmt.Sprintf(" @%d:%d", f.Locations[0].Start, f.Locations[0].End)
			}
			fmt.Fprintf(w, "- [%s] %q x%d%s%s — %s\n", f.ID, oneLine(f.Text), f.Count, scored, loc, f.Reason)
		}
	}
	if doc.Truncation != nil {
		omitted := 0
		for _, g := range doc.Truncation.Groups {
			omitted += g.Total - g.Shown
		}
		fmt.Fprintf(w, "truncated: %d findings omitted · rerun with --full\n", omitted)
	}
	fmt.Fprintln(w)
}

func renderAgentCorpus(w io.Writer, corpus *slop.CorpusInfo) {
	fmt.Fprintf(w, "== corpus (%d documents) ==\n", corpus.Documents)
	if corpus.Score != nil {
		fmt.Fprintf(w, "scores: min %.1f · median %.1f · mean %.1f · max %.1f\n",
			corpus.Score.Min, corpus.Score.Median, corpus.Score.Mean, corpus.Score.Max)
	}
	if len(corpus.SharedFindings) > 0 {
		fmt.Fprintln(w, "shared findings (2+ docs; habit candidates, not topic words):")
		for _, f := range corpus.SharedFindings {
			fmt.Fprintf(w, "- %s %q in %d docs (x%d total)\n", f.Kind, oneLine(f.Text), f.Documents, f.Count)
		}
	}
	fmt.Fprintln(w)
}

func renderAgentComparison(w io.Writer, cmp slop.Comparison) error {
	renderAgentHeader(w)
	fmt.Fprintf(w, "== compare: %s -> %s ==\n", cmp.Before.Name, cmp.After.Name)
	if cmp.Deltas.Score != nil {
		d := cmp.Deltas.Score
		fmt.Fprintf(w, "score: %.1f -> %.1f (%+.1f) · orientation only, no winner declared\n", d.Before, d.After, d.Delta)
	} else {
		fmt.Fprintln(w, "score: n/a (one side has no alphabetic tokens)")
	}
	for _, c := range cmp.Deltas.Components {
		fmt.Fprintf(w, "- %s: %.2f -> %.2f %s (%+.2f)\n", c.Name, c.Before, c.After, c.Unit, c.Delta)
	}
	changed := 0
	for _, f := range cmp.Deltas.Findings {
		if f.Status != "unchanged" {
			changed++
		}
	}
	if changed == 0 {
		fmt.Fprintln(w, "finding deltas: none changed")
	} else {
		fmt.Fprintf(w, "finding deltas (%d changed):\n", changed)
		for _, f := range cmp.Deltas.Findings {
			if f.Status == "unchanged" {
				continue
			}
			fmt.Fprintf(w, "- %s: %s %q x%d -> x%d\n", f.Status, f.Kind, oneLine(f.Text), f.Before, f.After)
		}
	}
	if cmp.Deltas.Truncation != nil {
		fmt.Fprintf(w, "truncated: showing %d of %d deltas · rerun with higher --max-findings\n",
			cmp.Deltas.Truncation.Shown, cmp.Deltas.Truncation.Total)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "guide: read changed findings in context; a lower score is not a better revision")
	return nil
}
