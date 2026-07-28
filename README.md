# slop-score

Analyze prose for repetitive, canned, or statistically overused AI-associated
writing tendencies. The tool returns structured editorial evidence, rates,
calibrated score components, and grouped findings with exact locations, so a
writer (or an agent) can judge what to revise.

It is not an AI-authorship detector and not a measure of writing quality. A
score orients review; the findings are the product.

- Offline and deterministic. No network access at runtime, no configuration.
- Single static Go binary plus a versioned JSON contract for agents.
- Scoring core compatible with the pinned upstream EQ-Bench method in the
  precisely named `eqbench-stage1-fallback` mode (see
  [docs/methodology.md](docs/methodology.md)).

## Install

Requires a current stable Go toolchain.

```sh
git clone <this repo> && cd Slop-Score-Tool
make install      # go install ./cmd/slop-score (into your GOBIN)
# or: make build  # writes bin/slop-score in the repo
```

No module path is published yet; building from a checkout is the supported
path.

### Use with an agent harness

Put the binary on `PATH`, then copy (or symlink) the skill into your
harness's skill directory:

```sh
cp -r .agents/skills/review-writing-slop <skills-dir>/   # e.g. .claude/skills/, .factory/skills/, .agents/skills/
```

Harnesses without skill support can call the same CLI/JSON contract
directly.

## Use

```sh
slop-score draft.txt                 # analyze files (or stdin when none given)
cat draft.txt | slop-score           # stdin
slop-score --format json draft.txt   # compact machine report (stdout stays clean)
slop-score --compare before.txt after.txt   # revision comparison, no winner declared
slop-score --full draft.txt          # all findings, no per-group limit
slop-score --exclude-term kubernetes docs/*.md   # omit a necessary domain term
```

Multiple files are analyzed independently, never concatenated, and the JSON
report adds a corpus aggregate (score spread, findings shared across
documents). Diagnostics go to stderr. Adverse findings are a successful
analysis: the exit code stays 0; nonzero exits mean usage or operational
failure.

Example text report:

```text
== draft.txt ==
slop score: 68.9 (eqbench-stage1-fallback)
  slop_words:          22.22 per_1k_words  norm 0.481 x 0.60 =  28.9 pts
  contrast_patterns:    2.78 per_1k_chars  norm 1.000 x 0.25 =  25.0 pts
  slop_trigrams:        5.56 per_1k_words  norm 1.000 x 0.15 =  15.0 pts
stats: 180 words · 1087 chars · 14 sentences · 3 paragraphs · 114 unique words
findings:
  slop words:
    [w1] "tapestry" x1 — in EQ-Bench AI-overused word list
         @538:546 › … crucible with ease, weaving a tapestry of words that draws everyone …
  ...
```

## Interpret

- **Score** (0-100): weighted mix of slop-word rate (60%), stage-1 contrast
  patterns (25%), and slop-trigram rate (15%), each padded-min-max
  normalized against the upstream leaderboard calibration. Treat it as
  orientation for where to look, never as a verdict.
- **Findings**: located, countable evidence: listed words/trigrams, "not X,
  but Y"-family contrast patterns, plus non-scored diagnostics (repeated
  sentence openings, repeated content trigrams, bigrams as evidence only).
  Repetition and concentration matter more than any single hit.
- **Signals**: descriptive context (vocabulary level, sentence and paragraph
  shape, dialogue frequency, lexical diversity, cadence variation).
- **Warnings** come first: short samples, few words, empty input, invalid
  UTF-8, and truncated findings all limit confidence. When findings are
  truncated, rerun with `--full` before concluding a pattern is absent.

Keep language that is precise, intentional, rhythmic, quoted, or necessary
to the subject. Use `--exclude-term` for necessary vocabulary, never to hide
findings.

## Limits

- Calibration and lists derive from model essays and creative writing.
  Technical, instructional, or other domains may under-report slop, and
  necessary domain repetition can produce contextual false positives.
- Inputs under ~300 chars or ~100 words produce unstable rates and scores.
- Dialogue-heavy and deliberately rhetorical prose can fire the repetition
  diagnostics; that is evidence to weigh, not a defect to erase.
- The tool explains findings. It never rewrites text, and it never decides
  who or what wrote a text.

## Method, data, compatibility

The scoring core implements the pinned upstream EQ-Bench slop-score method
(`sam-paech/slop-score@289264a`, MIT) in `eqbench-stage1-fallback` mode,
verified by golden differential tests against the vendored upstream
JavaScript oracle. Data provenance, calibration derivation, the measured
stage-2 delta, deliberate deviations, and the behavior-corpus evaluation
are documented in [docs/methodology.md](docs/methodology.md).

## Develop

```sh
make check   # format, mod, lint, tests (fast gate)
make audit   # check + race detector + vulnerability scan (thorough gate)
make help    # all targets, incl. data/oracle/eval regeneration
```

## License

MIT (this project; see [LICENSE](LICENSE)). Upstream method, lists, and the
vendored oracle carry their own licenses; see [NOTICE](NOTICE).
