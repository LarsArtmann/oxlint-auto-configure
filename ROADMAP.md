# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to `TODO_LIST.md`.
> Open questions that need user input are captured here, not in `TODO_LIST.md`.

## Themes

### 1. Ecosystem Integration

Deepen the tool's place in the linting ecosystem. The core loop — detect project type, pick profile, generate config — works. The next leap is connecting it to the surrounding workflow.

Raw ideas:

- **Monorepo support.** Auto-detect workspace layouts and generate per-package configs instead of a single root config.
- **CI integration.** First-class `oxlint-auto-configure` as a CI step: validate configs on PR, diff old vs. new config, flag rule drift.
- **Public documentation website.** A docs site (via the `website-launch` skill) for profile reference, rule explorer, and getting-started guide.

### 2. Profile and Rule Intelligence

Profiles are functional and differentiated: `strict` (default) sets correctness+suspicious at error and everything else at warn; the former `recommended` alias was removed after the two were confirmed byte-identical (see Resolved Questions). The rule registry is refreshed to oxlint `1.82.0` and matches the installed runtime.

Raw ideas:

- **`--explain` flag.** Print the decision tree for a given project — which plugins were enabled, which rules were set to which severity, and why. Makes the tool self-documenting.
- **`--profile` on `analyze`.** Let `analyze` accept a profile to scope findings to the rules that profile would enable.
- **Automatic rule updates.** A `nix run` app or `flake.nix` target that runs `oxlint -f json --rules > pkg/rule/rules_data.json` and updates the version + test count, reducing the manual update burden.

### 3. Developer Experience

Make the CLI more useful and ergonomic beyond config generation.

Raw ideas:

- **Shell completions.** Cobra completion subcommand for bash, zsh, fish.
- **Structured JSON logs.** `--log-format json` flag for all commands, for CI and tooling integration.
- **Custom output paths.** `--output` / `--config` flexibility for non-standard project layouts (partially covered by existing `--config` flag).
- **`--dry-run` with diff.** Show what would change without writing. Partially covered by the existing diff logging in `Configure`, but a dedicated `--dry-run` flag would surface this to the user explicitly.

### 4. Architecture and Quality

Keep the codebase clean, well-tested, and production-grade.

Raw ideas:

- **Extract write logic into `pkg/config`.** `writeConfig` lives in `internal/cli`; the atomic-write policy could be a `pkg/config` concern with a `ConfigWriter` interface.
- **Typed errors across packages.** `detect`, `config`, and `oxlint` packages still return generic `error`; migrate to typed errors via the `hierarchical-errors` pattern.
- **BDD tests for all commands.** Behavior-driven tests (via the `bdd-testing` skill) for `configure`, `analyze`, `validate`, `report`.
- **Coverage threshold in CI.** Gate PRs on >=80% coverage to prevent regression.
- **Upstream toolsdk `Outputs`.** Propose `Outputs []string` on `Spec` so BuildFlow regains `**/.oxlintrc.json` producer edges (lost in the `ProviderFromSpec` migration; affects dependabot the same way).
- **`doctor` command.** One-shot CLI diagnosis: oxlint version vs. `jsPlugins` needs, plugin installed-but-unregistered, registered-but-uninstalled, registry-version drift. (The config-drift slice is now covered by the provider's HealthCheck; the remaining checks are CLI-facing ideas.)
- **Docker image with oxlint.** The distroless image ships the binary only; `analyze`/`--fix` cannot run inside it. Either bake oxlint in or document the limitation.

## Open Questions

Decisions that need user input before they can become actionable tasks:

1. **Adopt `linter-autoconfigure-sdk` for `validate`?** → **Resolved 2026-09-22: adopted.** `validate` reads configs through the SDK's typed `LoadJSON[config.OxlintConfig]`; a missing config stays programmatically detectable (`errors.Is(err, fs.ErrNotExist)` via the SDK's `*ConfigError` chain) and the error names the fix. The provider already bridged through `ProviderFromSpec`.
2. **Should `strict` and `recommended` differ?** → **Resolved 2026-09-22: `recommended` removed, `strict` survives.** They were byte-identical; two names for one behavior invited drift. `-p recommended` now returns a migration error naming `strict` (`profile.ErrProfileRemoved`); README documents the change.
3. **testify to ginkgo/gomega migration?** → **Resolved 2026-09-22: testify stays; Ginkgo allowed for NEW behavior specs only.** Existing table-driven testify tests are not migrated; the first command-level behavior spec is the trigger to add the Ginkgo dependency.
5. **Markdown or HTML for status reports?** → **Resolved 2026-09-22: Markdown canonical — "always just markdown unless I ask for HTML" (owner).** HTML is produced only on explicit request.
6. **Config-drift detection scope?** → **Resolved 2026-09-22: advisory drift reporting via the toolsdk Spec's HealthCheck** (owner: "use toolsdk to the max"). BuildFlow treats health-check failures as warn-log + summary only — it never skips the tool nor triggers Repair, so drift is visible without stomp risk. Detect stays missing-only; Repair keeps never-overwrite.
7. **Preserve `overrides` blocks for external rules?** → **Resolved 2026-09-22: preserved wholesale** (owner: "preserve all, with smart deduplication"). `PreserveExternal` copies every `overrides` block verbatim, deduplicating exact duplicates by canonical JSON form (key-order-insensitive). The generator never emits the field.
8. **GitHub Discussions on/off, and social-preview branding?** → **Resolved 2026-09-22: Discussions stay off** (issues only); social-preview branding not pursued.

## Resolved Questions

- **Should `configure` use `WriteVerified` (TOCTOU protection)?** → **No.** `Write(path, data)` is correct. The tool's job is to regenerate config; overwriting is intended behavior. `go-atomic-write` v0.4.0 provides both APIs; we use plain `Write` for crash-durability without TOCTOU checking.
- **Are `Fingerprint` and `TOCTOU` domain terms?** → **No.** They are `go-atomic-write` API concepts (implementation details), not oxlint-auto-configure domain terms. Kept OUT of `docs/DOMAIN_LANGUAGE.md`.

## Non-goals

Things we are deliberately NOT pursuing, per the project scope boundary:

- **Running oxlint.** This tool generates `.oxlintrc.json`. Running the linter is oxlint's job, not ours.
- **Auto-fixing code.** The `fix` command exists for convenience but is not the core purpose.
- **Watch mode / file watching.** Not aligned with the generate-on-demand model.
- **Replacing oxlint.** This tool configures oxlint; it does not compete with it.
- **Enforcing lint rules.** Detection and reporting only (DryRun=true); we do not enforce.

---

_Last reviewed: 2026-09-22_
