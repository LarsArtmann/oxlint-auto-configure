# Status Report — 2026-04-30 05:15

**Session Focus:** Detection fixes, analyze output overhaul, honest self-assessment

---

## a) FULLY DONE

| #   | What                                            | Commit            | Impact                                                                                                                            |
| --- | ----------------------------------------------- | ----------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **ProjectTypeTest for vitest/jest**             | `8ca8017`         | HIGH — vitest/jest projects now correctly enable both `PluginNode` AND test plugin                                                |
| 2   | **Analyze summary output overhaul**             | `e2b6e7f`         | HIGH — replaced terrible slog key=value dump with structured: header, severity/category/fix breakdown, top 10 rules, top 10 files |
| 3   | **Iteration logging noise fixed**               | `e2b6e7f`         | MEDIUM — OnIteration moved from `slog.Info` to `slog.Debug`; no more spam unless `-v`                                             |
| 4   | **Config generator recognizes ProjectTypeTest** | `8ca8017`         | MEDIUM — `buildEnv()` adds `node: true` for test projects too                                                                     |
| 5   | **AGENTS.md updated**                           | `6d95d92`         | LOW — documented new gotchas                                                                                                      |
| 6   | **Scope boundary documented**                   | `0fc6ee7`         | MEDIUM — README, AGENTS.md, and CLI help all state "we configure, not lint"                                                       |
| 7   | **Nix build + dev shell**                       | previous sessions | DONE — `nix build .`, `nix run .`, `nix develop .` all work; oxlint wrapped in PATH                                               |
| 8   | **go-finding pipeline integration**             | previous sessions | DONE — full pipeline: detect → triage → report with Metrics, Retry, Callbacks                                                     |
| 9   | **SARIF + JSON + table + report output**        | previous sessions | DONE — all 5 output formats for analyze command                                                                                   |
| 10  | **Severity filter**                             | previous sessions | DONE — `-s/--severity` flag for analyze                                                                                           |

**All 8 packages pass** with `-race`, 0 failures. Coverage: 73–95% across packages.

---

## b) PARTIALLY DONE

| #   | What                                     | Status                                    | Gap                                                                                                       |
| --- | ---------------------------------------- | ----------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| 1   | **Plugin detection for test frameworks** | Works for direct deps                     | Does NOT detect test frameworks used without a package.json dep (e.g., global install, monorepo hoisting) |
| 2   | **Import plugin detection**              | `hasImportUsage()` checks `*.mjs`/`*.mts` | Only root dir, no subdirs. Risk of false positives if expanded naively                                    |
| 3   | **SummaryView.Findings field**           | Added, populated                          | Only used for top-rules/top-files in summary format. Could be used for richer table output                |

---

## c) NOT STARTED

| #   | What                                       | Priority     | Notes                                                                                                                                       |
| --- | ------------------------------------------ | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Migrate CLI to cmdguard v2.2**           | LOW          | Blocked: cmdguard has 31 compile errors (`undefined: FormatTable` etc.). Would give typed flags, DI, lifecycle hooks, signal handling       |
| 2   | **React + test framework cross-detection** | MEDIUM       | React project with vitest should auto-detect both React AND vitest plugins. Currently works via `depPluginRules` but could be more explicit |
| 3   | **Per-rule override documentation**        | LOW          | No way to explain WHY a specific rule is at a specific severity in the generated config                                                     |
| 4   | **Config migration/diff command**          | LOW          | `configure` shows diff but there's no standalone "show me what changed" command                                                             |
| 5   | **Monorepo support**                       | LOW          | Detection assumes single package.json at root                                                                                               |
| 6   | **Watch mode**                             | OUT OF SCOPE | Explicitly documented as "not our job"                                                                                                      |

---

## d) TOTALLY FUCKED UP (Honest Self-Assessment)

| #   | What went wrong                                                                                              | Root cause                                                                                                                | Lesson                                                             |
| --- | ------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| 1   | **Answered the cmdguard question with "it has 31 compile errors" instead of understanding the FULL picture** | Lazy analysis — looked at LSP diagnostics instead of reading cmdguard's code and understanding what it would give us      | Read the WHOLE codebase before giving architecture opinions        |
| 2   | **Didn't immediately see vitest→ProjectTypeNode bug from the paste**                                         | The paste literally showed `by_category=correctness=155` with no test plugin rules. Should have been obvious in 5 seconds | Look at what's MISSING, not just what's there                      |
| 3   | **The analyze output was shipped in this terrible state to begin with**                                      | Previous session built the pipeline but didn't invest in the output format                                                | Ship the experience, not just the plumbing                         |
| 4   | **GOWORK issue not documented earlier**                                                                      | Parent `go.work` at `/home/lars/projects/go.work` has been silently causing confusion                                     | Document environment gotchas IMMEDIATELY                           |
| 5   | **Test coverage gap in `internal/cli` (73.3%)**                                                              | Commands like `report` and `analyze` have thin test coverage; `cmd_report.go` has no direct unit tests                    | Test coverage should be added as part of the feature, not deferred |

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

- **Profile model is too flat** — `Decide()` is a big switch per profile. As rules grow (716→?), per-category decisions become coarse. Consider per-plugin or per-rule overrides within a profile
- **No config for the config generator** — Can't customize severity at category or rule level beyond the 4 hardcoded profiles. Users who want "recommended + restriction at error" are stuck
- **Detector is filesystem-only** — No AST analysis, no tsconfig reading, no `.oxlintrc.json` merging for incremental adoption

### Code Quality

- **`internal/cli` is 5 files doing too much** — `cmd_analyze.go` is 288 lines with render helpers, SARIF printing, view conversion. Should extract `renderFindings`, `printSARIF`, `printReportJSON` to `pkg/format`
- **Magic strings everywhere** — `"summary"`, `"json"`, `"sarif"`, `"table"`, `"report"` are bare strings, not typed enums
- **No error wrapping convention** — Mix of `fmt.Errorf("verb: %w", err)` and bare `fmt.Errorf("verb")` without chain

### DX

- **No `--init` interactive mode** — Users must know profile names. Should prompt: "Detected React+Vitest. Profile? (1) recommended (2) strict (3) maximal-typesafe"
- **No `--explain` flag** — Can't ask "why is `typescript/no-floating-promises` at error?"
- **No `--diff-only` flag** — Can't see what would change without dry-run dumping the entire config

---

## f) Top 25 Things to Do Next

Sorted by impact × effort (high impact + low effort first):

| #   | What                                                                     | Impact | Effort | Category     |
| --- | ------------------------------------------------------------------------ | ------ | ------ | ------------ |
| 1   | Extract render helpers from `cmd_analyze.go` to `pkg/format`             | MEDIUM | Small  | Code quality |
| 2   | Add typed format enum (replace bare `"summary"`, `"json"`, etc.)         | MEDIUM | Small  | Types        |
| 3   | Add `--init` interactive mode to configure command                       | HIGH   | Medium | DX           |
| 4   | Add `--explain` flag to report command (why is rule X at severity Y)     | MEDIUM | Small  | DX           |
| 5   | Add direct unit tests for `cmd_report.go`                                | MEDIUM | Small  | Testing      |
| 6   | Add integration test: React+Vitest project → verify both plugins enabled | HIGH   | Small  | Testing      |
| 7   | Add integration test: Next.js project → verify all 4 plugins enabled     | MEDIUM | Small  | Testing      |
| 8   | Fix cmdguard v2.2 compile errors (31 undefined FormatTable etc.)         | HIGH   | Medium | Dependency   |
| 9   | Migrate CLI to cmdguard v2.2 (typed flags, DI, signal handling)          | HIGH   | Large  | Architecture |
| 10  | Add custom profile support (`--override category:restriction=error`)     | HIGH   | Medium | Features     |
| 11  | Merge existing `.oxlintrc.json` for incremental adoption                 | HIGH   | Medium | Features     |
| 12  | Add `--diff-only` flag to configure (show changes without writing)       | MEDIUM | Small  | DX           |
| 13  | Add `hasImportUsage` subdirectory scanning                               | LOW    | Small  | Detection    |
| 14  | Add `ProjectTypeReactNative` detection                                   | LOW    | Small  | Detection    |
| 15  | Add `ProjectTypeSvelte` detection                                        | LOW    | Small  | Detection    |
| 16  | Read `tsconfig.json` for stricter TypeScript detection                   | MEDIUM | Medium | Detection    |
| 17  | Add per-plugin severity override in profiles                             | MEDIUM | Medium | Architecture |
| 18  | Add `just lint` to CI (golangci-lint)                                    | LOW    | Small  | CI           |
| 19  | Add GitHub Actions CI workflow                                           | MEDIUM | Small  | CI           |
| 20  | Add `configure --check` (verify config is up-to-date, exit 0/1)          | MEDIUM | Small  | DX           |
| 21  | Add `analyze --watch` — OUT OF SCOPE per AGENTS.md, but highly requested | HIGH   | —      | Rejected     |
| 22  | Add `report --format html` for browsable report                          | LOW    | Medium | Output       |
| 23  | Add `configure --merge` to preserve manual rule overrides                | HIGH   | Large  | Features     |
| 24  | Add version check against latest oxlint release                          | LOW    | Small  | DX           |
| 25  | Document all profiles with rule count breakdown in README                | LOW    | Small  | Docs         |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should we fix cmdguard v2.2 first (31 compile errors in `cli_output.go` / `output.go` — all `undefined: FormatTable`/`FormatCSV`/`FormatYAML` constants) and THEN migrate, or should we stay on raw cobra and invest the cmdguard effort elsewhere?**

The compile errors suggest cmdguard's output format subsystem was either partially implemented or recently refactored. Without understanding cmdguard's `output.go` design intent (is there a `go-output` library that should provide these constants? Was it a renamed package?), I can't judge whether:

- Fixing cmdguard is 1 hour (missing constants) or 3 days (missing subsystem)
- The migration benefits (typed flags, DI) outweigh the maintenance cost of a custom CLI library
- We should instead invest in `--init` / `--explain` / custom profiles which deliver more user value per hour

---

## Metrics Snapshot

| Metric                     | Value                       |
| -------------------------- | --------------------------- |
| Total Go LOC (excl vendor) | 5,351                       |
| Test LOC                   | 2,531                       |
| Embedded rules             | 716 (oxlint 1.59.0)         |
| Packages                   | 8                           |
| Test packages              | 8 (all covered)             |
| Coverage range             | 73.3%–95.8%                 |
| Lowest coverage            | `internal/cli` at 73.3%     |
| Highest coverage           | `pkg/format` at 95.8%       |
| All tests pass with -race  | YES                         |
| Go vet                     | CLEAN                       |
| Working tree               | CLEAN                       |
| Remote                     | UP TO DATE (pushed 6d95d92) |

---

_Generated by Crush — 2026-04-30 05:15_
