#!/usr/bin/env node
// Oracle runner: scores text with the pinned upstream EQ-Bench slop-score
// implementation in stage-1/fallback mode (POS tagger absent, stage-2
// contrast regexes disabled, exactly as the website behaves when its
// wink-pos-tagger CDN load fails).
//
// Usage: node runner.mjs [--stage2] [file ...]   (reads stdin when no files are given)
// Output: one JSON object per input on stdout, keyed by input name.
//
// --stage2 additionally enables the POS-tagged stage-2 contrast patterns,
// matching the full website pipeline. It requires a one-time dev-only setup:
//   cd testdata/upstream && npm ci
// and running with the CDN loader hook:
//   node --import ./register-loader.mjs runner.mjs --stage2 <files>
// (or `make stage2-delta` at the repository root).
// Goldens are always generated in the default stage-1/fallback mode.
//
// Data sources (single source of truth):
//   - slop lists:  ../../../slop/data/slop_list*.json (canonical embedded copies)
//   - ranges:      ../../../slop/data/calibration.json (datagen output)
//
// Refresh goldens with `make oracle` at the repository root.

import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const dataDir = join(here, '..', '..', 'slop', 'data');

// Patch fetch so the vendored loader reads our canonical data files.
global.fetch = async (url) => {
  const name = String(url).replace(/^\.\/data\//, '');
  const content = readFileSync(join(dataDir, name), 'utf-8');
  return { ok: true, status: 200, json: async () => JSON.parse(content), text: async () => content };
};

// The vendored contrast detector logs verbosely; keep stdout JSON-clean.
for (const k of ['group', 'groupEnd', 'log', 'warn', 'error']) console[k] = () => {};

const stage2 = process.argv.includes('--stage2');
if (stage2) {
  const { initPosTagger } = await import('./js/pos-tagger.js');
  await initPosTagger();
}

const { wordsOnlyLower, alphaTokens } = await import('./js/utils.js');
const { loadSlopSets, computeSlopIndex } = await import('./js/metrics.js');
const { scoreText } = await import('./js/contrast-detector.js');

await loadSlopSets();

const cal = JSON.parse(readFileSync(join(dataDir, 'calibration.json'), 'utf-8'));

function normalize(value, range) {
  const n = (value - range.min) / (range.max - range.min);
  return Math.max(0, Math.min(1, n));
}

function analyze(name, text) {
  const toks = alphaTokens(wordsOnlyLower(text));
  const slop = computeSlopIndex(toks, true);
  const contrast = scoreText(text);

  const nw = normalize(slop.wordScore, cal.ranges.slop_words);
  const nt = normalize(slop.trigramScore, cal.ranges.slop_trigrams);
  const nc = normalize(contrast.rate_per_1k, cal.ranges.contrast_patterns);
  const score = (nw * 0.6 + nc * 0.25 + nt * 0.15) * 100;

  return {
    name,
    chars: text.length,
    tokens: toks.length,
    word_hits: slop.wordHits || [],
    trigram_hits: slop.trigramHits || [],
    word_score_per_1k: slop.wordScore,
    trigram_score_per_1k: slop.trigramScore,
    contrast_hits: contrast.hits,
    contrast_rate_per_1k: contrast.rate_per_1k,
    contrast_matches: contrast.matches.map((m) => ({
      pattern_name: m.pattern_name,
      sentence: m.sentence,
      match_text: m.match_text,
      sentence_count: m.sentence_count,
    })),
    normalized: { words: nw, trigrams: nt, contrast: nc },
    score,
  };
}

async function readStdin() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  return Buffer.concat(chunks).toString('utf-8');
}

const files = process.argv.slice(2).filter((a) => a !== '--stage2');
const out = {};
if (files.length === 0) {
  out.stdin = analyze('stdin', await readStdin());
} else {
  for (const f of files) {
    out[f] = analyze(f, readFileSync(f, 'utf-8'));
  }
}
process.stdout.write(JSON.stringify(out, null, 2) + '\n');
