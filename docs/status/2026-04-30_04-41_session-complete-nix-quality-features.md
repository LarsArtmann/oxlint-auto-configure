# Status Report: Nix Installation Complete, Code Quality Hardened, Feature Enrichment

**Date:** 2026-04-30 04:41  
**Branch:** master (pushed to origin)  
**Commits this session:** 11 (pushed)  
**Tests:** 8/8 packages pass with `-race`  
**Coverage:** avg 89.6% (range 73.3%–95.7%)  
**Nix:** Build deterministic, `nix run` end-to-end verified  
**Lint:** Zero issues (go vet, golangci-lint)  
**Working tree:** Clean

---

## A. FULLY DONE

| #   | Item                                                                         | Commit    | Category |
| --- | ---------------------------------------------------------------------------- | --------- | -------- |
| 1   | **Nix flake** — packages, apps (oxlint-wrapped), devShells, overlays         | `90d9986` | Infra    |
| 2   | **Vendored deps** — vendor/ committed for nix sandbox (private go-finding)   | `90d9986` | Infra    |
| 3   | **`.gitattributes`** — vendor/ marked linguist-generated                     | `90d9986` | Infra    |
| 4   | **Dev shell** — go, gopls, gotools, golangci-lint, oxlint, just + GOPRIVATE  | `90d9986` | Infra    |
| 5   | **`nix run`** — wrapped with makeWrapper so oxlint is in PATH                | `90d9986` | Infra    |
| 6   | **Justfile recipes** — vendor, nix-build, nix-shell                          | `90d9986` | DX       |
| 7   | **Deduplicate profileNames** — `profile.AllProfileNames()`, removed 2 copies | `0859f3e` | Quality  |
| 8   | **Analyze help text** — added missing `report` format                        | `0859f3e` | Bug fix  |
| 9   | **CheckVersion error** — log warning instead of silently discarding          | `0859f3e` | Quality  |
| 10  | **Remove duplicate sort** — PrintFindingsTable no longer re-sorts            | `0859f3e` | Quality  |
| 11  | **Justfile clean safety** — trash with rm fallback                           | `4b85b8e` | Safety   |
| 12  | **`.gitignore`** — added `result` (nix build symlink)                        | `4b85b8e` | Infra    |
| 13  | **README** — nix install instructions, fix analyze format list, dev commands | `a16d2ec` | Docs     |
| 14  | **`interface{}` → `any`** — LSP lint fix in commands_test.go                 | `583e886` | Lint     |
| 15  | **Test for printReportJSON** — verifies tool info, summary, findings         | `aa69803` | Testing  |
| 16  | **`--severity` filter flag** — uses `finding.Filter(BySeverityAtLeast)`      | `74276d5` | Feature  |
| 17  | **`ToSARIFFiltered`** — used for SARIF when severity filter is set           | `74276d5` | Feature  |
| 18  | **Detect logging** — warns on malformed package.json                         | `778380c` | Quality  |
| 19  | **ByFixStrategy in summary** — uses `Report.Summary.ByFixStrategy`           | `660c19a` | Feature  |
| 20  | **AGENTS.md** — updated with all new patterns                                | `606da75` | Docs     |

---

## B. PARTIALLY DONE

Nothing. All started items are completed.

---

## C. NOT STARTED

### High Impact

| #   | Item                                                                       | Impact | Effort | Notes                                                                              |
| --- | -------------------------------------------------------------------------- | ------ | ------ | ---------------------------------------------------------------------------------- |
| 1   | `--fix` flag on analyze (enable pipeline fix+verify loop, `DryRun=false`)  | ★★★★★  | ★★★    | Pipeline is wired but inert; all findings have FixStrategy from registry           |
| 2   | Wire `pkg/oxlint/fix.go` into pipeline `FixApplier` or decide architecture | ★★★★   | ★★★★   | go-finding has `FixApplier` interface; oxlint has native `--fix`; need to pick one |
| 3   | GitHub Actions CI: `nix build .` + `go test`                               | ★★★★   | ★★     | Flake is ready; just need `.github/workflows/ci.yml`                               |
| 4   | `nix flake check` output in flake.nix                                      | ★★★    | ★      | Add `checks` to flake outputs                                                      |

### Medium Impact

| #   | Item                                                                        | Impact                                                 | Effort                    | Notes                                             |
| --- | --------------------------------------------------------------------------- | ------------------------------------------------------ | ------------------------- | ------------------------------------------------- | ---------------------------------------- | ---------------------- |
| 5   | Typed `OxlintConfig.Categories` (`map[rule.Category]rule.SeverityDecision`) | ★★★★                                                   | ★★★                       | Would eliminate string-typed severity comparisons |
| 6   | Config round-trip property test (generate → parse → compare)                | ★★★                                                    | ★★                        | Catch serialization drift                         |
| 7   | `--sort-by severity                                                         | position`flag on analyze (use`finding.SortBySeverity`) | ★★                        | ★                                                 | One-liner once `SortBySeverity` is wired |
| 8   | `--group-by file                                                            | severity                                               | category` flag on analyze | ★★                                                | ★★                                       | Use `finding.GroupBy*` |
| 9   | Parse oxlint `related` field into `finding.RelatedRef`                      | ★★                                                     | ★★                        | Schema mapping needed                             |
| 10  | `describe` subcommand (explain what a profile does for your project)        | ★★                                                     | ★★                        | Uses existing `Profile.Description()`             |
| 11  | Release automation (tag → nix build → GitHub release)                       | ★★★                                                    | ★★★                       | goreleaser or nix-based                           |
| 12  | Shell completions in nix (deterministic approach)                           | ★★                                                     | ★★                        | Removed due to non-determinism                    |
| 13  | CONTRIBUTING.md with nix-based dev setup                                    | ★★                                                     | ★                         |                                                   |

### Low Impact / Cleanup

| #   | Item                                                                        | Impact | Effort | Notes                                                              |
| --- | --------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------ |
| 14  | Trim test deps from vendor/ (testify, spew, difflib, yaml.v3)               | ★      | ★      | ~30% of vendor size but harmless                                   |
| 15  | `PluginConfig` metadata struct `{Enabled bool; Source string}`              | ★★     | ★★     | Debugging improvement                                              |
| 16  | `finding.Builder` not used — could simplify detector.go construction        | ★★     | ★      | Existing direct construction is fine                               |
| 17  | Benchmark tests for registry loading + config generation                    | ★★     | ★★     | Performance baseline                                               |
| 18  | Version check: warn if runtime oxlint major differs from embedded           | ★★     | ★      | Semver major comparison                                            |
| 19  | Pre-commit hook (nix develop + configure --dry-run)                         | ★★     | ★★     |                                                                    |
| 20  | `init` subcommand (interactive profile selection)                           | ★★     | ★★★    |                                                                    |
| 21  | Multi-project support (monorepo detection)                                  | ★★★★★  | ★★★★★  | Large scope                                                        |
| 22  | `OxlintConfig` uses `map[string]any` for Settings — could use typed structs | ★★     | ★★     |                                                                    |
| 23  | `FindingView` in `pkg/format` duplicates `finding.Finding` fields           | ★      | ★★     | Minor code smell                                                   |
| 24  | `detect.Detector` has no error return on `Detect()`                         | ★      | ★      | Only returns `(PluginConfig, []ProjectType, error)` with nil error |
| 25  | `internal/cli` coverage at 73.3% — lowest package                           | ★★     | ★★     | Missing: renderFindings edge cases, invalid severity paths         |

---

## D. TOTALLY FUCKED UP

Nothing. Clean session with zero regressions.

---

## E. WHAT WE SHOULD IMPROVE

### Architecture

1. **`--fix` is the biggest gap.** The pipeline wires Metrics, Retry, Callbacks, but `DryRun=true` makes the fix+verify loop completely inert. With FixStrategy now populated from the registry, enabling `DryRun=false` would unlock actual auto-fixing. Decision needed: use go-finding's `FixApplier` or oxlint's native `--fix`.

2. **`OxlintConfig` stringly-typed internals.** `Categories map[string]string` and `Rules map[string]string` should be `map[rule.Category]rule.SeverityDecision` internally, converting to strings only at the JSON boundary. This would make severity comparison type-safe and eliminate `string(d.Severity) != catSev` patterns.

3. **Test coverage for `internal/cli` (73.3%).** Missing coverage is mainly in `renderFindings` error paths and the new `printSARIF` with `minSev` set. Should add a SARIF output test.

4. **`detect.Detector` returns nil error always.** The `Detect()` signature returns `error` but always returns `nil`. Either remove the error return or use it for real errors (e.g., malformed package.json).

### Go-finding Utilization

5. **Unused features still available:**
   - `finding.GroupBy*` — no grouping anywhere
   - `finding.SortBySeverity` — no severity sort option
   - `Report.FindByRule` / `Report.FindByID` — no rule-level queries
   - `Finding.Confidence` / `Related` / `Suppression` / `BeforeCode` / `AfterCode` — not populated
   - `finding.Builder` — not used (direct construction is fine)
   - `pipeline.FixApplier` / `ConflictDetector` / `FileBackup` — not wired

### Nix / Infra

6. **No CI pipeline.** The flake is ready, just needs `.github/workflows/ci.yml`.

7. **No `nix flake check`.** Should add `checks` output that runs fmt-check, vet, test.

8. **Vendor dir is 1.9MB committed.** Works fine but could switch to `vendorSha256` with proper auth setup.

---

## F. TOP 25 NEXT (sorted by impact × effort)

| Rank | Item                                                    | Impact | Effort | Δ Score |
| ---- | ------------------------------------------------------- | ------ | ------ | ------- |
| 1    | `--fix` flag on analyze (enable pipeline fix loop)      | ★★★★★  | ★★★    | 15      |
| 2    | GitHub Actions CI with nix build                        | ★★★★   | ★★     | 8       |
| 3    | `nix flake check` output                                | ★★★    | ★      | 3       |
| 4    | Wire oxlint `--fix` vs go-finding `FixApplier` decision | ★★★★   | ★★★★   | 1       |
| 5    | Test for SARIF output with severity filter              | ★★★    | ★      | 3       |
| 6    | Typed `OxlintConfig.Categories`                         | ★★★★   | ★★★    | 1.3     |
| 7    | Config round-trip property test                         | ★★★    | ★★     | 1.5     |
| 8    | `--sort-by` flag (use `finding.SortBySeverity`)         | ★★     | ★      | 2       |
| 9    | `--group-by` flag (use `finding.GroupBy*`)              | ★★     | ★★     | 1       |
| 10   | Release automation (tag → nix build → GitHub release)   | ★★★    | ★★★    | 1       |
| 11   | Parse oxlint `related` field into `finding.RelatedRef`  | ★★     | ★★     | 1       |
| 12   | CONTRIBUTING.md                                         | ★★     | ★      | 2       |
| 13   | `internal/cli` coverage improvement (73.3% → 85%+)      | ★★     | ★★     | 1       |
| 14   | `describe` subcommand                                   | ★★     | ★★     | 1       |
| 15   | Remove error return from `detect.Detect()` or use it    | ★      | ★      | 1       |
| 16   | Shell completions in nix (deterministic)                | ★★     | ★★     | 1       |
| 17   | Version mismatch warning (major version diff)           | ★★     | ★      | 2       |
| 18   | Benchmark tests                                         | ★★     | ★★     | 1       |
| 19   | Trim test deps from vendor/                             | ★      | ★      | 1       |
| 20   | `PluginConfig` metadata                                 | ★★     | ★★     | 1       |
| 21   | `finding.Builder` adoption                              | ★★     | ★      | 2       |
| 22   | `OxlintConfig.Settings` typed structs                   | ★★     | ★★     | 1       |
| 23   | Pre-commit hook                                         | ★★     | ★★     | 1       |
| 24   | `init` subcommand (interactive)                         | ★★     | ★★★    | 0.7     |
| 25   | Multi-project / monorepo support                        | ★★★★★  | ★★★★★  | 1       |

---

## G. TOP #1 QUESTION

**Should the `--fix` feature use go-finding's `pipeline.FixApplier` or oxlint's native `--fix`?**

go-finding provides:

- `pipeline.FixApplier` interface with `Apply(ctx, finding) error`
- `pipeline.FileBackup` for safe rollback
- Conflict detection via `pipeline.ConflictDetector`
- The pipeline loop: detect → triage → **fix** → verify → repeat

Oxlint provides:

- `oxlint --fix` which auto-fixes all safe fixes in one pass
- Already implemented in `pkg/oxlint/fix.go` (`RunFix`)
- Used by `configure --fix` today

The tension: go-finding's pipeline does per-finding fix+verify with conflict detection. Oxlint does bulk `--fix` in a single pass. The go-finding approach is more robust (conflict detection, iterative verification) but requires implementing `FixApplier` that calls `oxlint --fix` for specific rules. Oxlint's native approach is simpler but doesn't integrate with the pipeline model.

**Recommendation:** Implement a `FixApplier` that calls `oxlint --fix` scoped to specific files/rules. This preserves the pipeline's detect→fix→verify loop while leveraging oxlint's native fix capability.

---

## Session Metrics

| Metric                              | Value                         |
| ----------------------------------- | ----------------------------- |
| Total commits (from origin/master)  | 13 pushed                     |
| Session commits (non-status-report) | 10                            |
| Lines of code (non-vendor)          | 5,237                         |
| Lines of tests                      | 2,501                         |
| Test coverage (avg)                 | 89.6%                         |
| Packages                            | 8 (all pass with -race)       |
| Nix build                           | Deterministic ✅              |
| `nix run`                           | End-to-end verified ✅        |
| Lint                                | Zero issues ✅                |
| Working tree                        | Clean                         |
| Pushed                              | Yes (all 13 to origin/master) |

_Assisted-by: Crush <crush@charm.land>_
