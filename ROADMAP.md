# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to `TODO_LIST.md`.
> Open questions that need user input are captured here, not in `TODO_LIST.md`.

## Themes

### 1. Ecosystem Integration

Deepen the tool's place in the linting ecosystem. The core loop — detect project type, pick profile, generate config — works. The next leap is connecting it to the surrounding workflow.

Raw ideas:

- **SDK evaluation.** The `linter-autoconfigure-sdk` offers `FindingFromIssue` / `ConfigIssue` for structured finding emission and `ReadConfig` / `LoadJSON` for typed config I/O. The `validate` command currently hand-wraps `os.ReadFile` + `fmt.Errorf` and emits only `slog` messages. Evaluate whether adopting the SDK's typed I/O and finding model is worth the early-adopter tax.
- **Monorepo support.** Auto-detect workspace layouts and generate per-package configs instead of a single root config.
- **CI integration.** First-class `oxlint-auto-configure` as a CI step: validate configs on PR, diff old vs. new config, flag rule drift.
- **Public documentation website.** A docs site (via the `website-launch` skill) for profile reference, rule explorer, and getting-started guide.

### 2. Profile and Rule Intelligence

Profiles are functional but `strict` and `recommended` are currently identical. The rule registry is pinned to oxlint `1.59.0` while runtime is `1.73.0`.

Raw ideas:

- **Profile differentiation.** Decide whether `strict` and `recommended` should produce different output, or remove one and document the equivalence.
- **`--explain` flag.** Print the decision tree for a given project — which plugins were enabled, which rules were set to which severity, and why. Makes the tool self-documenting.
- **`--profile` on `analyze`.** Let `analyze` accept a profile to scope findings to the rules that profile would enable.
- **Automatic rule updates.** A `nix run` app or `flake.nix` target that runs `oxlint -f json --rules > pkg/rule/rules_data.json` and updates the version + test count, reducing the manual update burden.

### 3. Developer Experience

Make the CLI more useful and ergonomic beyond config generation.

Raw ideas:

- **Shell completions.** Cobra completion subcommand for bash, zsh, fish.
- **Structured JSON logs.** `--log-format json` flag for all commands, for CI and tooling integration.
- **Custom output paths.** `--output` / `--config` flexibility for non-standard project layouts (partially covered by existing `--config` flag).
- **TOCTOU protection.** Evaluate `WriteVerified` for `configure` — capture the existing config fingerprint when reading for diff, fail loudly if the user edited `.oxlintrc.json` during execution.

### 4. Architecture and Quality

Keep the codebase clean, well-tested, and production-grade.

Raw ideas:

- **Extract write logic into `pkg/config`.** `writeConfig` lives in `internal/cli`; the atomic-write policy could be a `pkg/config` concern with a `ConfigWriter` interface.
- **Typed errors across packages.** `detect`, `config`, and `oxlint` packages still return generic `error`; migrate to typed errors via the `hierarchical-errors` pattern.
- **BDD tests for all commands.** Behavior-driven tests (via the `bdd-testing` skill) for `configure`, `analyze`, `validate`, `report`.
- **Coverage threshold in CI.** Gate PRs on >=80% coverage to prevent regression.

## Open Questions

Decisions that need user input before they can become actionable tasks:

1. **Adopt `linter-autoconfigure-sdk` for `validate`?** The SDK self-describes as "the weakest of the 5 SDKs" with "modest value over stdlib until a second auto-configurer lands." This is a product direction call: be the first consumer (absorbing early-adopter tax) or wait. _(Source: 2026-07-26 report, question g.1)_
2. **Should `configure` use `WriteVerified` (TOCTOU protection)?** The tool's job is to regenerate config (overwriting is intended), but silently clobbering manual edits made during execution could be surprising. _(Source: 2026-07-26 report, question g.2)_
3. **Should `strict` and `recommended` differ?** They are functionally identical in `pkg/profile/profile.go`. Either differentiate the code or consolidate and document the equivalence. _(Source: 2026-07-22 report, question g.2)_
4. **testify to ginkgo/gomega migration?** Establish a testing framework policy for this project.
5. **Modularization proposal: execute or archive?** The docs were deleted but the decision to pursue modularization remains open.
6. **Markdown or HTML for status reports?** The `status-report` skill prescribes styled HTML dashboards; the user has requested `.md` twice. A split format exists in `docs/status/`. Pick one canonical format and document the decision.
7. **Are `Fingerprint` and `TOCTOU` domain terms or implementation details?** They live in `docs/DOMAIN_LANGUAGE.md` but originate from `go-atomic-write` internals. Decide whether to keep them, move them to AGENTS.md, or restructure DOMAIN_LANGUAGE to separate domain terms from implementation entities.

## Non-goals

Things we are deliberately NOT pursuing, per the project scope boundary:

- **Running oxlint.** This tool generates `.oxlintrc.json`. Running the linter is oxlint's job, not ours.
- **Auto-fixing code.** The `fix` command exists for convenience but is not the core purpose.
- **Watch mode / file watching.** Not aligned with the generate-on-demand model.
- **Replacing oxlint.** This tool configures oxlint; it does not compete with it.
- **Enforcing lint rules.** Detection and reporting only (DryRun=true); we do not enforce.

---

_Last reviewed: 2026-07-26_
