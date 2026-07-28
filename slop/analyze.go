// Package slop analyzes prose for repetitive, canned, or statistically
// overused AI-associated writing tendencies. It returns structured evidence
// (rates, calibrated components, grouped findings with locations) for an
// editorial agent to judge; it is not an AI-authorship detector and does not
// generate replacement prose.
//
// The composite score and the word, trigram, and contrast rates implement
// the pinned upstream EQ-Bench slop-score method in stage-1/fallback mode
// (POS-tagged stage 2 disabled), as described in docs/methodology.md.
package slop

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// Analyzer holds loaded data and detectors. Construct once and reuse.
type Analyzer struct {
	data     *Data
	detector *contrastDetector
}

// NewAnalyzer loads the embedded data and compiles the contrast patterns.
func NewAnalyzer() (*Analyzer, error) {
	data, err := LoadData()
	if err != nil {
		return nil, err
	}
	detector, err := newContrastDetector()
	if err != nil {
		return nil, err
	}
	return &Analyzer{data: data, detector: detector}, nil
}

// Options tunes report shape. Zero values give the bounded default report.
type Options struct {
	MaxFindings  int      // per finding group; default 10
	Full         bool     // disable finding limits
	ExcludeTerms []string // exact (case-insensitive) finding texts to omit
}

const defaultMaxFindings = 10

// normalize applies padded min-max normalization, clamped to [0, 1]
// (upstream normalizeValue).
func normalize(value float64, r NormRange) float64 {
	n := (value - r.Min) / (r.Max - r.Min)
	return max(0, min(1, n))
}

// utf16Len mirrors JavaScript string.length for rate denominators.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// Input is one named text to analyze.
type Input struct {
	Name string
	Text string
}

// Analyze examines one document and returns its structured result.
func (a *Analyzer) Analyze(name, text string, opts Options) (Document, error) {
	if opts.MaxFindings <= 0 {
		opts.MaxFindings = defaultMaxFindings
	}
	doc := Document{Name: name}

	if !utf8.ValidString(text) {
		text = strings.ToValidUTF8(text, "\uFFFD")
		doc.Warnings = append(doc.Warnings, "invalid_utf8: invalid UTF-8 sequences replaced with U+FFFD before analysis")
	}

	nTok := newNormText(text, quoteRune, true)
	toks := alphaTokens(wordsOnlyLower(nTok))

	nCon := newNormText(text, contrastRune, false)
	spans := sentenceSpans(nCon.runes)
	matches, err := a.detector.detect(nCon, spans)
	if err != nil {
		return doc, err
	}

	doc.Stats = Stats{
		Chars:       len(text),
		Words:       len(toks),
		Sentences:   len(spans),
		Paragraphs:  len(splitParagraphs(text)),
		UniqueWords: uniqueTokens(toks),
	}

	// Warnings before aggregates: confidence context comes first.
	if len(toks) == 0 {
		doc.Warnings = append(doc.Warnings, "empty_input: no alphabetic tokens; score omitted")
	} else {
		if len(text) < 300 {
			doc.Warnings = append(doc.Warnings,
				fmt.Sprintf("short_sample: %d chars (<300); per-1k rates and score are unstable on tiny inputs", len(text)))
		}
		if len(toks) < 100 {
			doc.Warnings = append(doc.Warnings,
				fmt.Sprintf("few_words: %d words (<100); individual findings may be anecdotal", len(toks)))
		}
	}

	// Scored components (upstream-faithful rates and normalization).
	wordTotal, wordFindings := listHits(toks, 1, a.data.Words, nTok, text)
	_, bigramFindings := listHits(toks, 2, a.data.Bigrams, nTok, text)
	trigramTotal, trigramFindings := listHits(toks, 3, a.data.Trigrams, nTok, text)

	if len(toks) > 0 {
		wordsF := float64(len(toks))
		charsF := float64(utf16Len(text))
		wordRate := float64(wordTotal) / wordsF * 1000
		trigramRate := float64(trigramTotal) / wordsF * 1000
		var contrastRate float64
		if charsF > 0 {
			contrastRate = float64(len(matches)) / charsF * 1000
		}

		nw := normalize(wordRate, a.data.WordRange)
		nt := normalize(trigramRate, a.data.TrigramRange)
		nc := normalize(contrastRate, a.data.ContrastRange)
		// Upstream weights: 60% words, 25% contrast, 15% trigrams.
		score := (nw*0.6 + nc*0.25 + nt*0.15) * 100

		doc.Score = &Score{
			Value: score,
			Basis: CompatibilityStage1Fallback,
			Components: []Component{
				{
					Name: "slop_words", Count: wordTotal, Rate: wordRate, Unit: "per_1k_words",
					Range: a.data.WordRange, Normalized: nw, Weight: 0.6, Points: nw * 0.6 * 100,
				},
				{
					Name: "contrast_patterns", Count: len(matches), Rate: contrastRate, Unit: "per_1k_chars",
					Range: a.data.ContrastRange, Normalized: nc, Weight: 0.25, Points: nc * 0.25 * 100,
				},
				{
					Name: "slop_trigrams", Count: trigramTotal, Rate: trigramRate, Unit: "per_1k_words",
					Range: a.data.TrigramRange, Normalized: nt, Weight: 0.15, Points: nt * 0.15 * 100,
				},
			},
		}
	}

	// Non-scored signals.
	if len(toks) > 0 {
		doc.Signals = append(doc.Signals,
			Signal{Name: "vocab_level", Kind: "descriptive", Value: fleschKincaid(text), Unit: "grade_level"},
			Signal{Name: "avg_sentence_length", Kind: "descriptive", Value: avgSentenceLength(text), Unit: "words"},
			Signal{Name: "avg_paragraph_length", Kind: "descriptive", Value: avgParagraphLength(text), Unit: "words"},
			Signal{Name: "dialogue_frequency", Kind: "descriptive", Value: dialogueFrequency(text), Unit: "per_1k_chars"},
			Signal{Name: "mattr_500", Kind: "descriptive", Value: mattr500(toks), Unit: "ratio"},
		)
		if cv, ok := sentenceLengthCV(text); ok {
			doc.Signals = append(doc.Signals, Signal{
				Name: "sentence_length_cv", Kind: "diagnostic", Value: cv, Unit: "ratio",
				Note: "coefficient of variation; low values indicate uniform cadence",
			})
		}
	}

	// Findings, grouped by kind.
	byKind := map[string][]Finding{}
	for i := range wordFindings {
		wordFindings[i].Kind = "word"
		wordFindings[i].Scored = true
		wordFindings[i].Reason = "in EQ-Bench AI-overused word list"
	}
	byKind["word"] = wordFindings
	for i := range trigramFindings {
		trigramFindings[i].Kind = "trigram"
		trigramFindings[i].Scored = true
		trigramFindings[i].Reason = "in EQ-Bench AI-overused trigram list"
	}
	byKind["trigram"] = trigramFindings
	for i := range bigramFindings {
		bigramFindings[i].Kind = "bigram"
		bigramFindings[i].Scored = false
		bigramFindings[i].Reason = "in EQ-Bench AI-overused bigram list (evidence only; not scored)"
	}
	byKind["bigram"] = bigramFindings

	var contrastFindings []Finding
	for _, m := range matches {
		spanLo := spans[m.sentLo][0]
		spanHi := spans[m.sentHi][1]
		start, end := nCon.byteRange(spanLo, spanHi)
		matchStart, matchEnd := nCon.byteRange(m.rawLo, m.rawHi)
		matchText := strings.TrimSpace(text[matchStart:matchEnd])
		const capLen = 120
		if len(matchText) > capLen {
			cut := capLen
			for cut > 0 && !isRuneStart(matchText, cut) {
				cut--
			}
			matchText = matchText[:cut] + " …"
		}
		contrastFindings = append(contrastFindings, Finding{
			Kind:   "contrast",
			Text:   matchText,
			Count:  1,
			Reason: fmt.Sprintf("stage-1 contrast pattern %s (AI-overused \"not X, but Y\" family)", m.pattern),
			Scored: true,
			Locations: []Location{
				{Start: start, End: end, Excerpt: excerpt(text, start, end, 0)},
			},
		})
	}
	byKind["contrast"] = contrastFindings

	if len(toks) > 0 {
		byKind["repeated_opening"] = repeatedOpenings(toks, spans, nTok, text, 2)
		byKind["repeated_trigram"] = repeatedNgrams(contentTokens(toks), 3, 2, nTok, text, "repeated_trigram")
	}

	// Per-run exclusions (exact, case-insensitive finding text).
	if len(opts.ExcludeTerms) > 0 {
		excluded := map[string]struct{}{}
		for _, t := range opts.ExcludeTerms {
			excluded[strings.ToLower(t)] = struct{}{}
		}
		for kind, fs := range byKind {
			kept := fs[:0]
			for _, f := range fs {
				if _, skip := excluded[strings.ToLower(f.Text)]; !skip {
					kept = append(kept, f)
				}
			}
			byKind[kind] = kept
		}
	}

	doc.Findings, doc.Truncation = finalizeFindings(byKind, opts.MaxFindings, opts.Full)
	if doc.Truncation != nil {
		omitted := 0
		for _, g := range doc.Truncation.Groups {
			omitted += g.Total - g.Shown
		}
		doc.Warnings = append(doc.Warnings,
			fmt.Sprintf("findings_truncated: %d findings omitted by the per-group limit; rerun with --full for complete evidence", omitted))
	}
	return doc, nil
}

// AnalyzeAll analyzes each input separately and adds a corpus aggregate when
// more than one document is present. Inputs are never concatenated.
func (a *Analyzer) AnalyzeAll(inputs []Input, opts Options) (Report, error) {
	report := Report{
		SchemaVersion: SchemaVersion,
		Analyzer:      AnalyzerInfo{Name: AnalyzerName, Version: Version},
		Method:        a.methodInfo(),
	}
	for _, in := range inputs {
		doc, err := a.Analyze(in.Name, in.Text, opts)
		if err != nil {
			return report, err
		}
		report.Documents = append(report.Documents, doc)
	}
	if len(report.Documents) > 1 {
		report.Corpus = buildCorpus(report.Documents)
	}
	return report, nil
}

func (a *Analyzer) methodInfo() MethodInfo {
	m := a.data.Manifest
	lists := map[string]string{}
	for name, f := range m.Files {
		hash := f.SHA256
		if len(hash) > 12 {
			hash = hash[:12]
		}
		lists[name] = hash
	}
	return MethodInfo{
		Compatibility: CompatibilityStage1Fallback,
		CompatibilityNote: "Composite score and word/trigram/contrast rates match the pinned upstream " +
			"implementation with the POS-tagged stage-2 contrast patterns disabled (the upstream " +
			"site's documented fallback). Stage-2 is not implemented; see docs/methodology.md for " +
			"the measured delta.",
		Upstream: UpstreamInfo{Repo: m.Source.Repo, Commit: m.Source.Commit},
		Data: DataInfo{
			ListsSHA256:        lists,
			LeaderboardResults: m.Leaderboard.Results,
		},
		DomainNote: "Calibration and lists derive from model essays and creative writing; technical, " +
			"instructional, or other domains may under-report slop or produce contextual false positives. " +
			"Single-topic samples can skew rates.",
	}
}

// buildCorpus aggregates per-document results without merging inputs.
func buildCorpus(docs []Document) *CorpusInfo {
	corpus := &CorpusInfo{Documents: len(docs)}

	scores := make([]float64, 0, len(docs))
	for _, d := range docs {
		if d.Score != nil {
			scores = append(scores, d.Score.Value)
		}
	}
	if len(scores) > 0 {
		sort.Float64s(scores)
		sum := 0.0
		for _, s := range scores {
			sum += s
		}
		median := scores[len(scores)/2]
		if len(scores)%2 == 0 {
			median = (scores[len(scores)/2-1] + scores[len(scores)/2]) / 2
		}
		corpus.Score = &ScoreSummary{
			Min: scores[0], Median: median, Mean: sum / float64(len(scores)), Max: scores[len(scores)-1],
		}
	}

	type key struct{ kind, text string }
	type docSet map[int]struct{}
	shared := map[key]struct {
		docs  docSet
		count int
	}{}
	for i, d := range docs {
		for _, f := range d.Findings {
			k := key{f.Kind, f.Text}
			s := shared[k]
			if s.docs == nil {
				s.docs = docSet{}
			}
			s.docs[i] = struct{}{}
			s.count += f.Count
			shared[k] = s
		}
	}
	for k, s := range shared {
		if len(s.docs) >= 2 {
			corpus.SharedFindings = append(corpus.SharedFindings, SharedFinding{
				Kind: k.kind, Text: k.text, Documents: len(s.docs), Count: s.count,
			})
		}
	}
	sort.Slice(corpus.SharedFindings, func(i, j int) bool {
		a, b := corpus.SharedFindings[i], corpus.SharedFindings[j]
		if a.Documents != b.Documents {
			return a.Documents > b.Documents
		}
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Text < b.Text
	})
	return corpus
}
