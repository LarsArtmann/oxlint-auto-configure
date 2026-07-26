# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `FEATURES.md` and `TODO_LIST.md` for project tracking and honest feature inventory.
- `ROADMAP.md` for long-term project direction and open questions.
- Filled `docs/DOMAIN_LANGUAGE.md` with actual project domain terms.
- `docs/DOMAIN_LANGUAGE.md` now includes atomic-write vocabulary (Atomic Write, Crash Durability, Fingerprint, TOCTOU).
- `go-atomic-write` v0.3.0 dependency — config writes are now crash-durable (temp + fsync + atomic rename).

### Changed

- `writeConfig` (`internal/cli/cmd_configure.go:179`) now uses `atomicwrite.Write` instead of raw `os.WriteFile` — a crash mid-write can no longer truncate the user's `.oxlintrc.json`.
- Upgraded `go-finding` from v1.2.1 to v1.3.0.
- `vendor/` directory removed from git tracking; now gitignored and regenerated locally (via `go mod vendor` or automatically by `go build`).
- `CONTRIBUTING.md` expanded from a 27-line stub to a comprehensive guide: prerequisites, private dependencies, atomic-write policy, vendorHash workflow, rules update process, and CI overview.
- `go.mod` Go directive adjusted from `go 1.26.5` to `go 1.26.4` to align with the active toolchain.

### Fixed

- `README.md` profile table now reflects actual oxlint categories (removed non-existent `TypeScript` category column; `strict` and `recommended` are now aligned with the code).
- `README.md` project detection table now lists `node` for Vitest alongside `vitest`.
- `README.md` Development commands include `GOEXPERIMENT=jsonv2`.
- `AGENTS.md` dependency version updated to `go-finding` v1.2.1 and profile description corrected.
- `internal/cli/cmd_configure.go` help text no longer claims `TypeScript` is a severity category.
- `pkg/profile/profile.go` comments now match the actual `profileSpecs` table.
- `README.md` Next.js detection table now includes `react-perf` (was omitted; code enables it via shared React case).
- `FEATURES.md` and `TODO_LIST.md` golangci-lint count corrected from stale ~117 to fresh ~116; linter list fixed (removed `stdversion` which is gopls, not golangci-lint; added `varnamelen`, `mnd`, `tagliatelle`, `err113`).
- `FEATURES.md` and `TODO_LIST.md` Go version references corrected from stale `1.26.5` to `1.26.4` (after `go.mod` directive change in `ec08705`).
- `FEATURES.md` "Vendored dependencies" evidence column corrected — `vendor/` is gitignored, not a tracked path.
- `CONTRIBUTING.md` `gogenfilter` dependency clarified as transitive (`// indirect` in `go.mod`, pulled in by `go-finding`).

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
