package slop

// Schema and analyzer identity for versioned machine-readable output.
const (
	SchemaVersion = 1
	Version       = "0.1.0"
	AnalyzerName  = "slop-score"

	// CompatibilityStage1Fallback names the compatibility basis precisely:
	// the composite score, list hit rates, and contrast detection match the
	// pinned upstream implementation with its POS-tagged second stage
	// disabled (the upstream website's documented fallback when the
	// wink-pos-tagger CDN is unavailable).
	CompatibilityStage1Fallback = "eqbench-stage1-fallback"
)

// Report is the top-level analyzer result.
type Report struct {
	SchemaVersion int          `json:"schema_version"`
	Analyzer      AnalyzerInfo `json:"analyzer"`
	Method        MethodInfo   `json:"method"`
	Documents     []Document   `json:"documents"`
	Corpus        *CorpusInfo  `json:"corpus,omitempty"`
}

// AnalyzerInfo identifies the program producing the report.
type AnalyzerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MethodInfo describes the analysis basis, data provenance, and calibration
// caveats without burdening normal output.
type MethodInfo struct {
	Compatibility     string       `json:"compatibility"`
	CompatibilityNote string       `json:"compatibility_note"`
	Upstream          UpstreamInfo `json:"upstream"`
	Data              DataInfo     `json:"data"`
	DomainNote        string       `json:"domain_note"`
}

// UpstreamInfo pins the source implementation used for compatibility.
type UpstreamInfo struct {
	Repo   string `json:"repo"`
	Commit string `json:"commit"`
}

// DataInfo summarizes the embedded data provenance.
type DataInfo struct {
	ListsSHA256        map[string]string `json:"lists_sha256"`
	LeaderboardResults int               `json:"leaderboard_results"`
}

// Document is the analysis of one input text.
type Document struct {
	Name       string      `json:"name"`
	Stats      Stats       `json:"stats"`
	Warnings   []string    `json:"warnings,omitempty"`
	Score      *Score      `json:"score"`
	Signals    []Signal    `json:"signals,omitempty"`
	Findings   []Finding   `json:"findings"`
	Truncation *Truncation `json:"truncation,omitempty"`
}

// Stats are input statistics relevant to result confidence.
type Stats struct {
	Chars       int `json:"chars"`
	Words       int `json:"words"`
	Sentences   int `json:"sentences"`
	Paragraphs  int `json:"paragraphs"`
	UniqueWords int `json:"unique_words"`
}

// Score is the decomposed EQ-derived composite. It orients; it is not an
// authorship or quality verdict.
type Score struct {
	Value      float64     `json:"value"`
	Basis      string      `json:"basis"`
	Components []Component `json:"components"`
}

// Component is one scored signal with its calibration basis and contribution.
type Component struct {
	Name       string    `json:"name"`
	Count      int       `json:"count"`
	Rate       float64   `json:"rate"`
	Unit       string    `json:"unit"`
	Range      NormRange `json:"calibration_range"`
	Normalized float64   `json:"normalized"`
	Weight     float64   `json:"weight"`
	Points     float64   `json:"points"`
}

// Signal is a retained non-scored measurement (descriptive or diagnostic).
type Signal struct {
	Name  string  `json:"name"`
	Kind  string  `json:"kind"` // "descriptive" or "diagnostic"
	Value float64 `json:"value"`
	Unit  string  `json:"unit,omitempty"`
	Note  string  `json:"note,omitempty"`
}

// Location is a byte-offset span into the original UTF-8 input: [Start, End).
type Location struct {
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Excerpt string `json:"excerpt"`
}

// Finding is one grouped, ranked, locatable piece of evidence.
type Finding struct {
	ID        string     `json:"id"`
	Kind      string     `json:"kind"` // "word", "trigram", "bigram", "contrast", "repeated_opening", "repeated_trigram"
	Text      string     `json:"text"`
	Count     int        `json:"count"`
	Reason    string     `json:"reason"`
	Scored    bool       `json:"scored"` // contributes to the composite score
	Locations []Location `json:"locations,omitempty"`
}

// Truncation records deterministic finding limits and what was omitted.
type Truncation struct {
	MaxFindings int                        `json:"max_findings"`
	Groups      map[string]GroupTruncation `json:"groups,omitempty"`
}

// GroupTruncation counts shown vs. total findings in one kind group.
type GroupTruncation struct {
	Shown int `json:"shown"`
	Total int `json:"total"`
}

// CorpusInfo aggregates per-document results without concatenating inputs.
type CorpusInfo struct {
	Documents      int             `json:"documents"`
	Score          *ScoreSummary   `json:"score,omitempty"`
	SharedFindings []SharedFinding `json:"shared_findings,omitempty"`
}

// ScoreSummary describes the distribution of document scores.
type ScoreSummary struct {
	Min    float64 `json:"min"`
	Median float64 `json:"median"`
	Mean   float64 `json:"mean"`
	Max    float64 `json:"max"`
}

// SharedFinding is a finding present in more than one document, which
// suggests an author- or model-level habit rather than topic vocabulary.
type SharedFinding struct {
	Kind      string `json:"kind"`
	Text      string `json:"text"`
	Documents int    `json:"documents"`
	Count     int    `json:"count"`
}
