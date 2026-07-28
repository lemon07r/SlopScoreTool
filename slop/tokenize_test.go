package slop

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func tokensOf(text string) ([]token, normText) {
	n := newNormText(text, quoteRune, true)
	return alphaTokens(wordsOnlyLower(n)), n
}

func tokenTexts(toks []token) []string {
	out := make([]string, len(toks))
	for i, t := range toks {
		out[i] = t.text
	}
	return out
}

func TestWordsOnlyLower(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"plain", "The cat sat.", []string{"the", "cat", "sat"}},
		{"curly apostrophe", "It’s fine, isn’t it?", []string{"it's", "fine", "isn't", "it"}},
		{"mixed quotes", "“Hello,” she said.", []string{"hello", "she", "said"}},
		{"edge apostrophes stripped", "'tis 'nother one'", []string{"tis", "nother", "one"}},
		{"digits split", "abc123def", []string{"abc", "def"}},
		{"empty", "", nil},
		{"punct only", "!?.,;", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := newNormText(tc.input, quoteRune, true)
			got := tokenTexts(wordsOnlyLower(n))
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAlphaTokens(t *testing.T) {
	// utils.js alphaTokens keeps /^[a-z]+(?:'[a-z]+)?$/ only.
	cases := []struct {
		input string
		keep  bool
	}{
		{"don't", true},
		{"plain", true},
		{"o'clock", true},
		{"rock'n'roll", false}, // two apostrophes
		{"x''y", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isAlphaToken(tc.input); got != tc.keep {
			t.Errorf("isAlphaToken(%q) = %v, want %v", tc.input, got, tc.keep)
		}
	}
}

func TestOffsetMappingWithUnicode(t *testing.T) {
	input := "“Don’t go,” she said—twice."
	toks, n := tokensOf(input)
	want := []string{"don't", "go", "she", "said", "twice"}
	got := tokenTexts(toks)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("tokens %v, want %v", got, want)
	}
	// Byte offsets must recover the original span exactly.
	start, end := n.byteRange(toks[0].lo, toks[0].hi)
	if recovered := input[start:end]; recovered != "Don’t" {
		t.Errorf("recovered %q, want %q", recovered, "Don’t")
	}
	start, end = n.byteRange(toks[4].lo, toks[4].hi)
	if recovered := input[start:end]; recovered != "twice" {
		t.Errorf("recovered %q, want %q", recovered, "twice")
	}
}

func TestSentenceSpans(t *testing.T) {
	runes := []rune("One. Two? Three! Trailing")
	spans := sentenceSpans(runes)
	want := [][2]int{{0, 4}, {4, 9}, {9, 16}, {16, 25}}
	if len(spans) != len(want) {
		t.Fatalf("got %v, want %v", spans, want)
	}
	for i := range want {
		if spans[i] != want[i] {
			t.Errorf("span %d: got %v, want %v", i, spans[i], want[i])
		}
	}
}

func TestCoveredSentenceRange(t *testing.T) {
	spans := [][2]int{{0, 5}, {5, 11}, {11, 18}}
	// Fully inside sentence 1.
	if lo, hi, ok := coveredSentenceRange(spans, 1, 4); !ok || lo != 0 || hi != 0 {
		t.Errorf("inside: got (%d,%d,%v)", lo, hi, ok)
	}
	// Crossing sentences 0 and 1.
	if lo, hi, ok := coveredSentenceRange(spans, 2, 9); !ok || lo != 0 || hi != 1 {
		t.Errorf("crossing: got (%d,%d,%v)", lo, hi, ok)
	}
	// In the gap-free tail region (sentences are contiguous).
	if lo, hi, ok := coveredSentenceRange(spans, 12, 18); !ok || lo != 2 || hi != 2 {
		t.Errorf("tail: got (%d,%d,%v)", lo, hi, ok)
	}
}

func TestMergeIntervalsDedup(t *testing.T) {
	cands := []contrastMatch{
		{sentLo: 0, sentHi: 0, rawLo: 2, rawHi: 8, pattern: "S1_A"},
		{sentLo: 0, sentHi: 0, rawLo: 0, rawHi: 10, pattern: "S1_B"}, // same sentence span
		{sentLo: 2, sentHi: 2, rawLo: 20, rawHi: 25, pattern: "S1_C"},
	}
	merged := mergeIntervals(cands)
	if len(merged) != 2 {
		t.Fatalf("merged = %d, want 2", len(merged))
	}
	// The (sentLo, sentHi, rawLo) sort key puts rawLo 0 (S1_B) first.
	if merged[0].pattern != "S1_B" {
		t.Errorf("first merged pattern = %q, want S1_B", merged[0].pattern)
	}
}

func TestNormalizeClamp(t *testing.T) {
	r := NormRange{Min: -1, Max: 3}
	if got := normalize(-5, r); got != 0 {
		t.Errorf("below: %v", got)
	}
	if got := normalize(99, r); got != 1 {
		t.Errorf("above: %v", got)
	}
	if got := normalize(1, r); got != 0.5 {
		t.Errorf("mid: %v", got)
	}
}

func TestLoadListEntryNormalization(t *testing.T) {
	// Pinned upstream post-fix behavior: apostrophe entries stay intact.
	raw := []byte(`[["i'd spent last"],["Hello, World!"],["plain"]]`)
	set, err := loadList(raw, "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := set["i'd spent last"]; !ok {
		t.Error("apostrophe entry lost")
	}
	// JS takes the first regex match: "hello, world!" -> "hello".
	if _, ok := set["hello"]; !ok {
		t.Error("first-match truncation wrong")
	}
	if _, ok := set["plain"]; !ok {
		t.Error("plain entry lost")
	}
}

func TestUTF16Len(t *testing.T) {
	if got := utf16Len("ascii"); got != 5 {
		t.Errorf("ascii: %d", got)
	}
	if got := utf16Len("café"); got != 4 {
		t.Errorf("café: %d", got)
	}
	if got := utf16Len("a😀b"); got != 4 {
		t.Errorf("astral: %d", got)
	}
}

func TestExcerptBoundaries(t *testing.T) {
	input := "“quoted” text here"
	for start := 0; start <= len(input); start++ {
		for end := start; end <= len(input); end++ {
			out := excerpt(input, start, end, 10)
			if !utf8.ValidString(out) {
				t.Errorf("invalid UTF-8 excerpt for [%d,%d)", start, end)
			}
		}
	}
}
