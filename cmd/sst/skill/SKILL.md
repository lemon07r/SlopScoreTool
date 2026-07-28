---
name: review-writing-slop
description: Review or revise prose with the SlopScoreTool (sst) analyzer while preserving purpose and voice. Use when drafting, editing, comparing revisions, investigating repetitive or canned AI-associated language, or when a user explicitly asks to reduce AI slop. Do not use it to determine whether AI wrote a text.
---

# Review Writing Slop

Use the analyzer as editorial evidence, not as an authority or optimization target.

## Setup

The `sst` binary must be on `PATH`. If it is missing, install it from the
project releases page or with
`go install github.com/lemon07r/SlopScoreTool/cmd/sst@latest`. Inside the
source repository, build with `make build` if the binary is absent. If the
installed version exposes a different interface, inspect `sst --help` rather
than guessing.

## Analyze

1. Establish the draft's purpose, audience, genre, voice, and hard constraints.
2. Analyze a substantive draft. Do not compose from a prohibited-word list.
3. Request bounded, agent-optimized output:

   ```sh
   sst --format agent --max-findings 20 [FILE...]
   ```

   With no files, pass the draft on stdin. Use `--format brief` instead when a
   structured pipeline consumes the JSON; it is the same evidence with
   provenance, calibration detail, and excerpts stripped. Reserve
   `--format json` for the full-fidelity contract.

4. Read warnings first: short samples, few words, invalid UTF-8, and truncated
   findings all limit confidence. When output reports truncated findings,
   rerun with `--full` before concluding a pattern is absent.
5. Inspect the strongest repeated findings in their surrounding prose. The
   report gives byte offsets (`@start:end`) into the original text; re-read
   the source around them rather than judging phrases in isolation. Treat
   isolated low-confidence matches lightly. Use `--exclude-term` only for
   vocabulary that is necessary to the subject (names, domain terms), never
   to make findings disappear.

For a before/after question, prefer `sst --compare BEFORE AFTER` over two
separate runs; it reports score, component, and finding deltas without
declaring a winner.

## Judge

- Ask whether each finding reveals repetition, empty emphasis, vague
  abstraction, redundant explanation, canned structure, or another weakness a
  reader would notice without the metric.
- Keep language that is precise, intentional, genre-appropriate, rhythmic,
  quoted, or necessary to the subject.
- Give repeated patterns and concentrated habits more weight than a single
  flagged phrase. In multi-file runs, shared findings mark author-level
  habits, not topic vocabulary.
- Never infer authorship, competence, honesty, or overall writing quality
  from a score.

## Revise

Revise only findings that independently improve clarity, specificity,
economy, rhythm, or voice. Prefer fixing the underlying sentence or paragraph
over mechanically swapping a flagged word. Preserve useful texture and
unevenness.

For a meaningful revision, optionally compare drafts directly:

```sh
sst --format agent --compare BEFORE.txt AFTER.txt
```

Read the changed findings in context rather than the score delta. Stop when
the prose is better. Do not chase a target score, zero findings, or endless
passes.

Return the improved text when revision was requested. Summarize only the
editorial choices that help the user; mention important findings deliberately
retained when that explains the result.
