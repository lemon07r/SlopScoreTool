package slop

import "unicode"

// normText holds normalized text as runes alongside a map back to byte
// offsets in the original UTF-8 input. Normalization is 1:1 per rune
// (quote/dash substitution plus optional Unicode lowercase), so rune index i
// in the normalized text starts at byte orig[i] in the input.
type normText struct {
	runes []rune
	orig  []int // len(runes)+1; orig[len(runes)] == len(input)
}

// quoteRune implements utils.js normalizeQuotes: collapses the full set of
// curly and alternate quote marks onto ASCII ' and ".
func quoteRune(r rune) rune {
	switch r {
	case '‘', '’', '‚', '‛', '′', 'ʼ', '＇', '`':
		return '\''
	case '“', '”', '„', '‟', '″', '«', '»', '＂':
		return '"'
	}
	return r
}

// contrastRune implements utils.js normalizeText: the six-character mapping
// applied before contrast detection (quotes and dashes only, no case fold).
func contrastRune(r rune) rune {
	switch r {
	case '“', '”':
		return '"'
	case '‘', '’':
		return '\''
	case '—', '–':
		return '-'
	}
	return r
}

// newNormText normalizes s one rune at a time, recording original byte
// offsets. Mirrors the JavaScript preprocessing, which is also 1:1 per
// UTF-16 code unit for the mappings used here.
func newNormText(s string, mapRune func(rune) rune, lower bool) normText {
	n := normText{runes: make([]rune, 0, len(s)), orig: make([]int, 0, len(s)+1)}
	for byteOff, r := range s {
		n.orig = append(n.orig, byteOff)
		out := mapRune(r)
		if lower {
			out = unicode.ToLower(out)
		}
		n.runes = append(n.runes, out)
	}
	n.orig = append(n.orig, len(s))
	return n
}

// byteRange maps a normalized rune span [lo, hi) to original byte offsets.
func (n normText) byteRange(lo, hi int) (int, int) {
	return n.orig[lo], n.orig[hi]
}

// token is one wordsOnlyLower token with its normalized rune span.
type token struct {
	text string
	lo   int // rune index, inclusive
	hi   int // rune index, exclusive
}

// wordsOnlyLower matches utils.js: lowercase, normalize quotes, take runs of
// [a-z'], then strip leading and trailing apostrophes from each run.
func wordsOnlyLower(n normText) []token {
	var toks []token
	start := -1
	flush := func(end int) {
		if start < 0 {
			return
		}
		lo, hi := start, end
		for lo < hi && n.runes[lo] == '\'' {
			lo++
		}
		for hi > lo && n.runes[hi-1] == '\'' {
			hi--
		}
		if lo < hi {
			toks = append(toks, token{text: string(n.runes[lo:hi]), lo: lo, hi: hi})
		}
		start = -1
	}
	for i, r := range n.runes {
		if (r >= 'a' && r <= 'z') || r == '\'' {
			if start < 0 {
				start = i
			}
		} else {
			flush(i)
		}
	}
	flush(len(n.runes))
	return toks
}

// isAlphaToken reports whether s matches /^[a-z]+(?:'[a-z]+)?$/ (utils.js
// alphaTokens): letters with at most one interior apostrophe.
func isAlphaToken(s string) bool {
	i := 0
	for i < len(s) && s[i] >= 'a' && s[i] <= 'z' {
		i++
	}
	if i == 0 {
		return false
	}
	if i == len(s) {
		return true
	}
	if s[i] != '\'' {
		return false
	}
	i++
	j := i
	for j < len(s) && s[j] >= 'a' && s[j] <= 'z' {
		j++
	}
	return j > i && j == len(s)
}

// alphaTokens filters tokens to /^[a-z]+(?:'[a-z]+)?$/ (utils.js alphaTokens).
func alphaTokens(toks []token) []token {
	out := toks[:0]
	for _, t := range toks {
		if isAlphaToken(t.text) {
			out = append(out, t)
		}
	}
	return out
}

// sentenceSpans matches utils.js /[^.!?]*[.!?]/gs plus the trailing-text
// span: each span ends at the first [.!?] after its start.
func sentenceSpans(runes []rune) [][2]int {
	var spans [][2]int
	start := 0
	for i, r := range runes {
		if r == '.' || r == '!' || r == '?' {
			spans = append(spans, [2]int{start, i + 1})
			start = i + 1
		}
	}
	if start < len(runes) {
		spans = append(spans, [2]int{start, len(runes)})
	}
	return spans
}
