package slop

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Fuzz invariants per the repository guidance: no panic, valid spans,
// deterministic results, exact source recovery.

var fuzzSeeds = []string{
	"",
	"!?.,;:—–…",
	"The cat sat on the mat.",
	"“It wasn’t the storm — it was the silence after,” she said.",
	"It's not just late, but forgotten. It was never on time. It was early, even.",
	"rock'n'roll ain't o'clock \xff\xfe broken",
	"emoji 😀 and café naïve résumé. It's fine.",
	"a b c d e f g",
	strings.Repeat("It was not empty. It was full. ", 50),
	"One\r\nTwo\r\n\r\nThree\ttabs\t.\u00a0nbsp",
}

func FuzzAnalyzeInvariants(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		a, err := NewAnalyzer()
		if err != nil {
			t.Fatal(err)
		}
		doc, err := a.Analyze("fuzz", input, Options{Full: true})
		if err != nil {
			t.Fatalf("analyze: %v", err)
		}

		// The analyzer sanitizes invalid UTF-8 before reporting; locations
		// refer to the sanitized text, which we reconstruct identically.
		text := input
		if !utf8.ValidString(text) {
			text = strings.ToValidUTF8(text, "\uFFFD")
		}

		for _, f := range doc.Findings {
			for _, loc := range f.Locations {
				if loc.Start < 0 || loc.End > len(text) || loc.Start > loc.End {
					t.Fatalf("invalid span [%d,%d) for %q (len %d)", loc.Start, loc.End, f.Text, len(text))
				}
				if !utf8.ValidString(text[loc.Start:loc.End]) {
					t.Fatalf("span [%d,%d) splits a rune", loc.Start, loc.End)
				}
			}
		}

		// Determinism: same input, same report.
		again, err := a.Analyze("fuzz", input, Options{Full: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(doc.Findings) != len(again.Findings) {
			t.Fatalf("finding count changed between runs: %d vs %d", len(doc.Findings), len(again.Findings))
		}
		for i := range doc.Findings {
			aF, bF := doc.Findings[i], again.Findings[i]
			if aF.ID != bF.ID || aF.Kind != bF.Kind || aF.Text != bF.Text || aF.Count != bF.Count ||
				len(aF.Locations) != len(bF.Locations) {
				t.Fatalf("finding %d changed between runs: %+v vs %+v", i, aF, bF)
			}
			for j := range aF.Locations {
				if aF.Locations[j] != bF.Locations[j] {
					t.Fatalf("finding %d location %d changed between runs", i, j)
				}
			}
		}
		if (doc.Score == nil) != (again.Score == nil) {
			t.Fatal("score presence changed between runs")
		}
		if doc.Score != nil && doc.Score.Value != again.Score.Value {
			t.Fatal("score changed between runs")
		}
	})
}

func FuzzTokenizerSpans(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		n := newNormText(input, quoteRune, true)
		toks := wordsOnlyLower(n)
		prevHi := 0
		for _, tk := range toks {
			if tk.lo < prevHi || tk.hi > len(n.runes) || tk.lo >= tk.hi {
				t.Fatalf("bad token span (%d,%d) of %d runes", tk.lo, tk.hi, len(n.runes))
			}
			prevHi = tk.hi
			start, end := n.byteRange(tk.lo, tk.hi)
			if start < 0 || end > len(input) || start > end {
				t.Fatalf("bad byte range [%d,%d) of %d bytes", start, end, len(input))
			}
			if !utf8.ValidString(input[start:end]) {
				t.Fatalf("token span splits a rune: [%d,%d)", start, end)
			}
		}
		spans := sentenceSpans(n.runes)
		prevEnd := 0
		for _, sp := range spans {
			if sp[0] != prevEnd || sp[1] <= sp[0] {
				t.Fatalf("bad sentence span %v after %d", sp, prevEnd)
			}
			prevEnd = sp[1]
		}
		if len(n.runes) > 0 && len(spans) == 0 {
			t.Fatal("no spans for non-empty input")
		}
	})
}
