# Features

Honest inventory of what oxlint-auto-configure does and how complete it is.

Status legend:

- **FULLY_FUNCTIONAL** — Code present, exercised, and passing tests.
- **PARTIALLY_FUNCTIONAL** — Ships but has known gaps, edge-case bugs, or missing pieces.
- **BROKEN** — Code exists but does not work / is disabled / fails.
- **PLANNED** — Designed or documented but no code exists yet.

## Core Features

| Feature                                                      | Status           | Evidence                                                                                 | Notes                                                                                                               |
| ------------------------------------------------------------ | ---------------- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `configure` command — generate `.oxlintrc.json`              | FULLY_FUNCTIONAL | `internal/cli/cmd_configure.go`, `internal/cli/commands_test.go`, `pkg/config/*_test.go` | Supports `--profile`, `--config`, `--dry-run`, `--fix`, `--root`.                                                   |
| `analyze` command — run oxlint via go-finding pipeline       | FULLY_FUNCTIONAL | `internal/cli/cmd_analyze.go`, `pkg/oxlint/detector.go`                                  | Defaults to summary; supports `json`, `report`, `sarif`, `table`. Pipeline runs `DryRun=true`.                      |
| `validate` command — validate existing `.oxlintrc.json`      | FULLY_FUNCTIONAL | `internal/cli/cmd_validate.go`, `pkg/config/validate.go`, `pkg/config/validate_test.go`  | Checks unknown rules and invalid severities.                                                                        |
| `report` command — list all rules and recommended severities | FULLY_FUNCTIONAL | `internal/cli/cmd_report.go`, `pkg/format/format.go`                                     | Supports `table`, `json`, `summary`.                                                                                |
| Profile-driven severity decisions                            | FULLY_FUNCTIONAL | `pkg/profile/profile.go`, `pkg/profile/profile_test.go`                                  | Single `profileSpecs` map is the source of truth.                                                                   |
| Project type detection                                       | FULLY_FUNCTIONAL | `pkg/detect/detector.go`, `pkg/detect/detector_test.go`                                  | Detects React, Next.js, Vue, Jest, Vitest, Node, TypeScript.                                                        |
| Embedded 716-rule registry                                   | FULLY_FUNCTIONAL | `pkg/rule/registry.go`, `pkg/rule/rules_data.json`, `pkg/rule/registry_test.go`          | Loaded from embedded JSON; `TestRegistryTotal` verifies 716.                                                        |
| `.oxlintrc.json` generation and round-trip                   | FULLY_FUNCTIONAL | `pkg/config/generator.go`, `pkg/config/configure.go`, `pkg/config/*_test.go`             | Uses `encoding/json/v2` with `jsontext` indentation.                                                                |
| Crash-durable config writes (atomic write)                   | FULLY_FUNCTIONAL | `internal/cli/cmd_configure.go:179` (`atomicwrite.Write`)                                | Temp + fsync + atomic rename via `go-atomic-write` v0.3.0. A crash mid-write cannot truncate the config.            |
| Config before/after diffing                                  | FULLY_FUNCTIONAL | `pkg/diff/differ.go`, `pkg/diff/differ_test.go`                                          | Compares plugins, categories, rules, env, settings.                                                                 |
| Restriction denylist                                         | FULLY_FUNCTIONAL | `pkg/profile/profile.go:81-85`, `pkg/profile/profile_test.go`                            | Forces `oxc/no-async-await`, `oxc/no-optional-chaining`, `oxc/no-rest-spread-properties` to `off` in every profile. |
| CLI global flags (`--verbose`, `--quiet`, `--version`)       | FULLY_FUNCTIONAL | `internal/cli/cmd_root.go`, `internal/cli/commands_test.go`                              | Version metadata injected via ldflags.                                                                              |
| Nix build and dev shell                                      | FULLY_FUNCTIONAL | `flake.nix`                                                                              | `nix build .` and `nix flake check .` pass. `GOEXPERIMENT=jsonv2` set automatically.                                |
| CI pipeline (test, security, lint)                           | FULLY_FUNCTIONAL | `.github/workflows/ci.yml`                                                               | Three jobs: test, govulncheck security, golangci-lint.                                                              |

## Output Formats

| Feature                                     | Status           | Evidence                                              | Notes                                            |
| ------------------------------------------- | ---------------- | ----------------------------------------------------- | ------------------------------------------------ |
| Summary output (human-readable)             | FULLY_FUNCTIONAL | `pkg/format/format.go`, `pkg/format/format_test.go`   | Used by `analyze` default.                       |
| JSON output (flat findings)                 | FULLY_FUNCTIONAL | `pkg/format/format.go`, `internal/cli/cmd_analyze.go` | Used by `analyze -f json`.                       |
| Report output (full go-finding Report JSON) | FULLY_FUNCTIONAL | `internal/cli/cmd_analyze.go`                         | Used by `analyze -f report`.                     |
| SARIF output                                | FULLY_FUNCTIONAL | `internal/cli/cmd_analyze.go`                         | Uses `finding.ToSARIFWithOpts`.                  |
| Table output (Markdown)                     | FULLY_FUNCTIONAL | `pkg/format/format.go`, `internal/cli/cmd_report.go`  | Used by `report` default and `analyze -f table`. |

## Quality & Tooling

| Feature                                                        | Status               | Evidence                                             | Notes                                                                                                                                                                         |
| -------------------------------------------------------------- | -------------------- | ---------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Race-safe tests                                                | FULLY_FUNCTIONAL     | `go test -race ./...` passes.                        | `GOEXPERIMENT=jsonv2` required.                                                                                                                                               |
| golangci-lint clean (CI)                                       | PARTIALLY_FUNCTIONAL | `.golangci.yml`, `.github/workflows/ci.yml`          | CI passes (0 issues). Local `golangci-lint run ./...` reports ~116 issues (depguard, varnamelen, mnd, tagliatelle, err113, forbidigo) due to version/config mismatch with CI. |
| govulncheck security scanning                                  | FULLY_FUNCTIONAL     | `.github/workflows/ci.yml` security job.             | Runs on every push/PR.                                                                                                                                                        |
| Vendored dependencies                                          | FULLY_FUNCTIONAL     | `vendor/`, `go.mod`                                  | Required for nix sandbox.                                                                                                                                                     |
| `GOEXPERIMENT=jsonv2` plumbing                                 | FULLY_FUNCTIONAL     | `flake.nix`, `.github/workflows/ci.yml`, `AGENTS.md` | Set in nix package, dev shells, CI.                                                                                                                                           |
| Version metadata at build time                                 | PARTIALLY_FUNCTIONAL | `internal/cli/cmd_root.go`, `.goreleaser.yaml`       | Goreleaser injects `version`, `commit`, `date`, `builtBy`. `flake.nix` only injects `version`; `commit`/`date`/`builtBy` remain `unknown` in nix builds.                      |
| Entry-point test coverage                                      | PLANNED              | `cmd/oxlint-auto-configure/main.go`                  | No test files; coverage is 0%.                                                                                                                                                |
| E2E integration test (configure → validate → report roundtrip) | PLANNED              | —                                                    | Not yet implemented.                                                                                                                                                          |

## Documentation

| Feature                   | Status           | Evidence                  | Notes                                   |
| ------------------------- | ---------------- | ------------------------- | --------------------------------------- |
| `README.md`               | FULLY_FUNCTIONAL | `README.md`               | End-user quick start and reference.     |
| `AGENTS.md`               | FULLY_FUNCTIONAL | `AGENTS.md`               | AI session context.                     |
| `CHANGELOG.md`            | FULLY_FUNCTIONAL | `CHANGELOG.md`            | Keeps release notes.                    |
| `docs/DOMAIN_LANGUAGE.md` | FULLY_FUNCTIONAL | `docs/DOMAIN_LANGUAGE.md` | Domain glossary.                        |
| `CONTRIBUTING.md`         | FULLY_FUNCTIONAL | `CONTRIBUTING.md`         | Contributor setup.                      |
| `FEATURES.md` (this file) | FULLY_FUNCTIONAL | `FEATURES.md`             | Honest feature inventory.               |
| `TODO_LIST.md`            | FULLY_FUNCTIONAL | `TODO_LIST.md`            | Short-term open work.                   |
| `ROADMAP.md`              | FULLY_FUNCTIONAL | `ROADMAP.md`              | Long-term direction and open questions. |

## Known Gaps (captured in TODO_LIST.md)

- `go.mod` pins `go 1.26.5`; `encoding/json/v2` triggers gopls `stdversion` warnings (`json.Marshal` requires go1.27).
- No `go mod vendor` consistency check in CI.
- No entry-point tests for `cmd/oxlint-auto-configure` (0% coverage).
- No E2E round-trip test (configure -> validate -> report).
- Embedded rules pinned to oxlint `1.59.0` while runtime is `1.73.0` (produces a `WARN` on every run).
- Local `golangci-lint run ./...` reports ~116 issues (depguard, varnamelen, mnd, tagliatelle, err113, forbidigo) while CI passes (version/config mismatch).
- No dedicated test for the atomic-write contract (no `.tmp` leftovers, valid JSON always).
