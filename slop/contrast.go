package slop

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dlclark/regexp2"
)

// Stage-1 "not X, but Y" contrast patterns, ported verbatim from the pinned
// upstream js/regexes-stage1.js (sam-paech/slop-score @ 289264a, MIT).
// Stage-2 POS-tagged patterns are intentionally not implemented; see
// docs/methodology.md for the measured delta and rationale.
//
// Matching uses regexp2 with ECMAScript semantics so \w, \s, \b and
// case-insensitivity behave as they do for the JavaScript /gi originals.

const (
	maxGap = "160"
	pron   = `(?:it|they|this|that)`
	be     = `(?:is|are|was|were)`
	beNeg  = `(?:is\s+not|are\s+not|was\s+not|were\s+not|isn't|aren't|wasn't|weren't|ain't)`
)

var stage1Patterns = []struct {
	name    string
	pattern string
}{
	{
		"RE_NOT_BUT",
		`\b(?:(?:` + beNeg + `)|not(?!\s+(?:that|only)\b))\s+` +
			`(?:(?!\bbut\b|[.?!]).){1,100}?` +
			`[,;:]\s*but\s+` +
			`(?!when\b|while\b|which\b|who\b|whom\b|whose\b|where\b|if\b|that\b|as\b|because\b|although\b|though\b|till\b|until\b|unless\b|` +
			`here\b|there\b|then\b|my\b|we\b|I\b|you\b|it\s+seems\b|it\s+appears\b|it\s+felt\b|it\s+looks?\b|anything\b)`,
	},
	{
		"RE_NOT_DASH",
		`\b(?:\w+n't|not)\s+(?:just|only|merely)?\s+` +
			`(?:(?![.?!]).){1,` + maxGap + `}?` +
			`(?:-|\s-\s|[—–])\s*` +
			pron + `\s+(?:(?:'re|are|'s|is|were|was)\b|(?!'re|are|'s|is|were|was)[*_~]*[a-z]\w*)`,
	},
	{
		"RE_PRON_BE_NOT_SEP_BE",
		`(?:(?<=^)|(?<=[.?!]\s))\s*["']?` +
			`(?:(?:` + pron + `\s+` + be + `\s+not)|(?:` + pron + `\s+` + be + `n't)|(?:it's|they're|that's)\s+not)\b` +
			`[^.?!]{0,` + maxGap + `}[.;:?!]\s*["']?` +
			pron + `\s+(?:` + be + `|(?:'s|'re))\b(?!\s+not\b)`,
	},
	{
		"RE_NP_BE_NOT_SEP_THEY_BE",
		`(?:(?<=^)|(?<=[.?!]\s))\s*` +
			`(?![^.?!]{0,80}\b(?:knew|know|thought|think|said|says|told|heard|learned)\b[^.?!]{0,40}?\bthat\b)` +
			`(?!\s*not\s+without\b)` +
			`(?![^.?!]{0,50}\bnot\s+put\b)` +
			`[^.?!]{0,` + maxGap + `}?\b(?:` + beNeg + `)\b[^.?!]{0,` + maxGap + `}[.;:?!]\s*` +
			`["']?` + pron + `\b(?:'re|\s+(?:are|were|is|was))\b(?!\s+not\b)`,
	},
	{
		"RE_NO_LONGER",
		`(?:(?<=^)|(?<=[.?!]\s))\s*[^.?!]{0,` + maxGap + `}\bno\s+longer\b[^.;:?!]{0,` + maxGap + `}` +
			`[.;:?!]\s*(?:it|they|this|that)\s+(?:is|are|was|were)\b(?!\s+not\b)`,
	},
	{
		"RE_NOT_JUST_SEP",
		`(?:(?<=^)|(?<=[.?!]\s))\s*["']?` +
			pron + `\b(?:'s|'re|\s+(?:is|are|was|were))?\s+not\s+just\b[^.?!]{0,` + maxGap + `}[.?!]\s*["']?` +
			pron + `\b(?:'s|'re|\s+(?:is|are|was|were))\b(?!\s+not\b)`,
	},
	{
		"RE_NOT_PERIOD_SAMEVERB",
		`(?:(?<=^)|(?<=[.?!]\s))[^.?!]*?\b(?:do|does|did)n't\b\s+` +
			`(?:(?:\w+\s+){0,2})([a-z]{3,})\b[^.?!]*[.?!]\s*` +
			pron + `\s+\1(?:ed|es|s|ing)?\b`,
	},
	{
		"RE_SIMPLE_BE_NOT_IT_BE",
		`(?:(?<=^)|(?<=[.?!]\s))\s*["']?` +
			`(?!he\b|she\b|i\b|you\b|we\b)` +
			`(?![^.?!]{0,80}\b(?:knew|know|thought|think|said|says|told|heard|learned)\b[^.?!]{0,40}?\bthat\b)` +
			`[^.?!]{0,` + maxGap + `}?\b` + beNeg + `\b[^.?!]{0,` + maxGap + `}[.;:?!]\s*` +
			`["']?it(?:'s|\s+(?:is|are|was|were))\b`,
	},
	{
		"RE_EMBEDDED_NOT_JUST_SEP",
		`(?:(?<=^)|(?<=[.?!]\s))` +
			`[^.?!]{0,80}?\b(?:(?:it|they)\s+(?:is|are)|(?:it's|they're))\s+not\s+just\b` +
			`[^.?!]{0,` + maxGap + `}[.?!]\s*` +
			`(?:(?:it|they)\s+(?:is|are)|(?:it's|they're))\b`,
	},
	{
		"RE_DIALOGUE_NOT_JUST",
		`["']?` + pron + `(?:'re|'s|\s+(?:are|is|was|were))\s+not\s+just\b[^"']{0,` + maxGap + `}["']?\s*` +
			`(?:[^.?!]{0,80}\b(?:said|asked|whispered|muttered|replied|added|shouted|cried)\b[^.?!]{0,80}[.?!]\s*)?` +
			`["']?` + pron + `(?:'re|'s|\s+(?:are|is|was|were))\s+[*_~]?[a-z]\w*`,
	},
}

// contrastMatch is one deduplicated stage-1 match mapped onto sentence spans.
type contrastMatch struct {
	sentLo    int // index into sentence spans
	sentHi    int
	rawLo     int // normalized rune offsets of the raw match
	rawHi     int
	pattern   string // e.g. "S1_RE_NOT_BUT"
	matchText string // trimmed matched text on the normalized string
}

type contrastDetector struct {
	patterns []struct {
		name string
		re   *regexp2.Regexp
	}
}

func newContrastDetector() (*contrastDetector, error) {
	d := &contrastDetector{}
	for _, p := range stage1Patterns {
		re, err := regexp2.Compile(p.pattern, regexp2.IgnoreCase|regexp2.ECMAScript)
		if err != nil {
			return nil, fmt.Errorf("compile %s: %w", p.name, err)
		}
		d.patterns = append(d.patterns, struct {
			name string
			re   *regexp2.Regexp
		}{"S1_" + p.name, re})
	}
	return d, nil
}

// coveredSentenceRange finds the [lo, hi] sentence-span indices covering the
// normalized rune range [start, end), mirroring contrast-detector.js.
func coveredSentenceRange(spans [][2]int, start, end int) (int, int, bool) {
	if len(spans) == 0 || start >= end {
		return 0, 0, false
	}
	// lo: first span whose end > start.
	lo := sort.Search(len(spans), func(i int) bool { return spans[i][1] > start })
	// hi: last span whose start < end.
	hi := sort.Search(len(spans), func(i int) bool { return spans[i][0] >= end }) - 1
	if lo >= len(spans) || hi < 0 || lo > hi {
		return 0, 0, false
	}
	return lo, hi, true
}

// mergeIntervals merges candidates overlapping in sentence space, mirroring
// contrast-detector.js mergeIntervals.
func mergeIntervals(cands []contrastMatch) []contrastMatch {
	if len(cands) == 0 {
		return nil
	}
	sort.SliceStable(cands, func(i, j int) bool {
		a, b := cands[i], cands[j]
		if a.sentLo != b.sentLo {
			return a.sentLo < b.sentLo
		}
		if a.sentHi != b.sentHi {
			return a.sentHi < b.sentHi
		}
		return a.rawLo < b.rawLo
	})
	merged := []contrastMatch{cands[0]}
	for _, it := range cands[1:] {
		cur := &merged[len(merged)-1]
		if it.sentLo <= cur.sentHi {
			cur.sentHi = max(cur.sentHi, it.sentHi)
			cur.rawHi = max(cur.rawHi, it.rawHi)
		} else {
			merged = append(merged, it)
		}
	}
	return merged
}

// detect runs all stage-1 patterns over the normalized text and returns
// deduplicated matches in document order.
func (d *contrastDetector) detect(n normText, spans [][2]int) ([]contrastMatch, error) {
	text := string(n.runes)
	var cands []contrastMatch
	for _, p := range d.patterns {
		m, err := p.re.FindStringMatch(text)
		for err == nil && m != nil {
			lo := m.Index
			hi := m.Index + m.Length
			if sentLo, sentHi, ok := coveredSentenceRange(spans, lo, hi); ok {
				cands = append(cands, contrastMatch{
					sentLo:    sentLo,
					sentHi:    sentHi,
					rawLo:     lo,
					rawHi:     hi,
					pattern:   p.name,
					matchText: strings.TrimSpace(text[lo:hi]),
				})
			}
			m, err = p.re.FindNextMatch(m)
		}
		if err != nil {
			return nil, fmt.Errorf("match %s: %w", p.name, err)
		}
	}
	return mergeIntervals(cands), nil
}
