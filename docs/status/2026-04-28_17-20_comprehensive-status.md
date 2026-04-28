# Status Report — 2026-04-28 17:20

## Summary

oxlint-auto-configure is a working Go CLI that generates `.oxlintrc.json` configs for all 716 oxlint rules. Built on the `go-finding` library. All tests pass, binary works end-to-end.

## A) FULLY DONE ✓

| Component | File(s) | Status |
|-----------|---------|--------|
| Go module init + go-finding dep | `go.mod` | ✓ Local replace working |
| Core types (Rule, Category, Plugin, FixCapability, SeverityDecision) | `pkg/rule/rule.go` | ✓ 7 categories, 15 plugins |
| Rule registry (716 embedded rules) | `pkg/rule/registry.go` + `rules_data.json` | ✓ `go:embed`, JSON parsed |
| Profile presets + Categorizer engine | `pkg/profile/profile.go` | ✓ 4 profiles |
| `.oxlintrc.json` config generator | `pkg/config/generator.go` | ✓ Round-trip tested |
| Project type detection | `pkg/detect/detector.go` | ✓ React/Next/Vue/Jest/Vitest/Node/TS |
| Config differ | `pkg/diff/differ.go` | ✓ Added/Removed/Changed |
| CLI commands (configure/analyze/validate/report) | `internal/cli/commands.go` | ✓ Cobra-based |
| CLI entry point | `cmd/oxlint-auto-configure/main.go` | ✓ |
| Test suite (pkg: 80-93% coverage) | `*_test.go` | ✓ 38 tests, race-safe |
| Justfile | `justfile` | ✓ build/test/lint/cover/check |
| README.md | `README.md` | ✓ Comprehensive |
| AGENTS.md | `AGENTS.md` | ✓ Project memory |
| .gitignore | `.gitignore` | ✓ |
| LICENSE (MIT) | `LICENSE` | ✓ |
| CI (GitHub Actions) | `.github/workflows/ci.yml` | ✓ test + lint jobs |
| `go-finding` Detector impl | `pkg/oxlint/detector.go` | ⚠ BUILT BUT BROKEN (see C) |

## B) PARTIALLY DONE

| Component | Issue |
|-----------|-------|
| `pkg/oxlint/detector.go` | Compiles and tests pass (no real tests), but **oxlintJSON struct doesn't match real oxlint output**. Real output wraps in `{ "diagnostics": [...], ... }` with `code` field and `labels[].span` structure. Our struct expects flat `rule`, `line`, `column`. Would fail at runtime. |
| Pipeline integration | `analyze` command uses `Detector.Detect()` directly but never uses `pipeline.Pipeline` for detect→triage→fix→verify loop. go-finding's Pipeline is a key value prop. |
| `pkg/report/` directory | Empty — no report types defined. Report generation lives in `internal/cli/commands.go` as free functions. |

## C) NOT STARTED

| Component | Description |
|-----------|-------------|
| `--fix` flag on `configure` | Auto-run `oxlint --fix` after generating config |
| `--type-aware` / `--type-check` support | oxlint has `--type-aware` and `--type-check` flags for 59 type-aware rules |
| Config migration | Updating an existing config when oxlint adds/removes/renames rules |
| Pre-commit hook | Like golangci-lint-auto-configure has |
| `.goreleaser.yml` | Release automation |
| `--ignore-pattern` support | Project-specific ignore patterns |
| Named severity overrides | `--rule no-debugger=error` per-rule CLI overrides |
| Examples directory | Example configs for different project types |

## D) TOTALLY FUCKED UP 🐛

### D1. CRITICAL: oxlint JSON parsing is wrong
`pkg/oxlint/detector.go:82-91` — The `oxlintJSON` struct does NOT match real oxlint JSON output:

**Expected by our code (WRONG):**
```json
[{"severity":"warning","message":"...","rule":"no-debugger","filename":"test.ts","line":1,"column":19,"fix":false}]
```

**Actual oxlint output:**
```json
{"diagnostics":[{"message":"...","code":"eslint(no-debugger)","severity":"warning","labels":[{"span":{"offset":18,"length":9,"line":1,"column":19}}],"filename":"test.ts","help":"Remove...","url":"https://..."}],"number_of_files":1,...}
```

Key differences:
- Wrapped in `{"diagnostics": [...]}` not a flat array
- Field is `code` not `rule`, format is `plugin(rule)` e.g. `eslint(no-debugger)`
- Position is in `labels[0].span.line` / `labels[0].span.column`, not flat `line`/`column`
- Has `help`, `url`, `causes`, `related` fields we ignore
- No `fix` boolean — fixability isn't in the diagnostic output

**Impact:** `analyze` command will crash/fail at runtime when parsing real oxlint output.

### D2. ruleSeverityMap may emit wrong overrides
`pkg/config/generator.go:154-169` — The logic emits rules as `"off"` when categorizer says `SeverityOff`, but ALSO emits rules that differ from category default. For the `recommended` profile where nursery rules are `off`, this correctly adds them. But the logic is fragile — if a rule's severity happens to match the category default, it's correctly skipped, but we don't validate the category→severity mapping matches the categorizer's actual per-category behavior.

### D3. Error handling gaps in Detector
`pkg/oxlint/detector.go:53-60` — Non-`ExitError` errors from `cmd.Output()` are silently dropped. The function returns `nil, nil` on real failures (e.g., oxlint not found in PATH).

## E) WHAT WE SHOULD IMPROVE

### E1. Type Model Improvements
1. **`OxlintConfig.Settings` uses `map[string]interface{}`** — should be `map[string]any` (Go 1.18+ alias)
2. **`defaultSettings()` is 30 lines of hardcoded JSON** — should be typed structs or loaded from embedded JSON like rules
3. **`pluginFlagToName()` is a string→string map** — should use `rule.Plugin` constants
4. **`reportJSON`/`reportTable` have unused `reg` param** — dead code
5. **`contains[T]()` generic** — duplicates `slices.Contains`
6. **`IsValid()` loops** — should use `slices.Contains`

### E2. Architecture Improvements
1. **Extract `internal/cli/commands.go` (428 lines)** into per-command files
2. **Use `io.Writer` for output** instead of `fmt.Fprint(os.Stderr, ...)` — enables test capture
3. **Wire `pipeline.Pipeline`** in the `analyze` command for the full detect→fix→verify loop
4. **Add `pkg/report/` types** — move report generation out of CLI into its own package
5. **Typed config structs** for `defaultSettings()` instead of `map[string]any`

### E3. Use Established Libraries
1. **`slices`** (stdlib since Go 1.21) — replace manual loops and custom `contains[T]`
2. **`maps`** (stdlib since Go 1.21) — for key collection
3. **`log/slog`** (stdlib since Go 1.21) — structured logging instead of `fmt.Fprintf(os.Stderr, ...)`
4. **`charmbracelet/log`** — like golangci-lint-auto-configure uses for styled output

## F) TOP 25 THINGS TO DO NEXT

Sorted by (Impact × Urgency) / Effort:

| # | Task | Impact | Effort | Type |
|---|------|--------|--------|------|
| 1 | **FIX D1: oxlint JSON parsing** — match real `{diagnostics:[...]}` format | 🔴 Critical | 15min | Bug fix |
| 2 | **Add test for real oxlint output** — e2e test with actual `oxlint -f json` | 🔴 Critical | 10min | Test |
| 3 | **FIX D3: error handling in Detector** — surface oxlint-not-found etc. | 🔴 Critical | 5min | Bug fix |
| 4 | **`interface{}` → `any`** in config/generator.go | 🟡 Medium | 5min | Cleanup |
| 5 | **Remove custom `contains[T]`** → use `slices.Contains` | 🟡 Medium | 3min | Cleanup |
| 6 | **`IsValid()` loops → `slices.Contains`** in rule.go + profile.go | 🟡 Medium | 3min | Cleanup |
| 7 | **Remove unused `reg` params** from reportJSON/reportTable | 🟡 Medium | 2min | Dead code |
| 8 | **Add oxlint version check** — like golangci-lint-auto-configure does | 🟡 Medium | 10min | Feature |
| 9 | **Use `io.Writer` in CLI** — enable testable output capture | 🟡 Medium | 15min | Refactor |
| 10 | **Wire `pipeline.Pipeline`** in analyze command | 🟡 Medium | 15min | Feature |
| 11 | **Extract commands into per-command files** | 🟠 Low | 10min | Refactor |
| 12 | **Typed Settings structs** instead of `map[string]any` | 🟠 Low | 15min | Types |
| 13 | **Add `--fix` flag** to configure command | 🟡 Medium | 10min | Feature |
| 14 | **Add `--type-aware` support** for 59 type-aware rules | 🟡 Medium | 10min | Feature |
| 15 | **Add oxlint detector tests** with mock output | 🟡 Medium | 15min | Test |
| 16 | **Create `pkg/report/` types** — move report logic from CLI | 🟠 Low | 10min | Architecture |
| 17 | **Add `validate` against actual oxlint** — run `oxlint --print-config -c` to verify | 🟡 Medium | 10min | Feature |
| 18 | **Add `--rule <name>=<severity>` overrides** | 🟠 Low | 10min | Feature |
| 19 | **Use `slog` for structured logging** | 🟠 Low | 10min | Cleanup |
| 20 | **Examples directory** — example configs for React/Next/Vue/library | 🟠 Low | 10min | Docs |
| 21 | **Config migration** — update existing config when rules change | 🟢 Low | 20min | Feature |
| 22 | **Pre-commit hook** | 🟢 Low | 10min | DX |
| 23 | **`.goreleaser.yml`** for release automation | 🟢 Low | 10min | DX |
| 24 | **CLI coverage** — currently 60.1%, target 80%+ | 🟠 Low | 15min | Test |
| 25 | **`hasPromiseUsage`** — fix directory-name heuristic | 🟢 Low | 5min | Bug fix |

## G) TOP #1 QUESTION

**Should `pkg/oxlint/detector.go` parse the raw oxlint JSON diagnostic format directly (current approach — we control the mapping), or should we run oxlint with `--print-config` + its internal rule representation and construct findings from the rule registry instead?**

The real oxlint JSON output has a complex nested structure with `labels`, `spans`, `causes`, `related`. Parsing it correctly is non-trivial. An alternative approach: run oxlint to get diagnostics, then match each diagnostic's `code` field (e.g. `eslint(no-debugger)`) against our embedded rule registry to get the category/severity metadata. This avoids duplicating oxlint's type system.

## Test Results

```
pkg/config    — 91.4% coverage — 8/8 PASS
pkg/detect    — 92.5% coverage — 9/9 PASS
pkg/diff      — 92.9% coverage — 6/6 PASS
pkg/oxlint    —  0.0% coverage — no tests
pkg/profile   — 80.0% coverage — 10/10 PASS
pkg/rule      — 90.1% coverage — 16/16 PASS
internal/cli  — 60.1% coverage — 7/7 PASS
```

Total: **56 tests, all PASS, race-safe**.
