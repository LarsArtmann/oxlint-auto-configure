# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `overrides` preservation: regeneration now copies every `overrides` block from an existing `.oxlintrc.json` verbatim (deduplicating exact duplicates by canonical JSON form, key order irrelevant). The generator never emits the field, so blocks like @shadcn/lint's component-dir disables can no longer be silently dropped — the same data-loss class `PreserveExternal` fixed for `jsPlugins`/rules/settings. The config differ compares overrides too.
- Provider `HealthCheck` (toolsdk v1.13.0): an existing `.oxlintrc.json` that no longer matches what the tool would generate is reported as **advisory drift** — the error names the fix (`oxlint-auto-configure configure`) and states nothing was modified. BuildFlow treats health-check failures as warn-log + summary only (never skips the tool, never triggers a repair), so drift is visible without any risk of Repair stomping user customizations. Detect stays missing-only; Repair keeps its never-overwrite contract. Missing configs and user-curated `.oxlintrc.jsonc`/`oxlint.config.json` files are healthy.
- `validate` now reads configs through linter-autoconfigure-sdk's typed `LoadJSON[config.OxlintConfig]` (first SDK consumer for CLI I/O). A missing config produces an error naming the fix (`run oxlint-auto-configure configure`) that stays programmatically detectable via `errors.Is(err, fs.ErrNotExist)` through the SDK's `*ConfigError` chain.
- `profile.Parse` — the single doorway for profile-name strings: the removed `recommended` name maps to an actionable migration error (`profile.ErrProfileRemoved`, names `strict` as the replacement); unknown names wrap `profile.ErrInvalidProfile` listing the valid choices.
- `Differ.HasChanges()` — cheap boolean drift check over the full config comparison; the provider health check uses it instead of hand-rolled length tests.
- External JS plugin support (`@shadcn/lint`): `configure` detects `@shadcn/lint` in `package.json` dependencies and registers it under `jsPlugins` (requires oxlint >= 1.80; warns when the oxlint in PATH is older). Its `shadcn/*` design-system rules are never enabled automatically — the tool points to the rule docs instead of fighting your design-system policy. Regenerating a config preserves an existing setup verbatim: `jsPlugins`, every `shadcn/*` rule including array-form options, and `settings.shadcn` (`pkg/config/preserve.go`). `validate` accepts external rules (reported as `external`, not unknown) and understands oxlint's array-form rule values.
- `pkg/rule/external.go` — known external-plugin table (`@shadcn/lint` → `shadcn`) with lookup helpers; `pkg/detect.DetectExternalPlugins()` scans dependencies and dev-dependencies.
- Direct test coverage for the shared `internal/testregistry` helper (previously only transitively exercised).
- `scripts/pre-release-check.sh` — the full local release gate in one command: module consistency, build, vet, race tests, lint, `goreleaser check`, and a GoReleaser snapshot run. Encodes the manual v0.5.0-session validation that was never codified.
- Disabled-workflow canary (`.github/workflows/workflow-health.yml`): weekly scheduled job asserting every workflow in `.github/workflows/` has state `active` — the Jul–Sep 2026 billing block went unnoticed for ~2 months.
- Post-release smoke step in `release.yml`: downloads the just-published Linux archive from its own GitHub Release and runs `--version`, so a broken release fails visibly instead of silently.
- Repository hygiene: `SECURITY.md` (private vulnerability reporting), bug/feature issue templates, pull-request template, and `CODEOWNERS`.

### Changed

- Embedded rule registry refreshed to oxlint **1.82.0**: **870 rules** (was 841 at 1.73.0), 111 enabled by default. CI now installs oxlint pinned to the embedded version (`oxlint@$(cat pkg/rule/rules_version.txt)`) and fails when the installed binary and `rules_version.txt` drift apart.
- External-plugin handling polish: `configure` spawns `oxlint --version` once per run (was twice), reads `package.json` once for detection (memoized in the detector), logs how many external registrations were preserved, emits the enable-rules hint per detected plugin (was only the first), and warns when a preserved `jsPlugins` entry's package is no longer a dependency.
- GoReleaser config migrated off deprecated keys: `brews` → `homebrew_casks`, `dockers`/`docker_manifests` → a single multi-platform `dockers_v2` entry, archives `format` → `formats`. Container images are now cosign-signed (`docker_signs`, keyless) and get an SBOM (`dockers_v2.sbom`).

### Removed

- **The `recommended` profile.** It produced byte-identical output to `strict` — two names for one behavior invited drift. `-p recommended` now fails with a migration error pointing at `strict`; the CLI default and the BuildFlow Repair pipeline both use `strict` (generated output is unchanged). Scripts and CI pins referencing `recommended` must switch to `strict`.

### Fixed

- **Generated configs are byte-stable across runs.** `encoding/json/v2` serializes map keys in an order that changes between calls on the same map, so regenerating an unchanged config shuffled `rules`/`settings`/`env` key order (noisy VCS diffs, flaky comparisons). All output-facing marshals now set `json.Deterministic(true)`; pinned by `TestToJSONIsDeterministic`. The drift health check depends on this: its before/after comparison flapped without it.
- `analyze` no longer reports a clean project when oxlint itself fails to start. oxlint signals startup failures (e.g. a `jsPlugins` package that cannot load) with exit code 1 and the error text on stdout — indistinguishable from "findings found" — which the pipeline swallowed under graceful degradation into "no findings — project is clean". Detector parse failures now surface oxlint's own message (with a stdout snippet), pipeline partial errors fail the command visibly, and plain notices oxlint prepends to the JSON (e.g. "No files found to lint.") are still parsed correctly instead of becoming false errors.

## [0.6.4] - 2026-09-22

### Changed

- go-finding v1.13.0 (workspace-aware module floors, `go 1.27` root floor);
  toolsdk v1.13.0
- `go` directive normalized to the minor form `go 1.27` (fleet-adopted
  floor form; see ADR-0001 in go-version-auto-configure)

## [0.6.3] - 2026-09-13

### Fixed

- Config detection now recognizes all three oxlint config file names — `.oxlintrc.json`, `.oxlintrc.jsonc`, and `oxlint.config.json` (mirrors BuildFlow's `findOxlintConfig` priority). Previously only `.oxlintrc.json` was detected, so a repo with a curated `.oxlintrc.jsonc` was flagged as missing its config and the repair generated a shadowing `.oxlintrc.json` — violating the tool's never-overwrite contract and silently flipping lint to the wrong config. Regression tests cover both the detection and the no-shadow repair.

## [0.6.2] - 2026-09-11

### Fixed

- Restored the provider's full input-read contract after the `ProviderFromSpec` migration: `spec.Inputs` is again `["package.json", ".oxlintrc.json"]`, so detection re-runs when `package.json` changes (the spec bridge had derived `Inputs` from the config file alone).

## [0.6.1] - 2026-09-11

### Added

- BuildFlow provider now built through `linter-autoconfigure-sdk`'s `ProviderFromSpec` bridge (v0.2.0): the missing-config finding is declared as a `ConfigIssue` and the repair as a plain closure, with Trigger/DependsOn layered on the returned Spec. First SDK consumer migration.

### Changed

- DAG flip (owner decision): the provider no longer depends on `oxlint` — BuildFlow's oxlint lint step now depends on the provider, so a missing config is generated before linting in the same run.
- The `OXLOPT_CONFIG_MISSING` finding now carries `FixStrategySuggest`
  instead of `FixStrategyDirect` (go-finding validation requires before/after
  code for direct fixes, which a generate-the-config repair does not have)
  and category `configuration` instead of `style`. Confidence stays high,
  message/suggestion/file are unchanged, and the registered Repair is
  identical, so BuildFlow repair behavior is unchanged.
- CLI detect/repair semantics and the generated `.oxlintrc.json` bytes are
  the same.
- The local `replace` of `linter-autoconfigure-sdk` was removed in favor of the tagged v0.2.0 (a published tag must not carry local path replaces).

## [0.6.0] - 2026-09-11

### Added

- `pkg/provider` package — BuildFlow integration via the `go-finding/toolsdk` v1.10.0 Spec contract. A package-level `toolsdk.Register` declares the tool (name `oxlint-auto-configure`, JS/TS trigger, dry-run-aware detect/repair), detects a missing `.oxlintrc.json` in recognizable JS/TS projects (`OXLOPT_CONFIG_MISSING` warning), and repairs it by generating the recommended-profile config (honoring BuildFlow's dry-run flag; never overwrites an existing config). BuildFlow consumers blank-import the package and drop their hand-written glue.

### Changed

- Nix flake `go-finding` input bumped from the stale `v1.8.0` pin to `v1.10.0`, matching `go.mod` and the `toolsdk` sub-module's required core version.

## [0.5.0] - 2026-09-11

### Added

- `FEATURES.md`, `TODO_LIST.md`, and `ROADMAP.md` for project tracking and honest feature inventory; `docs/DOMAIN_LANGUAGE.md` filled with actual project domain terms, including atomic-write vocabulary.
- `go-atomic-write` v0.5.1 dependency — config writes are crash-durable (temp + fsync + atomic rename); `writeConfig` now uses `atomicwrite.Write` instead of raw `os.WriteFile`, so a crash mid-write can no longer truncate the user's `.oxlintrc.json`.
- Embedded rules data updated from oxlint `1.59.0` to `1.73.0` (716 → 841 rules, 108 → 113 enabled by default).
- Entry-point tests for `cmd/oxlint-auto-configure/main.go`; E2E `configure` round-trip via `config.FromJSON` (`e2e_test.go`); atomic-write contract tests (`atomic_write_test.go`); coverage tests for `renderFindings`, `printSARIF`, `printReportJSON`, `sortedByPosition`, `resolveConfigPath`, `logDiffIfExisting`, `marshalConfigJSON` (`coverage_test.go`).
- Shared `internal/testregistry` package for loading test rule registries.
- Sentinel errors for structured error contracts (`ErrUnexpectedVersionOutput`, `ErrOxlintStderr`, CLI validation sentinels).
- golangci-lint pinned to `v2.12.2` in CI (was `latest`); Nix flake check CI job; `go mod tidy` module-consistency check in CI; Dependabot for Go modules.
- Full version metadata in nix builds (`commit`, `date`, `builtBy` via ldflags).
- dprint as canonical formatter for JSON, YAML, Markdown, and Dockerfile.
- GitHub Actions supply-chain hardening: every third-party action pinned to a full commit SHA.

### Changed

- Upgraded `go-finding` to v1.10.0 (pipeline v1.9.2) — branded finding types, SARIF API, fix-strategy reporting.
- Upgraded Go toolchain to 1.26.7.
- `vendor/` directory removed from git tracking; now gitignored and regenerated locally (via `go mod vendor` or automatically by `go build`).
- golangci-lint baseline achieved: **0 issues** across all linters (was ~116 issues). Config includes depguard allow-list for actual dependencies, varnamelen exemptions for idiomatic Go short names, and tagliatelle `json: snake` for config output.
- `internal/cli` test coverage increased from 74.2% to 82.7%.
- Nix flake simplified from 196 to 122 lines while keeping reproducible builds.
- CI installs oxlint via pnpm instead of npm.
- Repository went public after full de-privatization — no `GOPRIVATE` or SSH keys needed anywhere (build, CI, nix).

### Fixed

- Documentation accuracy pass: README profile/detection tables aligned with actual oxlint categories, FEATURES.md and TODO_LIST.md corrections, CONTRIBUTING.md transitive-dependency clarification, and help-text/comment fixes (`configure` help no longer claims `TypeScript` is a severity category; `profileSpecs` comments match the table).
- Release pipeline: nix tap upload disabled (no `nur-packages` repository exists); release notes now advertise install paths that actually work (go install, flake run, verified binary downloads).

## [0.4.0] - 2026-05-17

- Tagged 2026-05-17 on the same commit as v0.3.0. No changelog was recorded at the time; the delta was release-engineering only (GoReleaser pipeline, CI signing).

## [0.3.0] - 2026-05-17

- Tagged 2026-05-17. No changelog was recorded at the time; see `git log v0.2.1..v0.3.0` for details.

## [0.2.1] - 2026-07-22

### Fixed

- Restriction denylist prevents auto-enabling rules that ban fundamental modern JS/TS syntax (`oxc/no-async-await`, `oxc/no-optional-chaining`, `oxc/no-rest-spread-properties`) — these are now forced `off` in all profiles, even `maximal-typesafe`

### Changed

- Migrated `errors.As` to `errors.AsType` for `ExitError` check (Go 1.26+)
- Modernized Go dependencies, aligned with go-finding v1.2.1
- Hardened CI workflows: cancel-in-progress, paths-ignore, timeout-minutes
- Enforced LF line endings repository-wide

## [0.2.0] - 2026-06-15

### Added

- go-finding integration: unified static analysis model with branded types, pipeline submodule, SARIF output
- `analyze` command with go-finding pipeline (Metrics, Retry, Callbacks, structured FindingError)
- Version metadata injection at build time (`version`/`commit`/`date`/`builtBy` via ldflags)
- govulncheck security scanning CI job
- Data-driven `profileSpecs` policy table as single source of truth for severity decisions
- `marshalConfigJSON` helper for shared config serialization
- Project changelog and enforced LF line endings

### Changed

- Migrated to `encoding/json/v2` (`GOEXPERIMENT=jsonv2`) across entire codebase
- Upgraded go-finding from v0.4.2 through v1.2.0 with branded types (`RuleName`/`ToolName`/`ID`/`FilePath`)
- Modernized golangci-lint config to v2 format with enhanced linter settings
- Modernized Nix flake with systems input, fileset source filtering, git-derived versioning
- Replaced justfile with direct Go/nix commands
- Enriched CLI error messages with contextual parameters

### Fixed

- Eliminated data race in parallel CLI tests
- Fixed byte offset in finding Position from oxlint spans
- Restored committed `vendor/` directory for nix sandbox builds
- Eliminated profile decision split-brain with data-driven policy table

## [0.1.0] - 2026-01-01

### Added

- Initial release
