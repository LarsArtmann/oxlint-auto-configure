# Status Report: Nix Installation + Code Quality Fixes

**Date:** 2026-04-30 03:58  
**Session:** 4 commits ahead of origin/master  
**Tests:** 8/8 packages pass with `-race`  
**Nix:** Build deterministic, `nix run` verified end-to-end

---

## A. FULLY DONE

| #   | Item                                                                              | Commit    |
| --- | --------------------------------------------------------------------------------- | --------- |
| 1   | **Nix flake** — `packages`, `apps` (oxlint-wrapped), `devShells`, `overlays`      | `90d9986` |
| 2   | **Vendored deps** — `vendor/` committed for nix sandbox (private go-finding)      | `90d9986` |
| 3   | **`.gitattributes`** — vendor/ marked `linguist-generated`                        | `90d9986` |
| 4   | **Dev shell** — go, gopls, gotools, golangci-lint, oxlint, just + GOPRIVATE       | `90d9986` |
| 5   | **`nix run`** — wrapped with `makeWrapper` so oxlint is in PATH                   | `90d9986` |
| 6   | **Analyze help text** — added missing `report` format to flag description         | `0859f3e` |
| 7   | **CheckVersion error** — log warning instead of silently discarding               | `0859f3e` |
| 8   | **Deduplicate profileNames** — `profile.AllProfileNames()`, removed 2 copies      | `0859f3e` |
| 9   | **Remove duplicate sort** — `PrintFindingsTable` no longer re-sorts (caller does) | `0859f3e` |
| 10  | **Justfile clean safety** — `trash` with `rm` fallback                            | `4b85b8e` |
| 11  | **`.gitignore`** — added `result` (nix build symlink)                             | `4b85b8e` |
| 12  | **README** — nix install instructions, fix analyze format list, add dev commands  | `a16d2ec` |
| 13  | **Nix build determinism** — verified with `--rebuild`                             | verified  |
| 14  | **End-to-end test** — `nix run . -- configure --dry-run` works, version = git rev | verified  |
| 15  | **Justfile recipes** — `vendor`, `nix-build`, `nix-shell`                         | `90d9986` |

## B. PARTIALLY DONE

| #   | Item                      | Status                                 | What's Missing                                  |
| --- | ------------------------- | -------------------------------------- | ----------------------------------------------- |
| 1   | **Shell completions**     | Removed from flake (non-deterministic) | Could re-add with fixed hash approach           |
| 2   | **AGENTS.md nix section** | Added basic section                    | Could be more detailed about vendoring workflow |

## C. NOT STARTED

| #   | Item                                                                                                   | Impact | Effort  |
| --- | ------------------------------------------------------------------------------------------------------ | ------ | ------- |
| 1   | `--fix` flag on analyze command (enable pipeline fix+verify loop)                                      | High   | Medium  |
| 2   | Wire `pkg/oxlint/fix.go` into pipeline or decide vs go-finding `FixApplier`                            | High   | Medium  |
| 3   | Test for `printReportJSON` (new `report` format has no dedicated test)                                 | Medium | Low     |
| 4   | Parse oxlint `related` field (G6 from prior audit)                                                     | Medium | Low     |
| 5   | `interface{}` → `any` in `commands_test.go:185` (LSP hint)                                             | Low    | Trivial |
| 6   | `go.mod` has `golang.org/x/sync` and `golang.org/x/tools` as indirect — only used by go-finding vendor | Low    | Trivial |
| 7   | Add `--verbose`/`--quiet` flags to analyze and configure root commands                                 | Low    | Low     |
| 8   | CI pipeline (GitHub Actions) with nix build + test                                                     | Medium | Medium  |
| 9   | `nix flake check` integration                                                                          | Medium | Low     |
| 10  | Release automation (goreleaser or nix-based)                                                           | Medium | Medium  |

## D. TOTALLY FUCKED UP

Nothing in this session. Clean execution, all tests pass, nix build is deterministic.

**Prior session issue (now resolved):** Shell completion generation in nix was non-deterministic. Fixed by removing it.

## E. WHAT WE SHOULD IMPROVE

### Architecture & Type Model

1. **`OxlintConfig` uses raw `map[string]string` for categories/rules** — Could use typed domain types (`map[rule.Category]rule.SeverityDecision`) internally and only convert to `map[string]string` at JSON boundary. Would eliminate string-typed severity/category comparisons scattered through `generator.go` and `validate.go`.

2. **`FindingView` in `pkg/format` duplicates `finding.Finding` fields** — The `FindingView` struct manually copies 11 fields. Could use a `finding.Finding` → `FindingView` converter in `pkg/format` that's auto-derived, or embed the finding directly.

3. **`PluginConfig` is `map[rule.Plugin]bool`** — Works fine but doesn't carry metadata (why enabled, source of detection). A small struct `{Enabled bool; Source string}` would improve debuggability.

4. **`detect.Detector` does no error reporting** — Silently returns nil on malformed `package.json`, missing `tsconfig.json`, etc. Should log warnings at minimum.

5. **`vendor/` includes test deps** (testify, spew, difflib, yaml.v3) — Not harmful but ~30% of vendor size. Could exclude with `-mod=mod` or trim script.

### Lib Usage

6. **`finding.Filter` not used in any command** — go-finding provides composable `BySeverity`, `ByCategory`, `ByFixStrategy` etc. The analyze command could expose `--severity=error` or `--category=security` filter flags using these.

7. **`finding.GroupBy*` not used** — Could add `--group-by file|severity|category` to analyze output.

8. **`finding.SortBySeverity` not used** — Could add `--sort-by severity|position` flag.

9. **`Report.ToSARIFFiltered` not used** — go-finding supports filtered SARIF (only errors, only unsuppressed, etc.). We use unfiltered `ToSARIF`.

10. **Pipeline `DryRun=true` hardcoded** — The fix+verify loop is completely inert. Need `--fix` flag on analyze.

### Nix

11. **No `nix flake check`** — Should add checks (fmt, vet, lint, test) to `checks` output.
12. **No Darwin ARM build tested** — `supportedSystems` includes `aarch64-darwin` but untested.
13. **`vendorHash = null`** — Works but means vendor dir must always be committed. Alternative: use `vendorHash` with pre-computed hash for better caching.

## F. TOP 25 NEXT ITEMS (sorted by impact × effort)

| #   | Item                                                                               | Impact | Effort | Category |
| --- | ---------------------------------------------------------------------------------- | ------ | ------ | -------- |
| 1   | `--fix` flag on analyze (enable pipeline fix loop)                                 | ★★★★★  | ★★★    | Feature  |
| 2   | GitHub Actions CI: `nix build .` + `go test`                                       | ★★★★   | ★★     | Infra    |
| 3   | Test for `printReportJSON`                                                         | ★★★    | ★      | Quality  |
| 4   | Add `--severity`/`--category` filter flags to analyze (use `finding.Filter`)       | ★★★    | ★★     | Feature  |
| 5   | Typed `OxlintConfig.Categories` (`map[rule.Category]rule.SeverityDecision`)        | ★★★★   | ★★★    | Arch     |
| 6   | `nix flake check` output                                                           | ★★★    | ★      | Infra    |
| 7   | Shell completions in nix (deterministic approach)                                  | ★★     | ★★     | UX       |
| 8   | `interface{}` → `any` in commands_test.go                                          | ★      | ★      | Lint     |
| 9   | Release automation (tag → nix build → GitHub release)                              | ★★★    | ★★★    | Infra    |
| 10  | Add `--sort-by` flag to analyze (use `finding.SortBySeverity`)                     | ★★     | ★      | Feature  |
| 11  | Add `--group-by` flag to analyze (use `finding.GroupBy*`)                          | ★★     | ★★     | Feature  |
| 12  | Trim test deps from vendor/                                                        | ★      | ★      | Cleanup  |
| 13  | Wire `oxlint/fix.go` into pipeline `FixApplier` interface                          | ★★★★   | ★★★★   | Arch     |
| 14  | Error reporting in `detect.Detector` (log warnings)                                | ★★     | ★      | Quality  |
| 15  | Use `Report.ToSARIFFiltered` for filtered SARIF output                             | ★★     | ★      | Feature  |
| 16  | Parse oxlint `related` field into `finding.RelatedRef`                             | ★★     | ★★     | Feature  |
| 17  | `PluginConfig` metadata (why enabled, detection source)                            | ★★     | ★★     | Arch     |
| 18  | Config round-trip property test (generate → parse → compare)                       | ★★★    | ★★     | Quality  |
| 19  | Add `describe` subcommand (explain what a profile does for your project)           | ★★     | ★★     | Feature  |
| 20  | Add `init` subcommand (interactive profile selection)                              | ★★     | ★★★    | Feature  |
| 21  | Benchmark tests for registry loading + config generation                           | ★★     | ★★     | Quality  |
| 22  | Version check: warn if runtime oxlint major differs from embedded                  | ★★     | ★      | Quality  |
| 23  | Pre-commit hook (nix develop + configure --dry-run)                                | ★★     | ★★     | DX       |
| 24  | CONTRIBUTING.md with nix-based dev setup instructions                              | ★★     | ★      | Docs     |
| 25  | Multi-project support (monorepo: detect subprojects, generate per-project configs) | ★★★★★  | ★★★★★  | Feature  |

## G. TOP #1 QUESTION

**Should we commit `vendor/` or switch to `buildGoModule` with a `vendorSha256`?**

Currently we use `vendorHash = null` which tells nix to use the committed `vendor/` dir. This:

- ✅ Works with private deps without Git auth in the nix sandbox
- ✅ No hash update step needed when deps change
- ❌ Adds 1.9MB / 178 files to every commit
- ❌ Clutters git history with vendor updates

Alternative: Use `vendorSha256 = "sha256-..."` (FOD hash) with `go mod download` — but this requires `netrc` / `git credential` config for private repos in the nix sandbox, which is complex and potentially insecure.

**Recommendation:** Keep current approach. The 1.9MB vendor dir is tiny, and `GOWORK=off go mod vendor` + `git add vendor/` is a simple workflow documented in AGENTS.md. The alternative adds significant complexity for minimal benefit.

---

## Session Summary

| Metric                     | Value               |
| -------------------------- | ------------------- |
| Commits                    | 4                   |
| Files changed (non-vendor) | 14                  |
| Lines added                | 198                 |
| Lines removed              | 55                  |
| Tests                      | 8/8 pass            |
| Nix build                  | Deterministic ✅    |
| `nix run`                  | Works end-to-end ✅ |
| Working tree               | Clean               |

_Assisted-by: Crush <crush@charm.land>_
