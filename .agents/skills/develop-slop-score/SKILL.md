---
name: develop-slop-score
description: Implement, review, debug, test, benchmark, or modernize this repository's Go writing analyzer and sst CLI. Use for scoring and diagnostic logic, Unicode tokenization and source spans, calibration data, CLI behavior, structured output, embedded assets, or analyzer evaluation.
---

# Develop Slop Score

Start from the current code, tests, executable help, and machine-output schema; treat them as operational truth. Durable method, provenance, and compatibility knowledge lives in `docs/methodology.md`.

## Shape the change

- Keep analysis independent of CLI parsing and presentation.
- Prefer direct code and the standard library. Add a dependency only when it removes meaningful complexity or supplies hard-to-reproduce semantics.
- Preserve one analyzer path and one result model. Remove superseded experiments instead of wrapping them.
- Keep compatibility scoring distinguishable from improved diagnostics. The compatibility mode is named `eqbench-stage1-fallback`; name deliberate deviations and support them with evidence in `docs/methodology.md`.
- Keep rules and calibration data versioned, attributable, inspectable, and reproducible. Embedded lists and calibration come from `make data` (pinned upstream commit); the manifest records provenance and hashes.

## Protect the text contract

- Define whether locations use bytes, runes, or line and column; keep that choice stable in versioned output.
- Separate matching normalization from source text so every finding maps to a valid original span and excerpt.
- Exercise Unicode punctuation, combining characters, invalid UTF-8 policy, overlap, boundaries, empty input, short samples, and large documents.
- Make ordering and truncation deterministic. Never let map iteration determine output.
- Keep aggregate values traceable to raw counts, rates or calibration, and component contributions.

## Keep the CLI composable

- Inject input, output, and diagnostic streams into the command layer so tests do not replace global process state.
- Let only `main` choose a process exit. Return errors from lower layers.
- Write requested results—and only those results—to stdout; write diagnostics to stderr.
- Treat findings as a successful analysis. Reserve nonzero exits for usage or operational failure.
- Preserve the agent contract: `sst --format json --max-findings 20 [FILE...]`, with stdin used when no files are given.

## Verify behavior

- Use focused examples and counterexamples for every heuristic; test intentional rhetoric and necessary domain repetition as well as obvious positives.
- Use differential tests only for compatibility claims against a pinned upstream revision. The vendored Node oracle lives in `testdata/upstream/`; its expectations are committed as `testdata/golden/*.json` and refreshed with `make oracle` (Node only needed for refresh, not for tests).
- Measure diagnostic firing rates with `make eval` (behavior corpus in `testdata/corpus/`) before keeping, cutting, or rethresholding a diagnostic.
- Fuzz tokenizer, normalization, span, and serialization boundaries. Useful invariants include no panic, valid spans, deterministic results, and exact source recovery.
- Test CLI streams, exits, limits, stdin and multi-file behavior, schema versioning, and stable JSON.
- Evaluate editorial usefulness on varied prose. A detector that merely lowers its own score or flattens voice has failed.
- Run `make check` for normal changes and the relevant parts of `make audit` for concurrency, dependency, release, or broad analyzer changes.

## Measure only when useful

- Benchmark representative short, medium, and corpus-sized inputs only when latency or allocation behavior matters.
- Use current Go benchmark APIs and compare repeated runs statistically before claiming an improvement.
- Profile before optimizing. Prefer a clearer algorithmic fix over pools, unsafe code, caching, or concurrency.
- Keep benchmarks out of the default fast check unless they prove cheap and stable.
