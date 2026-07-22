# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

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
