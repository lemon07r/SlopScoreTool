---
name: review-writing-slop
description: Review or revise prose with the slop-score analyzer while preserving purpose and voice. Use when drafting, editing, comparing revisions, investigating repetitive or canned AI-associated language, or when a user explicitly asks to reduce AI slop. Do not use it to determine whether AI wrote a text.
---

# Review Writing Slop

Use the analyzer as editorial evidence, not as an authority or optimization target.

## Analyze

1. Establish the draft's purpose, audience, genre, voice, and hard constraints.
2. Analyze a substantive draft. Do not compose from a prohibited-word list.
3. Request bounded machine output:

   ```sh
   slop-score --format json --max-findings 20 [FILE...]
   ```

   With no files, pass the draft on stdin. In this source repository, build with `make build` if the binary is absent. If the installed version exposes a different interface, inspect `slop-score --help` rather than guessing.

4. Read sample-size, calibration, truncation, and method warnings before interpreting aggregates.
5. Inspect the strongest repeated findings in their surrounding prose. Treat isolated low-confidence matches lightly.

## Judge

- Ask whether each finding reveals repetition, empty emphasis, vague abstraction, redundant explanation, canned structure, or another weakness a reader would notice without the metric.
- Keep language that is precise, intentional, genre-appropriate, rhythmic, quoted, or necessary to the subject.
- Give repeated patterns and concentrated habits more weight than a single flagged phrase.
- Never infer authorship, competence, honesty, or overall writing quality from a score.

## Revise

Revise only findings that independently improve clarity, specificity, economy, rhythm, or voice. Prefer fixing the underlying sentence or paragraph over mechanically swapping a flagged word. Preserve useful texture and unevenness.

For a meaningful revision, optionally analyze once more and compare the changed findings in context. Stop when the prose is better. Do not chase a target score, zero findings, or endless passes.

Return the improved text when revision was requested. Summarize only the editorial choices that help the user; mention important findings deliberately retained when that explains the result.
