package slop

import (
	"fmt"
	"sort"
	"strings"
)

// stopwords is the NLTK English stopword list (179 words) used by the
// pinned upstream contentTokens filter. Diagnostic only; nothing here feeds
// the composite score.
var stopwords = wordSet(`i me my myself we our ours ourselves you your yours
yourself yourselves he him his himself she her hers herself it its itself they
them their theirs themselves what which who whom this that these those am is
are was were be been being have has had having do does did doing a an the and
but if or because as until while of at by for with about against between into
through during before after above below to from up down in out on off over
under again further then once here there when where why how all any both each
few more most other some such no nor not only own same so than too very s t
can will just don should now`)

func wordSet(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for w := range strings.FieldsSeq(s) {
		out[w] = struct{}{}
	}
	return out
}

// contentTokens mirrors metrics.js: alpha tokens minus stopwords.
func contentTokens(toks []token) []token {
	out := make([]token, 0, len(toks))
	for _, t := range toks {
		if _, stop := stopwords[t.text]; !stop {
			out = append(out, t)
		}
	}
	return out
}

const maxLocations = 5

// excerpt returns a bounded original-text excerpt around [start, end),
// expanded by ctx bytes on each side and trimmed to rune boundaries.
func excerpt(text string, start, end, ctx int) string {
	lo := max(start-ctx, 0)
	hi := min(end+ctx, len(text))
	for lo > 0 && !isRuneStart(text, lo) {
		lo--
	}
	for hi < len(text) && !isRuneStart(text, hi) {
		hi++
	}
	out := text[lo:hi]
	const capBytes = 220
	if len(out) > capBytes {
		cut := start + capBytes/2
		if cut >= end {
			cut = end - 1
		}
		for cut > 0 && !isRuneStart(text, cut) {
			cut--
		}
		out = text[lo:cut] + " …"
	}
	if lo > 0 {
		out = "… " + out
	}
	if hi < len(text) && !strings.HasSuffix(out, " …") {
		out += " …"
	}
	return out
}

func isRuneStart(s string, i int) bool {
	return i < len(s) && (s[i]&0xC0) != 0x80
}

// location builds a Location from normalized rune spans via the offset map.
func location(n normText, text string, lo, hi, ctx int) Location {
	start, end := n.byteRange(lo, hi)
	return Location{Start: start, End: end, Excerpt: excerpt(text, start, end, ctx)}
}

// listHits counts sliding-window phrase hits against a list set, recording
// up to maxLocations locations per distinct phrase.
func listHits(toks []token, size int, set map[string]struct{}, n normText, text string) (int, []Finding) {
	type acc struct {
		count int
		locs  []Location
	}
	byPhrase := map[string]*acc{}
	total := 0
	for i := 0; i+size <= len(toks); i++ {
		parts := make([]string, size)
		for j := range size {
			parts[j] = toks[i+j].text
		}
		phrase := strings.Join(parts, " ")
		if _, ok := set[phrase]; !ok {
			continue
		}
		total++
		a := byPhrase[phrase]
		if a == nil {
			a = &acc{}
			byPhrase[phrase] = a
		}
		a.count++
		if len(a.locs) < maxLocations {
			a.locs = append(a.locs, location(n, text, toks[i].lo, toks[i+size-1].hi, 30))
		}
	}
	findings := make([]Finding, 0, len(byPhrase))
	for phrase, a := range byPhrase {
		findings = append(findings, Finding{Text: phrase, Count: a.count, Locations: a.locs})
	}
	return total, findings
}

// repeatedOpenings finds sentence-first alpha tokens occurring at least
// minRepeat times (diagnostic).
func repeatedOpenings(toks []token, spans [][2]int, n normText, text string, minRepeat int) []Finding {
	type acc struct {
		count int
		locs  []Location
	}
	byWord := map[string]*acc{}
	tokIdx := 0
	for _, sp := range spans {
		for tokIdx < len(toks) && toks[tokIdx].hi <= sp[0] {
			tokIdx++
		}
		if tokIdx < len(toks) && toks[tokIdx].lo < sp[1] {
			t := toks[tokIdx]
			a := byWord[t.text]
			if a == nil {
				a = &acc{}
				byWord[t.text] = a
			}
			a.count++
			if len(a.locs) < maxLocations {
				a.locs = append(a.locs, location(n, text, t.lo, t.hi, 30))
			}
		}
	}
	var findings []Finding
	for word, a := range byWord {
		if a.count >= minRepeat {
			findings = append(findings, Finding{
				Kind:      "repeated_opening",
				Text:      word,
				Count:     a.count,
				Reason:    fmt.Sprintf("%d sentences begin with %q", a.count, word),
				Locations: a.locs,
			})
		}
	}
	return findings
}

// repeatedNgrams finds in-document n-grams of the given token stream
// occurring at least minRepeat times (diagnostic, evidence only).
func repeatedNgrams(toks []token, size, minRepeat int, n normText, text string, kind string) []Finding {
	type acc struct {
		count int
		locs  []Location
	}
	byPhrase := map[string]*acc{}
	for i := 0; i+size <= len(toks); i++ {
		parts := make([]string, size)
		for j := range size {
			parts[j] = toks[i+j].text
		}
		phrase := strings.Join(parts, " ")
		a := byPhrase[phrase]
		if a == nil {
			a = &acc{}
			byPhrase[phrase] = a
		}
		a.count++
		if len(a.locs) < maxLocations {
			a.locs = append(a.locs, location(n, text, toks[i].lo, toks[i+size-1].hi, 30))
		}
	}
	var findings []Finding
	for phrase, a := range byPhrase {
		if a.count >= minRepeat {
			findings = append(findings, Finding{
				Kind:      kind,
				Text:      phrase,
				Count:     a.count,
				Reason:    fmt.Sprintf("phrase occurs %d times in this document", a.count),
				Locations: a.locs,
			})
		}
	}
	return findings
}

// kindOrder fixes the output order of finding groups.
var kindOrder = map[string]int{
	"word":             0,
	"trigram":          1,
	"bigram":           2,
	"contrast":         3,
	"repeated_opening": 4,
	"repeated_trigram": 5,
}

var kindPrefix = map[string]string{
	"word":             "w",
	"trigram":          "t",
	"bigram":           "b",
	"contrast":         "c",
	"repeated_opening": "o",
	"repeated_trigram": "r",
}

// sortGroup orders one finding group deterministically: document order for
// contrast, repetition-weighted (count desc, text asc) elsewhere.
func sortGroup(kind string, fs []Finding) {
	sort.Slice(fs, func(i, j int) bool {
		a, b := fs[i], fs[j]
		if kind == "contrast" && len(a.Locations) > 0 && len(b.Locations) > 0 {
			return a.Locations[0].Start < b.Locations[0].Start
		}
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return a.Text < b.Text
	})
}

// finalizeFindings ranks groups, applies the per-group limit, assigns IDs,
// and reports truncation.
func finalizeFindings(byKind map[string][]Finding, maxPerGroup int, full bool) ([]Finding, *Truncation) {
	kinds := make([]string, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, k)
	}
	sort.Slice(kinds, func(i, j int) bool { return kindOrder[kinds[i]] < kindOrder[kinds[j]] })

	var out []Finding
	var trunc *Truncation
	for _, kind := range kinds {
		fs := byKind[kind]
		sortGroup(kind, fs)
		total := len(fs)
		if !full && total > maxPerGroup {
			fs = fs[:maxPerGroup]
			if trunc == nil {
				trunc = &Truncation{MaxFindings: maxPerGroup, Groups: map[string]GroupTruncation{}}
			}
			trunc.Groups[kind] = GroupTruncation{Shown: len(fs), Total: total}
		}
		prefix := kindPrefix[kind]
		for i := range fs {
			fs[i].ID = fmt.Sprintf("%s%d", prefix, i+1)
		}
		out = append(out, fs...)
	}
	return out, trunc
}
