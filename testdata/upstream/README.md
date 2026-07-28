# Pinned upstream differential oracle

The vendored JavaScript under `js/` is the canonical EQ-Bench slop-score
implementation, used only as a development-time differential oracle. It is
not part of the shipped analyzer.

- Source: <https://github.com/sam-paech/slop-score>
- Commit: `289264ab2a0df1358ef81ced3c259610ea709514` (2025-11-22, "fix slop
  list normalisation bug")
- License: MIT (see `LICENSE.upstream`; `js/wordfreq.js` is Apache-2.0 and is
  vendored only because `metrics.js` imports it — it is never executed here)

## Compatibility mode

The oracle runs in **stage-1/fallback mode**: the POS tagger is never
initialized, so the 35 stage-2 POS-tagged contrast regexes stay disabled.
This matches the upstream website's documented fallback when its
wink-pos-tagger CDN load fails, and it is the compatibility mode the Go
analyzer implements. The Go module never claims parity with the
POS-enabled configuration; see `docs/methodology.md` for the measured
stage-2 delta (zero on every input tested, for the reason documented
there).

## Stage-2 delta measurement

`--stage2` initializes the POS tagger and appends stage-2 candidates,
matching the full website pipeline. One-time setup:

```sh
cd testdata/upstream && npm ci   # installs pinned wink-pos-tagger 2.2.2
```

Then compare both modes on the behavior corpus:

```sh
make stage2-delta   # or: cd testdata/upstream &&
                    # node --import ./register-loader.mjs runner.mjs --stage2 <files>
```

`loader.mjs` redirects the CDN imports in `js/pos-tagger.js` to the local
package. It points at the CommonJS `src/` entry (loaded as
`module.default`); the `dist/` browser bundle needs `window`/`document`
and cannot run under Node. `node_modules/` is git-ignored; the pinned
`package-lock.json` is committed.

## Use

`runner.mjs` reads files (or stdin) and prints token counts, list hit
counts and maps, stage-1 contrast matches, per-component rates, normalized
values, and the composite score as JSON. Slop lists and calibration ranges
come from the canonical embedded copies in `slop/data/`, so the oracle and
the Go analyzer always read the same data.

Refresh the committed golden expectations with:

```sh
make oracle   # requires node; rewrites testdata/golden/*.json
```

Ordinary `go test` runs always compare against the committed goldens and do
not need node. After a deliberate data or upstream-pin change, rerun
`make oracle`, inspect `git diff testdata/golden/`, and accept or reject the
behavioral change explicitly.
