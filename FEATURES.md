# Features

Honest inventory of what oxlint-auto-configure does and how complete it is.

Status legend:

- **FULLY_FUNCTIONAL** — Code present, exercised, and passing tests.
- **PARTIALLY_FUNCTIONAL** — Ships but has known gaps, edge-case bugs, or missing pieces.
- **BROKEN** — Code exists but does not work / is disabled / fails.
- **PLANNED** — Designed or documented but no code exists yet.

## Core Features

| Feature                                                      | Status           | Evidence                                                                                 | Notes                                                                                                                                                                                                                               |
| ------------------------------------------------------------ | ---------------- | ---------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `configure` command — generate `.oxlintrc.json`              | FULLY_FUNCTIONAL | `internal/cli/cmd_configure.go`, `internal/cli/commands_test.go`, `pkg/config/*_test.go` | Supports `--profile`, `--config`, `--dry-run`, `--fix`, `--root`.                                                                                                                                                                   |
| `analyze` command — run oxlint via go-finding pipeline       | FULLY_FUNCTIONAL | `internal/cli/cmd_analyze.go`, `pkg/oxlint/detector.go`                                  | Defaults to summary; supports `json`, `report`, `sarif`, `table`. Pipeline runs `DryRun=true`.                                                                                                                                      |
| `validate` command — validate existing `.oxlintrc.json`      | FULLY_FUNCTIONAL | `internal/cli/cmd_validate.go`, `pkg/config/validate.go`, `pkg/config/validate_test.go`  | Checks unknown rules and invalid severities; advisory drift check against a fresh generate run (`--fail-on-drift` promotes to error for CI).                                          |
| `report` command — list all rules and recommended severities | FULLY_FUNCTIONAL | `internal/cli/cmd_report.go`, `pkg/format/format.go`                                     | Supports `table`, `json`, `summary`.                                                                                                                                                                                                |
| Profile-driven severity decisions                            | FULLY_FUNCTIONAL | `pkg/profile/profile.go`, `pkg/profile/profile_test.go`                                  | Single `profileSpecs` map is the source of truth.                                                                                                                                                                                   |
| Project type detection                                       | FULLY_FUNCTIONAL | `pkg/detect/detector.go`, `pkg/detect/detector_test.go`                                  | Detects React, Next.js, Vue, Jest, Vitest, Node, TypeScript.                                                                                                                                                                        |
| Embedded 841-rule registry                                   | FULLY_FUNCTIONAL | `pkg/rule/registry.go`, `pkg/rule/rules_data.json`, `pkg/rule/registry_test.go`          | Loaded from embedded JSON; `TestRegistryTotal` verifies 841.                                                                                                                                                                        |
| `.oxlintrc.json` generation and round-trip                   | FULLY_FUNCTIONAL | `pkg/config/generator.go`, `pkg/config/configure.go`, `pkg/config/*_test.go`             | Uses `encoding/json/v2` with `jsontext` indentation.                                                                                                                                                                                |
| Crash-durable config writes (atomic write)                   | FULLY_FUNCTIONAL | `internal/cli/cmd_configure.go:179` (`atomicwrite.Write`)                                | Temp + fsync + atomic rename via `go-atomic-write` v0.4.0. A crash mid-write cannot truncate the config.                                                                                                                            |
| Config before/after diffing                                  | FULLY_FUNCTIONAL | `pkg/diff/differ.go`, `pkg/diff/differ_test.go`                                          | Compares plugins, categories, rules, env, settings.                                                                                                                                                                                 |
| Restriction denylist                                         | FULLY_FUNCTIONAL | `pkg/profile/profile.go:81-85`, `pkg/profile/profile_test.go`                            | Forces `oxc/no-async-await`, `oxc/no-optional-chaining`, `oxc/no-rest-spread-properties` to `off` in every profile.                                                                                                                 |
| BuildFlow integration (toolsdk Spec provider)                | FULLY_FUNCTIONAL | `pkg/provider/provider.go`, `pkg/provider/provider_test.go`                              | Self-registers via `toolsdk.Register`; the full lifecycle (missing-only Detect, never-overwrite dry-run-aware Repair, advisory drift HealthCheck) is derived by the SDK's `BootstrapProviderFromSpec`. BuildFlow consumes via blank import. |
| External JS plugins (`@shadcn/lint`)                         | FULLY_FUNCTIONAL | `pkg/rule/external.go`, `pkg/config/preserve.go`, `internal/cli/shadcn_e2e_test.go`      | Detects `@shadcn/lint` deps, registers under `jsPlugins` (oxlint >= 1.80, warn below), never enables `shadcn/*` rules, preserves existing setup verbatim on regeneration.                                                           |
| CLI global flags (`--verbose`, `--quiet`, `--version`)       | FULLY_FUNCTIONAL | `internal/cli/cmd_root.go`, `internal/cli/commands_test.go`                              | Version metadata injected via ldflags.                                                                                                                                                                                              |
| Nix build and dev shell                                      | FULLY_FUNCTIONAL | `flake.nix`                                                                              | `nix develop .` (dev shell), `nix build .`, and `nix flake check .` all pass. `GOEXPERIMENT=jsonv2` set automatically.                                                                                                              |
| CI pipeline (test, security, lint, nix)                      | FULLY_FUNCTIONAL | `.github/workflows/ci.yml`                                                               | Four jobs: test (with `go mod tidy` consistency check), govulncheck security, golangci-lint (pinned `v2.12.2`), `nix flake check`. oxlint installed via npm.                                                                        |

## Output Formats

| Feature                                     | Status           | Evidence                                              | Notes                                            |
| ------------------------------------------- | ---------------- | ----------------------------------------------------- | ------------------------------------------------ |
| Summary output (human-readable)             | FULLY_FUNCTIONAL | `pkg/format/format.go`, `pkg/format/format_test.go`   | Used by `analyze` default.                       |
| JSON output (flat findings)                 | FULLY_FUNCTIONAL | `pkg/format/format.go`, `internal/cli/cmd_analyze.go` | Used by `analyze -f json`.                       |
| Report output (full go-finding Report JSON) | FULLY_FUNCTIONAL | `internal/cli/cmd_analyze.go`                         | Used by `analyze -f report`.                     |
| SARIF output                                | FULLY_FUNCTIONAL | `internal/cli/cmd_analyze.go`                         | Uses `finding.ToSARIFWithOpts`.                  |
| Table output (Markdown)                     | FULLY_FUNCTIONAL | `pkg/format/format.go`, `internal/cli/cmd_report.go`  | Used by `report` default and `analyze -f table`. |

## Quality & Tooling

| Feature                                          | Status           | Evidence                                                    | Notes                                                                                                                            |
| ------------------------------------------------ | ---------------- | ----------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Race-safe tests                                  | FULLY_FUNCTIONAL | `go test -race ./...` passes.                               | `GOEXPERIMENT=jsonv2` required.                                                                                                  |
| golangci-lint clean (CI)                         | FULLY_FUNCTIONAL | `.golangci.yml`, `.github/workflows/ci.yml`                 | CI and local both report 0 issues. Version pinned to `v2.12.2` for reproducibility.                                              |
| govulncheck security scanning                    | FULLY_FUNCTIONAL | `.github/workflows/ci.yml` security job.                    | Runs on every push/PR.                                                                                                           |
| Vendored dependencies                            | FULLY_FUNCTIONAL | `go.mod`                                                    | `vendor/` is gitignored and regenerated locally (`go mod vendor`) or by `go build`; required for nix sandbox builds.             |
| `GOEXPERIMENT=jsonv2` plumbing                   | FULLY_FUNCTIONAL | `flake.nix`, `.github/workflows/ci.yml`, `AGENTS.md`        | Set in nix package, dev shells, CI.                                                                                              |
| Version metadata at build time                   | FULLY_FUNCTIONAL | `internal/cli/cmd_root.go`, `.goreleaser.yaml`, `flake.nix` | Goreleaser and nix both inject `version`, `commit`, `date`, `builtBy` via ldflags.                                               |
| Entry-point test coverage                        | FULLY_FUNCTIONAL | `cmd/oxlint-auto-configure/main_test.go`                    | `--version`, `--help`, and binary-build tests.                                                                                   |
| E2E round-trip test (configure → parse → verify) | FULLY_FUNCTIONAL | `internal/cli/e2e_test.go`                                  | Generates a config and re-parses it via `config.FromJSON`. A three-command chain (configure → validate → report) is not covered. |

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

## Known Gaps (tracked in TODO_LIST.md / ROADMAP.md)

- Embedded registry pinned to oxlint `1.73.0`; newer oxlint releases (1.82.x) trigger the version-mismatch warning on every run. Refresh pending (`oxlint -f json --rules` + `TestRegistryTotal`).
- CI installs unpinned oxlint (`npm install -g oxlint`); an upstream rule addition can break `TestRegistryTotal` without warning.
- `analyze` is untested against configs that register `jsPlugins`; a registered-but-uninstalled plugin makes oxlint fail to load and the pipeline's error surfacing is unknown.
- `report`/`analyze` have no external-plugin awareness (stats cover the embedded registry only).
- `internal/cli` coverage is 82.7%; the remaining gap is oxlint-integration code (`runAnalyze`, `runFixIfNeeded`).
- Docker image is distroless with the binary only — no oxlint inside, so `analyze` and `--fix` cannot run in the container (`configure` works and warns).
- Release-hardening backlog: no `scripts/pre-release-check.sh`, no disabled-workflow canary, no container image signing/SLSA, GoReleaser `brews`/`dockers` deprecation warnings, `anchore/sbom-action@v0` still tag-pinned.
- Community files missing: `SECURITY.md`, issue/PR templates, `CODEOWNERS`; no social preview image.
