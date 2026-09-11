# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `pkg/provider` package — BuildFlow integration via the `go-finding/toolsdk` v1.10.0 Spec contract. A package-level `toolsdk.Register` declares the tool (name `oxlint-auto-configure`, JS/TS trigger, `DependsOn: [oxlint]`), detects a missing `.oxlintrc.json` in recognizable JS/TS projects (`OXLOPT_CONFIG_MISSING` warning), and repairs it by generating the recommended-profile config (honoring BuildFlow's dry-run flag; never overwrites an existing config). BuildFlow consumers blank-import the package and drop their hand-written glue.

### Changed

- Nix flake `go-finding` input bumped from the stale `v1.8.0` pin to `v1.10.0`, matching `go.mod` and the `toolsdk` sub-module's required core version.

### Fixed

- Nothing yet.

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
