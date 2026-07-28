package main

import (
	"encoding/json"
	"fmt"
	"io"
	"slop-score/slop"
	"strings"
)

func renderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func renderReport(w io.Writer, format string, report slop.Report) error {
	if format == "json" {
		return renderJSON(w, report)
	}
	for _, doc := range report.Documents {
		renderDocument(w, doc)
	}
	if report.Corpus != nil {
		renderCorpus(w, report.Corpus)
	}
	renderMethodFooter(w, report.Method)
	return nil
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func renderDocument(w io.Writer, doc slop.Document) {
	fmt.Fprintf(w, "== %s ==\n", doc.Name)
	if doc.Score != nil {
		fmt.Fprintf(w, "slop score: %.1f (%s)\n", doc.Score.Value, doc.Score.Basis)
		for _, c := range doc.Score.Components {
			fmt.Fprintf(w, "  %-18s %7.2f %-13s norm %.3f x %.2f = %5.1f pts\n",
				c.Name+":", c.Rate, c.Unit, c.Normalized, c.Weight, c.Points)
		}
	} else {
		fmt.Fprintln(w, "slop score: n/a (no alphabetic tokens)")
	}
	fmt.Fprintf(w, "stats: %d words · %d chars · %d sentences · %d paragraphs · %d unique words\n",
		doc.Stats.Words, doc.Stats.Chars, doc.Stats.Sentences, doc.Stats.Paragraphs, doc.Stats.UniqueWords)

	if len(doc.Signals) > 0 {
		var parts []string
		for _, s := range doc.Signals {
			parts = append(parts, fmt.Sprintf("%s %.2f", s.Name, s.Value))
		}
		fmt.Fprintf(w, "signals: %s\n", strings.Join(parts, " · "))
	}
	for _, warn := range doc.Warnings {
		fmt.Fprintf(w, "warning: %s\n", warn)
	}

	if len(doc.Findings) == 0 {
		fmt.Fprintln(w, "findings: none")
	} else {
		fmt.Fprintln(w, "findings:")
		lastKind := ""
		for _, f := range doc.Findings {
			if f.Kind != lastKind {
				fmt.Fprintf(w, "  %s:\n", groupTitle(f.Kind))
				lastKind = f.Kind
			}
			scored := ""
			if !f.Scored {
				scored = " [not scored]"
			}
			fmt.Fprintf(w, "    [%s] %q x%d%s — %s\n", f.ID, oneLine(f.Text), f.Count, scored, f.Reason)
			if len(f.Locations) > 0 {
				loc := f.Locations[0]
				fmt.Fprintf(w, "         @%d:%d › %s\n", loc.Start, loc.End, oneLine(loc.Excerpt))
			}
		}
	}
	if doc.Truncation != nil {
		omitted := 0
		for _, g := range doc.Truncation.Groups {
			omitted += g.Total - g.Shown
		}
		fmt.Fprintf(w, "truncated: %d findings omitted by --max-findings %d (rerun with --full)\n",
			omitted, doc.Truncation.MaxFindings)
	}
	fmt.Fprintln(w)
}

func groupTitle(kind string) string {
	switch kind {
	case "word":
		return "slop words"
	case "trigram":
		return "slop trigrams"
	case "bigram":
		return "slop bigrams"
	case "contrast":
		return "contrast patterns"
	case "repeated_opening":
		return "repeated sentence openings"
	case "repeated_trigram":
		return "repeated phrases"
	}
	return kind
}

func renderCorpus(w io.Writer, corpus *slop.CorpusInfo) {
	fmt.Fprintf(w, "== corpus (%d documents) ==\n", corpus.Documents)
	if corpus.Score != nil {
		fmt.Fprintf(w, "scores: min %.1f · median %.1f · mean %.1f · max %.1f\n",
			corpus.Score.Min, corpus.Score.Median, corpus.Score.Mean, corpus.Score.Max)
	}
	if len(corpus.SharedFindings) > 0 {
		fmt.Fprintln(w, "shared findings (present in 2+ documents; habit candidates, not topic words):")
		for _, f := range corpus.SharedFindings {
			fmt.Fprintf(w, "  %s %q in %d docs (x%d total)\n", f.Kind, oneLine(f.Text), f.Documents, f.Count)
		}
	}
	fmt.Fprintln(w)
}

func renderComparison(w io.Writer, format string, cmp slop.Comparison) error {
	if format == "json" {
		return renderJSON(w, cmp)
	}
	fmt.Fprintf(w, "== compare: %s -> %s ==\n", cmp.Before.Name, cmp.After.Name)
	if cmp.Deltas.Score != nil {
		d := cmp.Deltas.Score
		fmt.Fprintf(w, "slop score: %.1f -> %.1f (%+.1f)  [%s]\n", d.Before, d.After, d.Delta, d.Basis)
		fmt.Fprintln(w, "  orientation only: judge the prose, not the number")
	} else {
		fmt.Fprintln(w, "slop score: n/a (one side has no alphabetic tokens)")
	}
	for _, c := range cmp.Deltas.Components {
		fmt.Fprintf(w, "  %-18s %7.2f -> %7.2f %-14s (%+.2f)\n", c.Name+":", c.Before, c.After, c.Unit, c.Delta)
	}
	changed := 0
	for _, f := range cmp.Deltas.Findings {
		if f.Status != "unchanged" {
			changed++
		}
	}
	if len(cmp.Deltas.Findings) == 0 {
		fmt.Fprintln(w, "finding deltas: none")
	} else {
		fmt.Fprintf(w, "finding deltas (%d changed of %d shown):\n", changed, len(cmp.Deltas.Findings))
		for _, f := range cmp.Deltas.Findings {
			fmt.Fprintf(w, "  %-10s %s %q x%d -> x%d\n", f.Status+":", f.Kind, oneLine(f.Text), f.Before, f.After)
		}
	}
	if cmp.Deltas.Truncation != nil {
		fmt.Fprintf(w, "truncated: showing %d of %d deltas (rerun with higher --max-findings)\n",
			cmp.Deltas.Truncation.Shown, cmp.Deltas.Truncation.Total)
	}
	fmt.Fprintln(w)
	renderMethodFooter(w, cmp.Method)
	return nil
}

func renderMethodFooter(w io.Writer, m slop.MethodInfo) {
	fmt.Fprintf(w, "method: %s (upstream %s@%s)\n", m.Compatibility, shortRepo(m.Upstream.Repo), shortSHA(m.Upstream.Commit))
	fmt.Fprintln(w, "reminder: evidence for editorial judgment, not an AI detector")
}

func shortRepo(repo string) string {
	return strings.TrimPrefix(repo, "https://github.com/")
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
