# Behavior corpus

Evaluation corpus for the diagnostic decision gates and methodology claims.
Not regression fixtures: scores here are observed, not golden-locked.

| File | Label | Intent |
| --- | --- | --- |
| human-essay.txt | human-like | Hand-written personal essay for this corpus |
| human-review.txt | human-like | Hand-written critical review for this corpus |
| model-like-fiction.txt | model-like | Canonical model-style fiction (cascades, tapestries, "not just X") |
| model-like-essay.txt | model-like | Canonical model-style listicle essay |
| technical-docs.txt | technical | Documentation prose (control group for false positives) |
| dialogue-heavy.txt | dialogue-heavy | Quotation-dense scene (control group) |
| rhetorical.txt | rhetorical | Deliberate anaphora/parallelism (control group) |
| short-note.txt | short | ~50-word note (tests short-input warnings) |
| pair-before.txt / pair-after.txt | revision pair | Same scene, model-style draft vs hand revision (compare-mode eval) |

The five VellumFilter2 regression fixtures in `../fixtures/` (sample1-5.txt) are
reported as their own `vellum-fixture` group: their provenance is known but their
human/AI labels were not carried over, so the eval makes no claim about them.

Run the evaluation with `make eval` (see `tools/eval/`).
