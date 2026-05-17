# DELETE TO IMPROVE — oxlint-auto-configure

**Date:** 2026-05-17
**Total deletion candidates:** 10 files + 30+ config lines
**Estimated lines removed:** ~2,200+ lines
**Risk level:** Low (all candidates are dead code, stale docs, or unused config)

---

## TL;DR — Deletion Priority Matrix

| # | Item | Lines | Impact | Risk | Action |
|---|------|-------|--------|------|--------|
| 1 | 8 stale status reports | ~1,675 | HIGH | Zero | **DELETE** |
| 2 | 1 completed planning doc | ~263 | HIGH | Zero | **DELETE** |
| 3 | 12 dead linters in `.golangci.yml` | 12 | HIGH | Zero | **REMOVE** |
| 4 | 5 dead build tags in `.golangci.yml` | 5 | HIGH | Zero | **REMOVE** |
| 5 | `PUBLIC_OR_PRIVATE.md` | 172 | MEDIUM | Zero | **DELETE** |
| 6 | Stale `CHANGELOG.md` | 27 | MEDIUM | Zero | **DELETE** |
| 7 | Dead ldflags (Dockerfile + goreleaser) | 6 | MEDIUM | Low | **FIX** (add vars or remove ldflags) |
| 8 | Stale `.gitignore` entries | 3 | LOW | Zero | **REMOVE** |
| 9 | Dead goreleaser `darwin/386` ignore | 2 | LOW | Zero | **REMOVE** |
| 10 | Redundant `justfile` `vet` in `check` | 1 | LOW | Zero | **REMOVE** |

---

## 1. DELETE: 8 Stale Status Reports (1,675 lines)

**Path:** `docs/status/`
**Files:**

| File | Lines | Why It's Dead |
|------|-------|---------------|
| `2026-04-28_17-20_comprehensive-status.md` | 172 | References D1/D3 critical bugs that were fixed. 56 tests — now 86+. Coverage was 56% — now 73-95%. |
| `2026-04-28_18-47_pipeline-integration-and-refactor.md` | 159 | "No CI, no README, no planning docs" — all now exist. Pure historical snapshot. |
| `2026-04-29_21-06_full-project-audit.md` | 213 | Lists 17 functions at 0% coverage — all now tested. "go-finding published?" — yes, v0.2.0. |
| `2026-04-29_22-13_architecture-deepening-complete.md` | 183 | 121 golangci-lint warnings — now zero. Local replace directive — removed. |
| `2026-04-29_23-11_architecture-audit-complete.md` | 228 | All M01-M30 executed. Historical record of completed audit. |
| `2026-04-30_03-07_go-finding-full-utilization.md` | 222 | 21 gaps found, 9 implemented. Remaining gaps tracked in AGENTS.md. |
| `2026-04-30_03-58_nix-installation-and-code-quality.md` | 150 | Nix setup done and current. "Not started" items tracked elsewhere. |
| `2026-04-30_04-41_session-complete-nix-quality-features.md` | 200 | 89.6% coverage, zero lint — all done. |

**Why delete:** These are point-in-time session snapshots from AI-assisted development. Every "not started" item is tracked in AGENTS.md or the modularization docs. Keeping 9 stale status reports creates confusion about what's current. The git history preserves this information forever.

**Keep only:** `docs/status/2026-04-30_05-15_detection-fix-and-output-overhaul.md` — the most recent status doc with the freshest "not started" items. But honestly, even this could go — AGENTS.md is the living document.

---

## 2. DELETE: Completed Planning Doc (263 lines)

**Path:** `docs/planning/2026-04-29_22-48-architecture-audit-improvements.md`

**Why:** Explicitly marked ✅ COMPLETE. All M01-M30 executed across 22 commits. This is a completed checklist — the git log is the audit trail.

**Keep:** `docs/planning/2026-04-29_22-15_coverage-lint-ci-hardening.md` — partially done, still has open items (CI hardening, complexity reduction). But consider folding open items into AGENTS.md and deleting this too.

**Also keep:** `docs/modularization/` (3 files) — this is an active proposal, not yet executed.

---

## 3. REMOVE: 12 Dead Linters from `.golangci.yml`

These linters are enabled but have **zero matching code patterns** in the codebase:

```yaml
# REMOVE — no net/http usage
- bodyclose
- contextcheck
- noctx

# REMOVE — no database/sql usage
- rowserrcheck
- sqlclosecheck

# REMOVE — uses testify, not Ginkgo
- ginkgolinter

# REMOVE — uses slog, not logr/zap/zerolog
- loggercheck
- zerologlint

# REMOVE — no i18n/locale usage
- gosmopolitan

# REMOVE — no protobuf types
- protogetter

# REMOVE — no custom types requiring struct tags
- musttag
```

**Impact:** Faster lint runs. Zero false positives from irrelevant linters. Cleaner config.

---

## 4. REMOVE: 5 Dead Build Tags from `.golangci.yml`

```yaml
# REMOVE — none of these experiment tags are used anywhere
build-tags:
  - goexperiment.arenas
  - goexperiment.goroutineleakprofile
  - goexperiment.jsonv2
  - goexperiment.runtimesecret
  - goexperiment.simd
```

**Why:** Zero usage of any `goexperiment.*` tag in the entire codebase. These were likely copied from a template or another project's config. They slow down lint runs by enabling unnecessary experiment features.

---

## 5. ALSO REMOVE: Empty `formats: {}` from `.golangci.yml`

```yaml
# REMOVE — this is a no-op
output:
  formats: {}
```

---

## 6. DELETE: `PUBLIC_OR_PRIVATE.md` (172 lines)

**Why:** This is a one-time decision document dated 2026-05-04. The decision was made: "make public conditionally." Once a decision is made, the document has zero ongoing value. It clutters the repo root and confuses contributors.

The checklist items are either done or tracked in AGENTS.md. The analysis is preserved in git history.

---

## 7. DELETE: `CHANGELOG.md` (27 lines)

**Why:**
- Contains a **wrong date** (`2026-01-01` — project was created 2026-04-28)
- Has empty boilerplate sections that were never filled in
- `.goreleaser.yaml` already has a `changelog` section that auto-generates from git history
- GoReleaser's auto-generated changelogs will **supersede** this manual file
- A stale, wrong-dated CHANGELOG is worse than no CHANGELOG

**Alternative:** If you want a manual changelog, fix the date and fill in actual changes. But GoReleaser makes this unnecessary.

---

## 8. FIX: Dead Ldflags in Dockerfile + GoReleaser

### The Problem

Both `Dockerfile` and `.goreleaser.yaml` inject these ldflags:
```
-X main.commit=${COMMIT}
-X main.date=${BUILD_DATE}
-X main.builtBy=goreleaser
```

But `cmd/oxlint-auto-configure/main.go` only has `var version = "dev"`. There are **no** `var commit`, `var date`, or `var builtBy` variables. These ldflags silently do nothing.

### Options

**Option A (Recommended):** Add the missing variables to `main.go` and wire them into `--version` output:
```go
var (
    version = "dev"
    commit  = "unknown"
    date    = "unknown"
    builtBy = "unknown"
)
```

**Option B:** Remove the dead ldflags from Dockerfile and `.goreleaser.yaml`. Simpler but loses useful build metadata.

---

## 9. REMOVE: Stale `.gitignore` Entries

```gitignore
# REMOVE — no jscpd tool in this project, no report/ directory
/report/jscpd-report.json

# REMOVE — no coverage/ directory exists (coverage.out is already listed)
/coverage

# REMOVE — this is a Go project; node_modules only relevant in target projects
node_modules/
```

The `jscpd-report.json` entry references a copy/paste detection tool that was never set up in this repo. The `/coverage` directory has never existed. `node_modules/` is only relevant in the target projects being analyzed, not in this tool's own repo.

---

## 10. REMOVE: Dead Goreleaser `darwin/386` Ignore Rule

```yaml
# REMOVE — 386 is not in the goarch list, so this rule can never match
ignore:
  - goos: darwin
    goarch: 386    # ← dead: only amd64 and arm64 are built
```

The `goarch` list only has `amd64` and `arm64`. The `darwin/386` ignore rule is unreachable dead config.

---

## 11. REMOVE: Redundant `vet` in `justfile`

```just
# This is redundant — golangci-lint already runs govet
check: fmt-check vet lint test
```

`golangci-lint run ./...` with `govet` enabled (which it is in `.golangci.yml`) already runs `go vet`. Having a separate `vet` step in `check` doubles the work.

---

## 12. BONUS: `go.sum` Cleanup

Run `go mod tidy` to remove 4-5 stale entries:
- `github.com/cpuguy83/go-md2man/v2` — stale transitive
- `github.com/russross/blackfriday/v2` — stale transitive
- `go.yaml.in/yaml/v3` — only `/go.mod` line, likely stale
- `gopkg.in/check.v1` — old testify test dependency
- `github.com/spf13/pflag v1.0.9` — superseded by v1.0.10

---

## Summary: What Stays

| Item | Why |
|------|-----|
| `docs/modularization/` (3 files) | Active proposal, not yet executed |
| `docs/planning/2026-04-29_22-15_*.md` | Partially done, still has open items |
| `docs/status/2026-04-30_05-15_*.md` | Most recent status (but consider folding into AGENTS.md) |
| `justfile` | Active and used (just remove redundant `vet` from `check`) |
| `.goreleaser.yaml` | Active (just clean dead entries) |
| `Dockerfile` | Active (just fix ldflags) |
| `git-town.toml` | Active |
| `AUTHORS` | Active |
| `flake.nix` + `flake.lock` | Active |
| `AGENTS.md` | The living document — this is the single source of truth |

---

## Execution Plan

1. **Delete stale docs** — `git rm` 8 status files + 1 completed planning doc
2. **Delete root-level stale files** — `git rm PUBLIC_OR_PRIVATE.md CHANGELOG.md`
3. **Clean `.golangci.yml`** — Remove 12 dead linters, 5 build tags, empty `formats: {}`
4. **Clean `.gitignore`** — Remove 3 stale entries
5. **Clean `.goreleaser.yaml`** — Remove dead `darwin/386` ignore
6. **Fix Dockerfile** — Remove dead ldflags or add missing vars
7. **Clean `justfile`** — Remove redundant `vet` from `check`
8. **Run `go mod tidy`** — Clean stale `go.sum` entries
9. **Verify** — `just check` + `nix build .`

---

_Research by Crush <crush@charm.land>_
