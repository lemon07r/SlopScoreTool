# Working Agreement

This is an experimental tool. Optimize for a useful, coherent artifact and fast learning; do not import production ceremony into it.

## How to work

- Take ownership of the requested outcome. Inspect the relevant evidence, make sensible assumptions, and finish the change. Ask only when a missing choice would materially change the product.
- Prefer the smallest design that explains itself. Standard library and direct code usually beat frameworks, speculative abstractions, configuration, compatibility layers, and defensive machinery with no demonstrated need.
- Use the latest stable Go toolchain, libraries, and development tools available when doing the work. Pin what the repository relies on, verify upgrades, and document only meaningful exceptions.
- Keep one active path for each capability. Improve or replace existing code instead of adding parallel versions, wrappers, summaries, or “new” variants.
- Verify claims against code, tests, evaluations, or primary sources. Preserve useful uncertainty rather than turning a heuristic into a fact.
- Make tests proportional to behavior and risk. A change is done when the relevant behavior works, focused checks pass, and nearby stale material has been removed.
- Refactor freely when it leaves the repository simpler. Preserve unrelated user work; do not preserve accidental complexity merely because it already exists.
- Install useful development tools and dependencies as needed, preferably repository-local or user-local; system installation is also in scope when it is the sensible option. Keep credentials out of tracked files and documentation.

## Keep the repository fresh

- Let code, tests, and executable help be the operational truth. Keep generated data reproducible and record its source and version near the generator or manifest.
- Update affected documentation in the same change. Delete superseded guidance rather than layering corrections on top.
- When a recurring failure reveals a durable lesson, encode it at the narrowest useful level: a test first, then tooling, documentation, or a skill only when those add distinct value.
- Do not store session history, implementation diaries, duplicated plans, or facts that a capable agent can cheaply infer.

## Documentation and skills

Give each fact one canonical home and link to it elsewhere:

- `README.md` is the concise user path: purpose, install, core commands, interpretation, and limitations.
- `docs/` is for durable design, methodology, provenance, or contributor knowledge that would overload the README. Create a document only when it will remain useful.
- `AGENTS.md` contains stable repository-wide working norms, not project status or a second architecture document.
- `SKILL.md` is a short, conditional playbook for an agent: precise trigger metadata, the tool workflow, how to interpret results, and when judgment should override a metric. It should not duplicate the README, teach generic writing, or embed implementation details.
- Put optional, domain-specific material in a skill's `references/` only when progressive disclosure saves context. Keep scripts for deterministic repeated actions and avoid decorative skill files.

The executable interface and versioned machine-readable output are the contract between the program and its skill. Prefer having the skill invoke that contract over copying rules or word lists into prose.

## Agent guidance

Assume the agent is capable. State intent, constraints, truth sources, and acceptance evidence; leave implementation sequencing to current context. Use repository skills only when they clearly match the task. Avoid model-specific prompt forks or rituals unless an evaluation demonstrates a real advantage across the models used here.

Keep this file lean and suggestive. Amend it only for a durable, repository-wide lesson, and consolidate before adding.
