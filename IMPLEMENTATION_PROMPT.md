# Build a Writing-Slop Analysis Tool for Agents

## Mission

Turn this repository into a small, portable Go analyzer and companion agent skill that help writers notice repetitive, canned, or statistically overused AI-writing tendencies.

The goal is better editorial judgment, not obedience to a number. A score may orient or compare, but the useful product is evidence: the passages involved, repeated habits, component signals, context, and uncertainty. Do not present the tool as an AI-authorship detector or a general measure of writing quality.

The sources below are starting evidence, not a product boundary. Inspect them, explore the feature space, and choose the smallest design that proves broadly useful. You may correct, simplify, extend, or depart from both the JavaScript implementation and the older Go port.

## Sources and known starting points

- Go seed: `/home/lamim/Development/Vellum/VellumFilter2/internal/filter/slop/`
- Fixtures and tests: `/home/lamim/Development/Vellum/VellumFilter2/testdata/` and `/home/lamim/Development/Vellum/VellumFilter2/tests/slop_scorer_test.go`
- Debug command: `/home/lamim/Development/Vellum/VellumFilter2/cmd/debug_scorer/`
- Public method: <https://eqbench.com/slop-score.html>
- Current upstream: <https://github.com/sam-paech/slop-score>
- Site source supplied by the user: <https://github.com/EQ-bench/EQ-bench-site>

Pin the upstream revisions actually used and preserve their licenses and attribution.

Verify these observations early:

- VellumFilter2 scores words, trigrams, and ten stage-one contrast patterns with leaderboard-derived normalization and 60/25/15 weights.
- Current upstream describes an additional POS-tagged contrast stage and descriptive metrics. Decide what is worth carrying forward from the actual source.
- The Go seed loads bigrams without using them in its composite, contains unrelated dataset-threshold helpers, embeds a large leaderboard to derive a few ranges, and has a debug display with obsolete weights.
- The five existing samples are useful regressions, not proof of parity or editorial value.
- Go is not inherently more accurate than JavaScript. Better semantics, Unicode and span handling, calibration, detectors, and tests are the opportunities.

## Explore the product

Treat this as a feature map, not a checklist. A useful tool could combine some of these roles:

- **Draft linting:** identify a bounded set of passages worth reconsidering.
- **Corpus profiling:** analyze documents separately and together so recurring habits can be distinguished from topic-specific language.
- **Revision comparison:** show changed signals and passages without declaring that the lower score is automatically better.
- **Agent context:** emit a compact report that a writing model can interpret without loading rule lists or methodology.
- **Method workbench:** make lists, calibration, and experimental detectors easy to update and evaluate without exposing that machinery in normal use.

EQ-derived word, phrase, and contrast signals are the natural baseline. Other hypotheses worth testing include local repetition, repeated openings and transitions, rhetorical templates, redundant restatement, unusually uniform cadence or paragraph shape, vague abstraction, canned intensifiers or qualifiers, and formatting habits that are genuinely relevant to the writing mode.

Do not implement those ideas wholesale. Retain a signal only when it can be explained, located, tested against counterexamples, and used without pushing agents toward blander prose. Prefer multiple interpretable dimensions to a new opaque aggregate.

Useful interface experiments may include stable finding IDs, short “why flagged” descriptions, compact and full reports, per-run finding limits, document boundaries, corpus-relative concentration, and lightweight exclusions for necessary domain terms. Persistent configuration, genre profiles, custom rule packs, and additional integrations can wait for evidence that they earn their complexity.

## Likely first shape

A strong default is:

- a reusable analysis package with no CLI concerns;
- one offline `slop-score` command;
- embedded and versioned analysis data;
- human-readable output and a compact versioned JSON contract;
- stdin and file input, with clear multi-document behavior;
- focused tests, representative evaluation fixtures, and a distributable `.agents/skills/review-writing-slop` skill;
- a concise README and, only if needed, one methodology/provenance document.

This is a preference, not a prohibition. A server, UI, MCP adapter, remote API, automatic rewriting, plugin framework, or substantial configuration system probably does not belong in the first version unless exploration reveals a compelling need.

Start with the latest stable Go release and latest stable compatible dependencies, verified from primary sources. Pin chosen versions and record only meaningful exceptions. Installing useful developer or system tools is in scope.

The supplied Makefile and golangci-lint configuration establish the developer workflow. Refresh their pinned tools and adapt command paths or noisy linters as the code takes shape. Keep `make check` quick and `make audit` thorough.

## Analysis result

Design the result around structured evidence. Exact fields are your decision, but an agent-facing report should make it possible to recover:

- schema, analyzer, data, and method versions;
- document statistics that affect interpretation;
- each retained signal's raw count, rate or calibration basis, and aggregate contribution if any;
- grouped findings with counts, locations, short excerpts, and concise reasons;
- warnings for short samples, weakly calibrated domains, empty input, or truncation;
- deterministic ordering and an indication of omitted findings.

Keep default JSON small enough for an agent context. Offer fuller evidence deliberately rather than dumping dictionaries. The analyzer should explain findings, not generate replacement prose.

The CLI should keep machine output clean, send diagnostics to stderr, reserve nonzero exits for operational failures, and treat adverse findings as a successful analysis. Avoid universal pass/fail thresholds.

## Method and data

Where EQ compatibility is useful, separate it from improved diagnostics and test it against a pinned upstream oracle. A deliberate deviation is welcome when it produces clearer or more useful behavior; label it instead of silently redefining an old score.

Make calibration and data refreshes reproducible. If a compact generated artifact can replace the large leaderboard, preserve enough source metadata and generation logic to audit and regenerate it. Runtime analysis should remain local and deterministic.

Version source lists and rules. An update should produce an inspectable behavioral diff and run the evaluation corpus before replacing the canonical data.

## Agent skill

Build the skill with the executable so its commands and interpretation cannot drift. Its metadata should trigger for drafting, revision, prose review, repetitive AI-associated language, or an explicit request to reduce AI slop—not for authorship detection.

Keep `SKILL.md` short. It should direct an agent to:

1. Preserve purpose, voice, audience, genre, and hard constraints.
2. Analyze a substantive draft rather than writing from a blocklist.
3. Read sample warnings before aggregates and inspect the strongest repeated findings.
4. Decide which findings are real weaknesses and which are intentional or necessary.
5. Revise only when the change improves the prose independently of the metric.
6. Optionally compare once, then stop when the writing is better rather than chasing zero.

Have the skill request bounded JSON and explain editorial choices in normal writing terms. Keep detailed methodology in project documentation. Validate the skill structure and forward-test it with fresh agents using raw drafts and analyzer output.

Installing the binary on `PATH` plus copying or linking the skill should cover skill-aware harnesses. Other harnesses can call the same CLI/JSON contract through a thin adapter.

## Validation and finish line

Use upstream differential tests for compatibility claims and independent examples for product usefulness. Cover tokenization, Unicode punctuation, n-grams, spans, overlap handling, calibration, empty and short input, multi-document behavior, deterministic output, CLI streams and exits, and stable JSON.

Use varied prose: human, model-like, technical, dialogue-heavy, rhetorical, short, and revised. Include counterexamples where intentional contrast or a necessary technical term should remain. Skill evaluations should catch number chasing, voice flattening, ignored warnings, authorship claims, and excessive context use.

Keep dependencies and public APIs small. Run formatting, static analysis, tests, race detection, vulnerability checks, and focused evaluations before handoff. Benchmark only if realistic use reveals a performance concern.

The result is ready when a fresh checkout builds easily; analysis works offline for stdin and files; reports expose actionable evidence and limitations; compatibility and improvements are honestly named and tested; data provenance is reproducible; the skill exercises judgment; common use is easy to discover; and no duplicate implementation paths, stale docs, or unrelated VellumFilter2 code remain.
