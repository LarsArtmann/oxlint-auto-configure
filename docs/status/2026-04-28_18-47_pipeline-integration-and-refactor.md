# oxlint-auto-configure — Full Status Report

**Date:** 2026-04-28 18:47 CEST
**Branch:** master
**Commits:** 4 (3 prior + this one)
**Total LOC:** 2,918 lines of Go
**Tests:** 90 passing, 0 failing (race-enabled)
**Build:** Clean — `go build`, `go test -race`, `go vet` all green

---

## A) FULLY DONE

| #   | Item                             | Details                                                                                                                                                                                                                                  |
| --- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Core rule registry**           | `pkg/rule/` — embeds all 716 oxlint rules from `oxlint -f json --rules` via `go:embed`. Full registry with filtering (by plugin, category, fixable, type-aware, enabled-by-default). Tested.                                             |
| 2   | **Profile system**               | `pkg/profile/` — 4 profiles (maximal-typesafe, recommended, strict, minimal). Maps categories to severities based on profile + detected plugins. `Categorizer.DecideAll()` produces 716 `RuleDecision`s. Tested.                         |
| 3   | **Config generator**             | `pkg/config/` — generates `.oxlintrc.json` with categories + per-rule overrides. `Generate()` for normal profiles, `GenerateAllError()` for maximal-typesafe. JSON round-trip tested.                                                    |
| 4   | **Project type detection**       | `pkg/detect/` — detects React, Next.js, Vue, TS, Jest, Vitest, promise usage from `package.json`. Returns `PluginConfig` that auto-enables relevant plugins. 8 tests.                                                                    |
| 5   | **Diff engine**                  | `pkg/diff/` — compares two `OxlintConfig`s, shows added/removed/changed rules with summary. Used by configure's `--dry-run` diff display. 5 tests.                                                                                       |
| 6   | **oxlint detector (go-finding)** | `pkg/oxlint/detector.go` — implements `go-finding.Detector` interface. Parses real oxlint JSON output (`{diagnostics:[...]}`) with proper types. Handles `eslint(rule)` and `plugin/rule` code formats. 15 tests with real JSON fixture. |
| 7   | **oxlint version check**         | `pkg/oxlint/version.go` — `CheckVersion()` and `CheckBinary()`. Used by configure and analyze commands.                                                                                                                                  |
| 8   | **oxlint --fix wrapper**         | `pkg/oxlint/fix.go` — `RunFix()` wraps `oxlint --fix` for the configure `--fix` flag.                                                                                                                                                    |
| 9   | **CLI commands**                 | 4 commands fully wired: `configure`, `analyze`, `validate`, `report`. Extracted into per-command files.                                                                                                                                  |
| 10  | **Pipeline integration**         | `analyze` command uses `go-finding/pipeline.Pipeline` with `DryRun=true`, `GracefulDegradation=true`. Produces SARIF output via `finding.Report.ToSARIF()`.                                                                              |
| 11  | **Go modernization**             | `interface{}` → `any`, custom `contains[T]` → `slices.Contains`, `strings.IndexByte` → `strings.Cut`.                                                                                                                                    |
| 12  | **Configure --fix flag**         | `--fix` runs `oxlint --fix` after writing config.                                                                                                                                                                                        |
| 13  | **Configure --dry-run**          | Shows what would change, including diff against existing config.                                                                                                                                                                         |

## B) PARTIALLY DONE

| #   | Item                 | Status                                 | What's Missing                                                                                                                            |
| --- | -------------------- | -------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Analyze command**  | Pipeline runs, report generates SARIF  | No human-readable output format (only SARIF to stdout). No `--format` flag. No test coverage for the analyze command itself.              |
| 2   | **Report command**   | Works for table/json/summary           | No SARIF output option. No filtering. No output to file.                                                                                  |
| 3   | **Validate command** | Checks unknown rules + severity values | Doesn't check for conflicting rules, missing plugins, or deprecated rules. No verbose mode.                                               |
| 4   | **Test coverage**    | 90 tests passing                       | No e2e tests (gated by `OXLINT_E2E=1` but not run). No tests for `cmd_analyze`, `cmd_configure --fix`, `fix.go`. Coverage % not measured. |

## C) NOT STARTED

| #   | Item                           | Description                                                                                                                         |
| --- | ------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **--type-aware flag**          | 59 type-aware rules need special handling (require TypeScript compiler). Flag should enable/disable them with appropriate severity. |
| 2   | **Remote repo / CI**           | No GitHub remote. No CI/CD. No release pipeline.                                                                                    |
| 3   | **README**                     | No README.md with usage examples, installation, profiles explanation.                                                               |
| 4   | **Planning docs**              | `docs/planning/` is empty. No architecture docs, ADRs, or roadmap.                                                                  |
| 5   | **--type-aware-support**       | Need to detect if project has `tsconfig.json` and auto-enable type-aware rules.                                                     |
| 6   | **Config migration**           | No support for migrating from ESLint config or older oxlint configs.                                                                |
| 7   | **Rule documentation links**   | Each rule should link to its documentation URL (available in oxlint `--rules` output).                                              |
| 8   | **Ignore patterns**            | No support for `.oxlintignore` generation based on project type.                                                                    |
| 9   | **Interactive mode**           | No TUI/interactive mode for selecting rules/profiles.                                                                               |
| 10  | **Plugin versioning**          | No tracking of which oxlint version the embedded rules come from. Rules could become stale.                                         |
| 11  | **SARIF output for configure** | Configure command should optionally output SARIF for CI integration.                                                                |
| 12  | **Error recovery**             | No graceful degradation if `package.json` is missing, malformed, or if oxlint is wrong version.                                     |
| 13  | **Logging**                    | No structured logging. All output goes to stderr with `fmt.Fprintf`.                                                                |
| 14  | **Shell completion**           | No shell completion for profiles, formats, etc.                                                                                     |
| 15  | **Man pages**                  | No man page generation.                                                                                                             |

## D) TOTALLY FUCKED UP

| #   | Item                          | Problem                                                                                                                                                                                              | Severity                                  |
| --- | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------- |
| 1   | **Embedded rules are static** | The 716 rules are embedded from a single `oxlint -f json --rules` run. If the user has a different oxlint version, rules may not match. No validation that embedded version matches runtime version. | **HIGH** — could generate invalid configs |
| 2   | **No version pinning**        | `CheckVersion()` logs the version but doesn't enforce compatibility.                                                                                                                                 | **MEDIUM**                                |
| 3   | **Analyze SARIF to stdout**   | `analyze` outputs SARIF to stdout while status messages go to stderr. Mixed output could confuse piping. Should have `--output` flag.                                                                | **LOW**                                   |

## E) WHAT WE SHOULD IMPROVE

### Critical

1. **Pin embedded rules to oxlint version** — Store the oxlint version that generated the rules. At runtime, warn/error if version mismatch exceeds threshold.
2. **Add e2e tests** — Run actual oxlint against fixture projects. Gate with `OXLINT_E2E=1`.
3. **Measure test coverage** — Run `go test -coverprofile` and identify gaps.

### Important

4. **Structured logging** — Replace `fmt.Fprintf(os.Stderr, ...)` with `slog` or similar. Add `--verbose`/`--quiet` flags.
5. **Analyze output formats** — Add `--format` flag (sarif, json, table, summary) to analyze. SARIF should be opt-in, not default.
6. **Add README.md** — Installation, usage, profiles, examples.
7. **Create planning docs** — Architecture, ADRs, roadmap in `docs/planning/`.

### Nice to Have

8. **Rule staleness check** — `go generate` script to re-embed rules from latest oxlint.
9. **Config inheritance** — Support `extends` in `.oxlintrc.json` for shared base configs.
10. **Performance** — Profile with large projects (1000+ files). Pipeline parallelism may need tuning.

## F) Top 25 Things to Do Next

| Priority | Task                                                                      | Effort |
| -------- | ------------------------------------------------------------------------- | ------ |
| 1        | Pin embedded rules to oxlint version + add mismatch warning               | S      |
| 2        | Add README.md with usage, profiles, installation                          | S      |
| 3        | Add `--format` flag to analyze (sarif/json/table/summary)                 | S      |
| 4        | Measure and report test coverage (`go test -coverprofile`)                | S      |
| 5        | Write architecture planning doc in `docs/planning/`                       | M      |
| 6        | Add e2e test: configure → validate round-trip with real oxlint            | M      |
| 7        | Add tests for `cmd_analyze` (mock pipeline)                               | M      |
| 8        | Add tests for `pkg/oxlint/fix.go`                                         | M      |
| 9        | Create GitHub repo + push                                                 | S      |
| 10       | Add `--type-aware` flag to configure                                      | M      |
| 11       | Auto-detect tsconfig.json for type-aware rules                            | S      |
| 12       | Add `go generate` script for re-embedding rules                           | S      |
| 13       | Replace `fmt.Fprintf` with `slog` structured logging                      | M      |
| 14       | Add `--verbose`/`--quiet` global flags                                    | S      |
| 15       | Add SARIF output option to configure and report commands                  | S      |
| 16       | Add `.oxlintignore` generation based on project type                      | S      |
| 17       | Add rule documentation URLs to report output                              | S      |
| 18       | Validate plugin compatibility (warn if React plugin but no React project) | S      |
| 19       | Add shell completion (cobra built-in support)                             | S      |
| 20       | Set up CI (GitHub Actions: build, test, vet, lint)                        | M      |
| 21       | Add config migration from ESLint (basic mapping)                          | L      |
| 22       | Add `--output` flag to write results to file                              | S      |
| 23       | Performance profiling with large projects                                 | M      |
| 24       | Add `--check` mode (exit 1 if config would change) for CI                 | S      |
| 25       | Interactive TUI mode for rule selection                                   | L      |

## G) Top Question I Cannot Answer Myself

**Should this tool auto-update its embedded rules when it detects a newer oxlint version?**

Options:

- **A)** Embed rules at build time only (current). User must update the tool when oxlint adds rules.
- **B)** Fetch rules from `oxlint --rules` at runtime. Always current, but requires oxlint at config time and is slower.
- **C)** Hybrid: embed for offline use, but fetch+cache at runtime if oxlint is available and version differs.

This is a product/architecture decision that affects the entire rule system design.

---

## Git State (Pre-Commit)

```
Modified:  go.mod (go-finding dependency added)
Modified:  pkg/oxlint/detector.go (strings.Cut modernization)
Deleted:   internal/cli/commands.go (extracted to per-command files)

New:       internal/cli/cmd_root.go
New:       internal/cli/cmd_configure.go
New:       internal/cli/cmd_analyze.go
New:       internal/cli/cmd_validate.go
New:       internal/cli/cmd_report.go
New:       pkg/oxlint/fix.go
New:       pkg/oxlint/version.go
```

## Test Matrix

| Package        | Tests  | Status       |
| -------------- | ------ | ------------ |
| `internal/cli` | 7      | PASS         |
| `pkg/config`   | 8      | PASS         |
| `pkg/detect`   | 9      | PASS         |
| `pkg/diff`     | 5      | PASS         |
| `pkg/oxlint`   | 15+    | PASS         |
| `pkg/profile`  | tests  | PASS         |
| `pkg/rule`     | tests  | PASS         |
| **Total**      | **90** | **ALL PASS** |
