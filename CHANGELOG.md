# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `FEATURES.md` and `TODO_LIST.md` for project tracking and honest feature inventory.
- `ROADMAP.md` for long-term project direction and open questions.
- Filled `docs/DOMAIN_LANGUAGE.md` with actual project domain terms.
- `docs/DOMAIN_LANGUAGE.md` includes atomic-write vocabulary (Atomic Write, Crash Durability). Fingerprint and TOCTOU are kept out as implementation details of `go-atomic-write`.
- `go-atomic-write` v0.4.0 dependency — config writes are crash-durable (temp + fsync + atomic rename). v0.4.0 refactored `Write` to take only path+data; the TOCTOU-aware `WriteVerified` is available but unused (overwriting is intended behavior).
- Embedded rules data updated from oxlint `1.59.0` to `1.73.0` (716 → 841 rules, 108 → 113 enabled by default).
- `go-error-family` v0.10.0 transitive dependency (via `go-atomic-write` v0.4.0) — structured error family helpers.
- Entry-point tests for `cmd/oxlint-auto-configure/main.go` (`main_test.go`).
- E2E integration tests: `configure` round-trip via `config.FromJSON` (`e2e_test.go`).
- Atomic-write contract tests verifying no `.tmp` files and idempotent overwrite (`atomic_write_test.go`).
- Coverage tests for `renderFindings`, `printSARIF`, `printReportJSON`, `sortedByPosition`, `resolveConfigPath`, `logDiffIfExisting`, `marshalConfigJSON` (`coverage_test.go`).
- golangci-lint pinned to `v2.12.2` in CI (was `latest`).
- Nix flake check CI job (`nix` job in `.github/workflows/ci.yml`).
- Module consistency check in CI (`go mod tidy` + `git diff --exit-code`).
- Full version metadata in nix builds (`commit`, `date`, `builtBy` via ldflags).
- Sentinel errors for structured error contracts (`ErrUnexpectedVersionOutput`, `ErrOxlintStderr`, CLI validation sentinels).

### Changed

- `writeConfig` (`internal/cli/cmd_configure.go`) now uses `atomicwrite.Write` instead of raw `os.WriteFile` — a crash mid-write can no longer truncate the user's `.oxlintrc.json`.
- Upgraded `go-finding` from v1.2.1 to v1.4.0.
- `vendor/` directory removed from git tracking; now gitignored and regenerated locally (via `go mod vendor` or automatically by `go build`).
- `CONTRIBUTING.md` expanded from a 27-line stub to a comprehensive guide: prerequisites, private dependencies, atomic-write policy, vendorHash workflow, rules update process, and CI overview.
- golangci-lint baseline achieved: **0 issues** across all linters (was ~116 issues). Config includes depguard allow-list for actual dependencies, varnamelen exemptions for idiomatic Go short names, and tagliatelle `json: snake` for config output.
- `internal/cli` test coverage increased from 74.2% to 82.7%.

### Fixed

- `README.md` profile table now reflects actual oxlint categories (removed non-existent `TypeScript` category column; `strict` and `recommended` are now aligned with the code).
- `README.md` project detection table now lists `node` for Vitest alongside `vitest`.
- `README.md` Development commands include `GOEXPERIMENT=jsonv2`.
- `AGENTS.md` dependency versions updated to `go-finding` v1.4.0, `go-atomic-write` v0.4.0, `go-error-family` v0.10.0. Profile description corrected.
- `internal/cli/cmd_configure.go` help text no longer claims `TypeScript` is a severity category.
- `pkg/profile/profile.go` comments now match the actual `profileSpecs` table.
- `README.md` Next.js detection table now includes `react-perf` (was omitted; code enables it via shared React case).
- `FEATURES.md` and `TODO_LIST.md` golangci-lint count corrected from stale ~117 to fresh ~116; linter list fixed (removed `stdversion` which is gopls, not golangci-lint; added `varnamelen`, `mnd`, `tagliatelle`, `err113`).
- `FEATURES.md` Go version references corrected to `1.26.5` (aligned with `go.mod`).
- `FEATURES.md` "Vendored dependencies" evidence column corrected — `vendor/` is gitignored, not a tracked path.
- `CONTRIBUTING.md` `gogenfilter` dependency clarified as transitive (`// indirect` in `go.mod`, pulled in by `go-finding`).
- `FEATURES.md` Nix row restored to FULLY_FUNCTIONAL — `nix flake check` passes; `go build`/`go test`/`go vet`/`golangci-lint` all clean.

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
