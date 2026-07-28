# Build a Writing-Slop Analysis Tool for Agents

## Assignment

Turn this repository into a small, portable Go tool and companion agent skill for reviewing prose for repetitive, canned, or statistically overused AI-writing tendencies.

The product should help a strong writing agent notice concrete tendencies in a draft, reason about them in context, and make better editorial decisions. It must not turn a composite heuristic into an authority. A score can provide orientation; matched passages, repetition, context, genre, intent, and editorial judgment are the useful evidence.

Work autonomously from the sources below. They are a strong starting point, not the boundary of the product. You have broad freedom to correct, simplify, redesign, add, or remove behavior when that better serves the goal. Inspect current upstream code and the local port before settling the design, make good product decisions where this brief leaves room, and deliver a finished, tested repository. Keep the implementation lean enough to understand in one sitting.

## Product principles

1. **Evidence before verdict.** Surface exact findings, locations, counts, rates, contributions, and limitations. Never present the tool as an AI detector or a measure of whether writing is good.
2. **Editorial judgment stays in charge.** Help an agent distinguish a genuine habit from intentional phrasing, domain language, character voice, or necessary contrast.
3. **Improve writing, not the meter.** Discourage synonym-swapping, threshold chasing, flattening voice, and repeated passes toward zero. Prefer changes that improve specificity, rhythm, structure, clarity, or freshness independently of a metric.
4. **Useful dimensions beat a magic number.** Preserve an EQ-inspired score if it aids orientation or comparison, but design the report around interpretable signals. New signals should earn their place through examples or evaluations, not intuition alone.
5. **Deterministic and portable.** Run locally without an LLM, network access, or service. A stable CLI and compact JSON result are more useful across agent harnesses than a bespoke integration.
6. **One source of truth.** Keep analysis logic, lists, calibration, and schemas in canonical machine-readable or code locations. The CLI, docs, and skill should consume or reference them rather than restating drifting copies.
7. **Accuracy means validated behavior.** Go is not intrinsically more accurate than JavaScript. Improvements can come from explicit semantics, correct Unicode and span handling, better detectors, better calibration, and stronger tests. Name compatibility and deviations honestly.

## Starting evidence

Use these as inputs, not unquestionable specifications:

- Local Go seed: `/home/lamim/Development/Vellum/VellumFilter2/internal/filter/slop/`
- Local fixtures: `/home/lamim/Development/Vellum/VellumFilter2/testdata/` and `/home/lamim/Development/Vellum/VellumFilter2/tests/slop_scorer_test.go`
- Local debug CLI: `/home/lamim/Development/Vellum/VellumFilter2/cmd/debug_scorer/`
- Public analyzer and methodology: <https://eqbench.com/slop-score.html>
- Current canonical implementation: <https://github.com/sam-paech/slop-score>
- Site repository supplied by the user: <https://github.com/EQ-bench/EQ-bench-site>

Record the upstream revisions actually used. Preserve required licenses and attribution for imported logic and data.

Important findings to verify rather than rediscover late:

- VellumFilter2 implements word and trigram frequency plus ten stage-one contrast regexes, normalized with leaderboard-derived ranges and weighted 60/25/15.
- Current upstream documentation describes a second POS-tagged contrast stage and additional descriptive metrics. Inspect the actual source and decide what improves this product; do not assume either the local port or current upstream is ideal.
- The local scorer loads a bigram list but does not use it in the composite. Bigrams may be useful evidence, but must not silently alter a score presented as compatible.
- The local debug command explains obsolete 45/40/15 weights while its scorer uses 60/25/15.
- Dataset threshold/statistics functions and the Vellum models dependency are unrelated and should not be ported.
- The embedded leaderboard is roughly 1.4 MB although the scorer uses only derived calibration ranges. Prefer a compact reproducible artifact unless the full data has real runtime value.
- Five local samples offer regression evidence, but loose score tolerances do not prove complete equivalence or editorial value.

## Discover the best product before locking the port

Start with a short evidence-driven design pass:

1. Compare the local Go port with the pinned upstream implementation: tokenization, source lists, normalization, contrast stages, overlap behavior, calibration, secondary metrics, and licensing.
2. Build a small behavior matrix: signal, source behavior, local behavior, proposed behavior, reason, and test/evaluation evidence.
3. Try representative writing: human prose, obvious model-like filler, technical text, dialogue, rhetoric, short text, and mixed-quality revisions.
4. Decide which product concepts should be:
   - a faithfully ported EQ-derived compatibility component;
   - improved deterministic editorial diagnostics;
   - excluded because they add noise, bloat, or false authority.
5. Prefer a multi-signal analysis model over inventing a new opaque aggregate. If an improved aggregate is genuinely useful, expose its definition and keep individual signals primary.

Use an upstream implementation as a differential oracle where compatibility is claimed. Elsewhere, optimize for demonstrated usefulness rather than 1:1 JavaScript precision. A corrected or deliberately different result is welcome when its semantics and evidence are stronger.

## Product opportunity space

Treat this section as a map for exploration, not a feature specification. The implementation agent should select the smallest combination that proves genuinely useful.

The tool could serve several related roles:

- **Draft linter:** inspect one piece of prose and return a compact set of high-value passages worth reconsidering.
- **Corpus profiler:** analyze multiple documents separately and in aggregate to distinguish a repeated author/model habit from a topic-specific word. Preserve document boundaries rather than concatenating inputs blindly.
- **Revision comparator:** show which signals and passages changed between drafts without equating a lower aggregate with a better revision.
- **Agent context provider:** emit a bounded, stable report that another model can use without loading word lists, methodology, or excessive excerpts into context.
- **Method workbench:** make source lists, calibration, and experimental detectors easy to update and evaluate without exposing maintenance machinery in the everyday CLI.

The EQ-derived lexical, phrase, and contrast signals are the natural foundation. Other hypotheses worth testing include:

- repeated sentence openings, transitions, rhetorical templates, or paragraph shapes;
- local n-gram repetition and redundant restatement;
- unusually uniform sentence length, cadence, or syntax;
- concentrations of vague abstraction, canned intensifiers, qualifiers, or meta-commentary;
- formatting habits such as excessive headings, fragments, bullets, or punctuation patterns when they are actually relevant to the writing mode;
- corpus-relative overuse and per-document concentration, which may be more informative than global list membership.

Do not implement that list wholesale. Some signals are easy to count but poor editorial advice, and many legitimate styles are deliberately patterned. Retain a signal only when its finding can be explained, located, tested against counterexamples, and used by an agent without encouraging generic prose.

Potential usability features include finding IDs, “why this was flagged” metadata, per-run finding limits, machine-readable document boundaries, a concise/full report distinction, and lightweight per-run exclusions for known domain terms. Genre-specific profiles, persistent configuration, automatic baselines, and custom rule packs should wait until evaluations show that their benefit exceeds their conceptual and maintenance cost.

The first useful version does not need to answer every writing-quality question. It should establish an extensible result model and a strong evaluation corpus so later signals can be added or removed without breaking the skill or redefining old results silently.

## Intended product shape

Favor a conventional Go module with:

- one reusable analysis package with no CLI concerns;
- one `slop-score` command;
- embedded, versioned analysis data;
- focused tests and representative fixtures;
- a distributable skill under `.agents/skills/` with a clear name such as `review-writing-slop`;
- a concise README and only additional durable documentation that earns its place.

Do not add a server, UI, daemon, database, telemetry, remote API, LLM rewriting, plugin framework, or MCP layer for the first version. The CLI contract is the portable integration surface; adapters can follow real demand.

Keep dependencies few and justified by concrete behavior. At the start of implementation, verify the latest stable Go release and relevant dependency/tool releases from their primary sources. Target the latest stable Go toolchain and select the latest stable compatible libraries rather than copying versions from the older module. Pin chosen versions and test them; if a current release is unsuitable, record the concrete incompatibility instead of silently staying old.

You are authorized to install whatever development tools or dependencies materially help the work, including system packages when that is the best route. Prefer repository-local or user-local installs when they are equally effective. Do not store credentials in this repository.

The repository already includes a Makefile and a deliberately selective golangci-lint v2 configuration. Refresh their pinned tool versions during implementation, preserve their simple target contract, and adjust paths or individual linters when the implemented code gives a concrete reason. `make check` should remain the quick local gate; `make audit` may include slower race and vulnerability checks. Do not enable every available linter or keep a noisy check merely because it exists.

## Analysis contract

Design the core result as structured evidence. Exact names may improve during implementation, but JSON should include the equivalent of:

- schema and analyzer/method versions;
- input statistics relevant to confidence;
- any composite score clearly labeled and decomposed;
- each signal's raw count, rate, calibration or severity basis, and contribution where applicable;
- grouped word, phrase/n-gram, contrast, repetition, and other retained findings;
- locations and short excerpts where practical, with a documented offset convention;
- warnings for short samples, weakly calibrated domains, empty input, truncation, or other interpretation hazards;
- deterministic ranking/truncation metadata so an agent knows what was omitted.

Keep default output compact enough for an agent context. Rank findings by editorial value rather than dumping dictionaries. Provide a finding limit and an explicit fuller-report option.

Do not generate replacement prose in the deterministic analyzer. Its job is to return evidence; the writing agent supplies contextual judgment and revision.

### CLI behavior

Make the common path obvious from `slop-score --help`.

- Accept stdin naturally and support one or more text files.
- Provide clear human output and versioned JSON, for example `--format text|json`.
- Consider a lightweight before/after comparison if it reuses the same result model cleanly. Compare signals and passages; do not declare the lower number the winner.
- Treat adverse findings as successful analysis. Reserve nonzero exits for invalid usage, unreadable input, or analyzer failure.
- Keep stdout machine-clean in JSON mode and diagnostics on stderr.
- Define sensible behavior for empty, tiny, large, Unicode, and multi-document inputs.

Avoid configuration files, universal thresholds, and gating behavior until a real workflow demonstrates the need.

## Data, calibration, and evolution

Separate source-derived compatibility behavior from experimental or improved diagnostics so either can evolve without silently redefining the other. Method metadata should make the distinction inspectable without burdening normal output.

Calibration must be reproducible. If derived constants replace a large leaderboard file, include a small generator or manifest that pins the source, validates its shape, and emits the canonical artifact. Runtime code should not need upstream repositories or network access.

Keep lists and rules versioned and attributable. Prefer a straightforward update/generation path over hand-editing embedded blobs. A future agent should be able to refresh a source, inspect the behavioral diff, run evaluations, and accept or reject the update without creating a second implementation.

## Agent skill

Build the skill with the executable so they share a tested contract. Follow `AGENTS.md` for the boundary between documentation and skills.

The skill metadata should trigger when an agent is drafting, revising, or reviewing prose and would benefit from detecting repetitive AI-associated diction, phrases, structures, or contrast. It should also trigger when the user explicitly asks to reduce AI slop. It must not describe the tool as authorship detection.

Keep `SKILL.md` short and action-oriented. Its workflow should be:

1. Preserve the user's purpose, voice, audience, genre, and hard constraints.
2. Run analysis after a substantive draft exists; do not let a blocklist dictate the first draft.
3. Read warnings and sample context before any aggregate.
4. Inspect the strongest repeated findings and their passages.
5. Decide which are real weaknesses versus justified language.
6. Revise the underlying sentence, paragraph, or idea only where doing so improves the prose independently of the metric.
7. Optionally compare once, then stop when the writing is better. Never optimize toward zero or invent a universal acceptable threshold.
8. Explain editorial choices in normal writing terms, not as obedience to a detector.

Have the skill request compact JSON with only enough findings for context. Include at most a tiny command example and essential interpretation guidance. Put detailed methodology in project docs, not the skill. Add `agents/openai.yaml` when supported and validate the skill structure.

Document the real portability boundary: installing the binary on `PATH` plus copying or linking the skill works for skill-aware harnesses; other harnesses can call the same CLI/JSON contract through a thin adapter. Do not pretend one skill format is natively universal.

## Documentation

The README should quickly answer:

- what the tool is and is not;
- how to install/build it;
- the shortest human and agent invocations;
- how to read findings and scores without over-trusting them;
- how to install/use the skill;
- method sources, limitations, and attribution;
- how to run verification.

Use one methodology/provenance document only if the README would otherwise become crowded. Explain known calibration limits: upstream EQ analysis is optimized for essays and creative writing; short or single samples can be skewed; technical and other domains can under-report or create contextual false positives. Add limitations discovered by this implementation's own evaluations.

Do not create implementation summaries, quick-start duplicates, changelogs, or documents that repeat `--help`, schemas, tests, or one another.

## Verification

Build evidence strong enough to trust the port, improvements, and agent behavior:

- unit tests for normalization, tokenization, Unicode punctuation, n-grams, spans, overlap handling, calibration, and component math;
- positive and negative cases for each retained pattern family, including plausible false positives;
- empty, punctuation-only, tiny, Unicode, multi-paragraph, and multi-document inputs;
- regression checks using the five VellumFilter2 fixtures;
- differential tests against pinned upstream behavior for every compatibility claim;
- comparative fixtures or small labeled evaluations for improved diagnostics;
- CLI tests for stdin, files, JSON schema/stability, clean streams, limits, errors, and exit codes;
- deterministic repeated-run tests;
- skill validation and fresh-agent evaluations on realistic writing tasks.

Skill evaluations should test judgment, not mere invocation. Include cases where an agent should revise repeated filler, retain an intentional “not X but Y,” preserve a necessary technical term, and refuse to infer AI authorship. Give evaluators raw drafts and tool output rather than the intended answer. Iterate if agents chase a number, flatten voice, ignore warnings, or consume excessive context.

Run formatting, vetting, the full test suite, and focused parity/evaluation commands before handoff. Benchmark only if profiling or realistic inputs show a reason; this is an editorial tool, not a throughput contest.

Keep the supplied developer workflow healthy as code arrives: `make format`, `make check`, and `make audit` should work from a fresh checkout after `make tools`. Static-analysis suppressions should be narrow and justified; fixing the code or removing a mismatched linter is preferable to accumulating broad exclusions.

## Acceptance criteria

The work is complete when:

- a fresh checkout builds and tests with straightforward commands;
- the binary analyzes stdin and files offline in useful text and stable JSON forms;
- results expose actionable evidence and uncertainty rather than merely a score;
- all compatibility claims are pinned, accurately named, and supported by differential tests;
- improvements beyond EQ-bench have clear semantics and evaluation evidence;
- data and imported logic have reproducible provenance and correct attribution;
- the skill uses analysis as context while retaining editorial judgment;
- installation and the common workflow are easy to discover;
- the repository has no unrelated VellumFilter2 code, duplicate paths, hollow abstractions, or stale docs;
- `AGENTS.md` remains lean and project guidance has not been duplicated into it.

When tradeoffs arise, favor a smaller, honest, unusually useful tool with an excellent evidence contract over a larger analyzer making stronger claims than its tests support.
