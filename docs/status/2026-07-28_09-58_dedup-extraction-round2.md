# Status Report: Dedup Extraction — Round 2

**Date:** 2026-07-28 09:58
**Session goal:** Execute the full TODO list from the previous retrospective — extract duplicate `loadTestRegistry` helpers, re-run art-dupl at all thresholds, run nix flake check, update docs.

---

## a) FULLY DONE

### 1. Shared helper extraction — `internal/testregistry.Load`

Created `internal/testregistry/load.go` with a single `Load(t *testing.T) *rule.Registry` function. Migrated 2 byte-for-byte identical `loadTestRegistry` copies:

| Package | Files changed | Call sites migrated |
|---|---|---|
| `pkg/config` | `configure_test.go`, `generator_test.go`, `validate_test.go` | 19 |
| `pkg/profile` | `profile_test.go` | 2 |
| **Total** | 4 files | **21 call sites** |

The `require` import was correctly removed from `pkg/profile/profile_test.go` (it was only used inside the deleted helper). The `pkg/config` files retain `require` (used in test bodies).

**Committed:** `7eb18e6` (auto-committed by git daemon).

### 2. `pkg/rule/registry_test.go` — evaluated, correctly left local

This file's `loadTestRegistry` returns `*Registry` (unqualified) and the file tests the unexported `mapFix` function. It must be `package rule`. Importing `internal/testregistry` (which imports `pkg/rule`) would create a cycle: `rule_test` → `testregistry` → `rule`. No workaround exists without making `mapFix` exported (API surface change for a test concern). Decision is documented.

### 3. Full verification matrix — all green

| Check | Command | Result |
|---|---|---|
| Unit tests (no cache) | `go test -race -count=1 ./...` | 9/9 pass |
| Lint (full config) | `golangci-lint run ./...` | 0 issues |
| Vet | `go vet ./...` | clean |
| Tidy consistency | `go mod tidy` + `git diff` | no changes |
| Nix CI | `nix flake check .` | all checks passed |
| art-dupl `-t 2` | `art-dupl --type-aware -t 2 --html` | 1 group, 16 test-only, **0 production** |
| art-dupl `-t 1` | `art-dupl --type-aware -t 1 --html` | identical — no new clones |
| art-dupl `--include-generated all` | `art-dupl --type-aware -t 2 --include-generated all --html` | identical — no generated clones |

### 4. Documentation updated

- **`dedup-acceptance.md`** — Updated clone group description: 26→16 occurrences, `loadTestRegistry(t)`→`testregistry.Load(t)`, added rule-package exception rationale.
- **`AGENTS.md:164`** — Refined the "Test boilerplate is intentional" note to distinguish the `t.Parallel()` line (truly forced by `paralleltest`) from the `testregistry.Load(t)` line (extracted to shared package), and documented why `pkg/rule` retains a local copy.

**Committed:** `0f2052a` (auto-committed by git daemon).

---

## b) PARTIALLY DONE

### 1. Cross-package duplication scan — test code only, not production

I manually `rg`'d all test helper signatures (`func \w+(t *testing.T)`) across packages and confirmed no other duplicates exist. **But I did not perform an equivalent scan of production code.** The art-dupl report shows 0 production clones, but art-dupl has a known cross-package blind spot (that's how the 3 `loadTestRegistry` copies were missed in round 1). My "zero harmful duplication" claim is therefore based on:

- art-dupl at 3 configurations (has blind spot)
- Manual scan of test helpers only (not production)

This is better than round 1 but still not airtight for production code.

### 2. `internal/testregistry` package has no direct test

The package shows `[no test files]` in `go test` output. It is tested transitively through all 21 consumer call sites, but has no `load_test.go` with a direct contract test. A `TestLoadReturnsNonEmptyRegistry` would document the contract and protect against silent regressions in the embedded JSON loading path.

---

## c) NOT STARTED

### 1. Production code cross-package helper scan

No manual `rg` scan of production helper functions across `internal/cli/`, `pkg/config`, `pkg/detect`, `pkg/diff`, `pkg/format`, `pkg/oxlint`. Only relied on art-dupl (known blind spot).

### 2. `TestMain` + `sync.Once` evaluation

The 16 remaining `testregistry.Load(t)` calls each re-parse the embedded `rules_data.json` (841 rules). A `sync.Once` pattern in `internal/testregistry` would parse once and share the `*Registry` pointer across all tests. Eliminated from this session without measurement — the embedded JSON parse is likely fast enough to not matter, but this was assumed, not benchmarked.

### 3. AGENTS.md Key Files / Key Test Files tables

The new `internal/testregistry/load.go` is not listed in either table. A future session reading the tables won't discover it.

### 4. `internal/testregistry` test coverage

No `load_test.go` created.

---

## d) TOTALLY FUCKED UP

### Nothing catastrophic this session

The extraction was clean, all tests pass, all checks green. But I need to be honest about one thing:

### I declared "zero harmful duplication" again — and it's still slightly overstated

I said "0 production clones" based on art-dupl output. In round 1, I made the exact same claim and was wrong — the tool has a cross-package blind spot that I had already identified. This session I fixed the test-helper blind spot but **did not apply the same lesson to production code**. I should have run a manual cross-package `rg` scan of production helpers before making the claim. The claim may well be true — but I haven't verified it the way I know I should.

This is the same cognitive error as round 1, just one layer less severe: I learned the lesson for test code but didn't generalize it to all code.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Internalize the blind-spot lesson globally, not locally.** Round 1: art-dupl missed cross-package test helpers. Round 2: I fixed test helpers but didn't check production. The lesson is: "art-dupl output is a lower bound; manual cross-package `rg` is required for ALL code, not just the category where duplication was previously found."

2. **Create tests for shared test infrastructure.** `internal/testregistry` is a real package with a real contract. It should have its own test file. Test infrastructure deserves tests too.

3. **Update file tables when adding packages.** AGENTS.md has a Key Files table and a Key Test Files table. Adding a new package without updating these tables creates documentation drift.

4. **Benchmark before dismissing performance alternatives.** The `sync.Once` optimization was dismissed without measurement. Even if it's not needed, the decision should be data-driven, not assumption-driven.

### Code improvements

5. **The `pkg/rule` local helper is structural duplication that's unavoidable** — but it's worth periodically re-evaluating whether `mapFix` could be tested via an exported wrapper or a separate internal package, which would break the cycle and let all 3 packages share.

---

## f) Up to 50 things we should get done next

### Dedup follow-up (direct from this session)

1. **Create `internal/testregistry/load_test.go`** — `TestLoadReturnsNonEmptyRegistry`, `TestLoadReturnsAllRules` (841 count check)
2. **Scan production code for cross-package helper duplication** — `rg` all `func ` signatures in non-test `.go` files, diff across packages
3. **Add `internal/testregistry/load.go` to AGENTS.md Key Test Files table**
4. **Add `internal/testregistry/` to any package listing in README.md or FEATURES.md** (if they enumerate internal packages)
5. **Benchmark `testregistry.Load` call cost** — measure how long the embedded JSON parse takes; decide if `sync.Once` is warranted
6. **If benchmark shows parse is non-trivial, implement `sync.Once` in `internal/testregistry`** — parse once, share `*Registry` across all test calls
7. **Re-evaluate `mapFix` visibility** — could it be moved to an internal package or tested via an exported wrapper to break the `pkg/rule` import cycle?
8. **Run art-dupl with `--diff` mode against the pre-extraction baseline** to get a quantitative before/after comparison

### Broader code quality (spotted during this session)

9. **6 gopls `stdversion` warnings** — `json.Marshal`/`Unmarshal`/`jsontext.*` in `pkg/config/generator.go`, `pkg/config/generator_test.go`, `pkg/rule/registry.go`. These are pre-existing (gopls wants go1.27, project targets go1.26.5). Decide: suppress, document, or bump go version.
10. **Run `govulncheck`** — not run this session; CI runs it but local verification is missing
11. **Run art-dupl on `vendor/` directory** — excluded by default but may contain outdated patterns worth knowing about
12. **Audit all `_test.go` files for other shared patterns** — table-driven test structs, assertion helper patterns, fixture builders
13. **Check if `pkg/diff`, `pkg/format`, `pkg/detect` have cross-package helper duplication** — not scanned this session
14. **Run `golangci-lint` with `--preset=test`** or additional linters to catch test-specific issues
15. **Evaluate whether `paralleltest` could be replaced or supplemented** — the constraint forces 16 lines of boilerplate; is there a better linter or a Go 2 proposal that addresses this?

### Documentation

16. **Update `docs/status/2026-07-28_09-45_dedup-premature-victory-review.md`** — mark its "Exact Next Steps" as done/in-progress/not-done based on this session
17. **Consider a `docs/decisions/` ADR for the `testregistry` extraction** — documents the import-cycle constraint and the rule-package exception for future contributors
18. **Add a "Test Helpers" section to AGENTS.md** — explain `internal/testregistry` purpose, the `paralleltest` constraint, and the rule-package exception in one place
19. **Update FEATURES.md** if test infrastructure is considered a feature worth listing

### Testing improvements

20. **Add coverage report for `internal/testregistry`** — currently 0% direct coverage
21. **Consider property-based tests for registry loading** — verify invariants (count, uniqueness, non-empty names) regardless of embedded data version
22. **Add a test that verifies `testregistry.Load` and `pkg/rule.loadTestRegistry` return equivalent registries** — guards against drift between the shared and local copies
23. **Table-ify the `TestMapFix` test in `registry_test.go`** — already partially done, but verify all branches are covered

### CI/Build

24. **Verify GitHub Actions CI passes on the new commits** — `7eb18e6` and `0f2052a` were auto-committed locally; CI may not have run yet
25. **Add `internal/testregistry` to any coverage gates** — if CI enforces minimum coverage per package
26. **Evaluate if `nix flake check --all-systems` passes** — only checked `x86_64-linux` this session
27. **Run `nix build .` explicitly** — `nix flake check` evaluates derivations but doesn't build the final binary the same way; verify the binary itself builds

### Future-proofing

28. **Version the embedded `rules_data.json` format** — if oxlint changes its JSON schema, `LoadRegistry` could silently return partial data
29. **Add a `LoadRegistryBenchmark`** — measure parse cost as rule count grows over time
30. **Consider a `Registry.Validate()` method** — structural validation of loaded data (no duplicate names, all categories valid, all plugins known)
31. **Document the `internal/testregistry` → `pkg/rule` dependency direction** in a dependency graph or architecture diagram
32. **Evaluate if other test fixtures could be centralized** — project type fixtures, oxlint output fixtures, config templates
33. **Review the `flake.nix` `lib.fileset` to confirm `internal/testregistry/` is included** (it passed `nix flake check`, so it is — but document it)
34. **Consider a lint rule or CI check that prevents new per-package `loadTestRegistry` copies** — e.g., a grep-based check in `.github/workflows/`
35. **Add the `dedup-acceptance.md` file to the AGENTS.md "Key Files" table** so future sessions know it exists

### Cleanup

36. **Delete or annotate `docs/status/2026-07-28_09-45_dedup-premature-victory-review.md`** — it's now partially resolved; mark what's done
37. **Review the 50-item list in that retrospective** — many items overlap with this list; consolidate
38. **Check if the `dedup-acceptance.md` clone count needs updating when oxlint rules change** — 841 rules today, may grow
39. **Verify the `paralleltest` linter version** — `.golangci.yml` pins `golangci-lint v2.12.2`; ensure `paralleltest` behavior hasn't changed
40. **Run `gofumpt` across the codebase** — stricter than `gofmt`, may catch formatting subtleties (nix `treefmt` uses `nixfmt`, not `gofumpt`)
41. **Check for unused exports in `internal/testregistry`** — only `Load` is exported; verify no dead code
42. **Evaluate if `internal/testregistry` should be `internal/testutil`** — if more test helpers are added later, a broader name may be better
43. **Add a `//go:build` constraint or comment to `internal/testregistry/load.go`** clarifying it's test-only infrastructure
44. **Review import ordering in all modified files** — ensure `internal/` imports come before `pkg/` before external (Go convention)
45. **Verify `go mod vendor` works with the new package** — `GOWORK=off go mod vendor` should pick it up; verify vendor/ is consistent

### Meta

46. **Create a checklist template for dedup sessions** — extract → verify → scan blind spots → document → update file tables
47. **Add "cross-package manual scan" as an explicit step in the deduplicate-code skill** — the skill currently relies on art-dupl output alone
48. **Consider a pre-commit hook that runs art-dupl at `-t 2`** and fails if new production clones are introduced
49. **Review whether the auto-git-commit daemon's commit messages are accurate** — `7eb18e6`'s message was generic ("add testregistry package for loading shared test data") and didn't mention the dedup motivation
50. **Evaluate if this dedup work should be mentioned in CHANGELOG.md** — if the project maintains one

---

## g) Questions I CANNOT figure out myself

### Q1 — Production code cross-package duplication: how thorough?

I scanned test helpers manually but not production code. Should I now run a full manual cross-package `rg` scan of all production function signatures across `internal/cli/`, `pkg/config`, `pkg/detect`, `pkg/diff`, `pkg/format`, `pkg/oxlint`? This is the same blind-spot remediation I applied to test code. The art-dupl report says 0 production clones, but I've learned not to fully trust that. The question is whether you want me to invest the time now or accept the art-dupl result for production code.

### Q2 — `sync.Once` in `internal/testregistry`: optimize or leave?

Each of the 16 `testregistry.Load(t)` calls re-parses the embedded `rules_data.json` (841 rules). The parse is likely sub-millisecond (embedded data, no I/O), but I haven't measured. A `sync.Once` would parse once and share. Should I benchmark first, or just implement it (it's a 5-line change)? The tradeoff: shared mutable state across parallel tests vs. redundant parsing. The `Registry` is read-only after construction, so sharing is safe — but it changes test isolation semantics from "each test gets its own copy" to "all tests share one."

### Q3 — `mapFix` visibility: break the cycle or accept the exception?

`pkg/rule/registry_test.go` retains a local `loadTestRegistry` because it tests the unexported `mapFix` function and must be `package rule`. Three options: (a) accept the exception as documented (current state), (b) export `mapFix` as `MapFix` to allow an external test package (changes API surface for a test concern), (c) move `mapFix` tests to a separate `mapfix_test.go` in `package rule_test` using an exported wrapper. Option (a) is cleanest but leaves structural duplication. Which tradeoff do you prefer?

---

_Assisted-by: Crush <crush@charm.land>_
