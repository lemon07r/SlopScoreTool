# Working Agreement

This is an experimental tool. Optimize for a useful, coherent artifact and fast learning; avoid production ceremony without a demonstrated benefit.

## Work

- Own the requested outcome. Inspect relevant evidence, make sensible assumptions, and continue until it is complete. Do not invent human-like stamina or deadline limits, or defer clear in-scope work because it is difficult, tedious, broad, or a convenient follow-up. Ask only when a missing choice would materially change the product.
- Prefer the smallest design that explains itself. Each new file, abstraction, dependency, check, or workflow should address a concrete need here. Direct code and the standard library usually beat speculative machinery or production-grade security, release, process, and portability work without relevant stakes.
- Use current stable Go, libraries, and development tools. Pin repository dependencies, verify upgrades, and keep an older version only for a concrete compatibility reason.
- Verify claims against code, tests, evaluations, or primary sources. Keep useful uncertainty visible rather than turning a heuristic into a fact.
- Test behavior in proportion to its risk. A change is done when the relevant behavior works, focused checks pass, and nearby stale material is gone.
- Refactor freely when the result is simpler. Preserve unrelated user work, not accidental complexity.
- Install tools needed for the work; prefer repository-local or user-local installation when it is equally effective.

## Keep the repository fresh

- Maintain one active path for each capability. Replace superseded code and guidance instead of adding parallel versions.
- Let code, tests, and executable help define operational truth. Keep generated data reproducible and record its source beside the generator or manifest.
- Update affected documentation with the change and delete obsolete guidance rather than layering corrections onto it.
- Encode recurring lessons at the narrowest useful level: usually a test, then tooling, documentation, or a skill when it adds distinct value.
- Do not keep session history, implementation diaries, duplicate summaries, or facts a capable agent can cheaply infer.

## Documentation and skills

Give each fact one canonical home:

- `README.md` is the concise user path: purpose, installation, core use, interpretation, and limitations.
- `docs/` holds durable methodology, provenance, design, or contributor knowledge only when the README would become crowded.
- `AGENTS.md` contains stable repository-wide working norms, not project status or architecture.
- `SKILL.md` contains a short conditional workflow and non-obvious domain or tool knowledge. Do not use skills to restate generic engineering practice, the README, or implementation details.
- Add skill references or scripts only when progressive disclosure or deterministic reuse saves meaningful context.

Keep this file lean and suggestive. Consolidate before adding a durable repository-wide lesson.
