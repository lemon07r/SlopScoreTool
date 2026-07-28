package slop

import (
	"math"
	"regexp"
	"strings"
)

// Descriptive metrics, ported from the pinned upstream
// generate-leaderboard.mjs. They are context signals only: none of them
// feeds the composite score.

var bareWordRe = regexp.MustCompile(`[A-Za-z0-9_]+`) // JS \b\w+\b with ASCII \w

// splitSentences matches text.split(/[.!?]+/).filter(non-empty-trimmed).
func splitSentences(text string) []string {
	parts := regexp.MustCompile(`[.!?]+`).Split(text, -1)
	out := parts[:0]
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return out
}

// fleschKincaid ports the upstream grade-level estimate, including its
// deliberately simple syllable counter.
func fleschKincaid(text string) float64 {
	sentences := splitSentences(text)
	words := bareWordRe.FindAllString(text, -1)
	if len(sentences) == 0 || len(words) == 0 {
		return 0
	}

	totalSyllables := 0
	for _, w := range words {
		totalSyllables += countSyllables(w)
	}
	avgSyllables := float64(totalSyllables) / float64(len(words))
	avgWords := float64(len(words)) / float64(len(sentences))
	grade := 0.39*avgWords + 11.8*avgSyllables - 15.59
	return math.Max(0, grade)
}

var (
	syllableSuffixRe = regexp.MustCompile(`(?:[^laeiouy]es|ed|[^laeiouy]e)$`)
	syllableGroupRe  = regexp.MustCompile(`[aeiouy]{1,2}`)
)

func countSyllables(word string) int {
	word = strings.ToLower(word)
	if len(word) <= 3 {
		return 1
	}
	word = syllableSuffixRe.ReplaceAllString(word, "")
	word = strings.TrimPrefix(word, "y")
	groups := syllableGroupRe.FindAllString(word, -1)
	if len(groups) == 0 {
		return 1
	}
	return len(groups)
}

// avgSentenceLength ports the upstream average words per sentence.
func avgSentenceLength(text string) float64 {
	sentences := splitSentences(text)
	words := bareWordRe.FindAllString(text, -1)
	if len(sentences) == 0 {
		return 0
	}
	return float64(len(words)) / float64(len(sentences))
}

// splitParagraphs matches text.split(/\n\s*\n/).filter(non-empty-trimmed).
func splitParagraphs(text string) []string {
	parts := regexp.MustCompile(`\n\s*\n`).Split(text, -1)
	out := parts[:0]
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return out
}

// avgParagraphLength ports the upstream average words per paragraph.
func avgParagraphLength(text string) float64 {
	paragraphs := splitParagraphs(text)
	words := bareWordRe.FindAllString(text, -1)
	if len(paragraphs) == 0 {
		return 0
	}
	return float64(len(words)) / float64(len(paragraphs))
}

// dialogueFrequency ports the upstream quoted-dialogue rate per 1k chars:
// count of ", ", and " characters, halved, per 1000 characters.
func dialogueFrequency(text string) float64 {
	if len(text) == 0 {
		return 0
	}
	quotes := strings.Count(text, `"`) + strings.Count(text, "“") + strings.Count(text, "”")
	return (float64(quotes) / 2 / float64(len(text))) * 1000
}

// mattr500 ports the upstream moving-average type-token ratio with a
// 500-token window, falling back to simple TTR on shorter inputs.
func mattr500(toks []token) float64 {
	total := len(toks)
	if total == 0 {
		return 0
	}
	const window = 500
	if total < window {
		return float64(uniqueTokens(toks)) / float64(total)
	}
	sum := 0.0
	count := 0
	for i := 0; i <= total-window; i++ {
		sum += float64(uniqueTokens(toks[i:i+window])) / window
		count++
	}
	return sum / float64(count)
}

func uniqueTokens(toks []token) int {
	seen := make(map[string]struct{}, len(toks))
	for _, t := range toks {
		seen[t.text] = struct{}{}
	}
	return len(seen)
}

// sentenceLengthCV is a diagnostic (not upstream-derived): the coefficient
// of variation of words per sentence. Values near 0 indicate unusually
// uniform cadence. Computed over the same sentence/word splits as the
// descriptive metrics; emitted only with at least two sentences.
func sentenceLengthCV(text string) (float64, bool) {
	sentences := splitSentences(text)
	if len(sentences) < 2 {
		return 0, false
	}
	lengths := make([]float64, len(sentences))
	for i, s := range sentences {
		lengths[i] = float64(len(bareWordRe.FindAllString(s, -1)))
	}
	mean := 0.0
	for _, l := range lengths {
		mean += l
	}
	mean /= float64(len(lengths))
	if mean == 0 {
		return 0, false
	}
	variance := 0.0
	for _, l := range lengths {
		d := l - mean
		variance += d * d
	}
	variance /= float64(len(lengths))
	return math.Sqrt(variance) / mean, true
}
