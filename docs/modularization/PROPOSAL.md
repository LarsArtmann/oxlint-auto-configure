# Modularization Proposal — oxlint-auto-configure

**Date:** 2026-05-13
**Status:** Draft
**Approach:** Greenfield split from single-module monolith

---

## 1. Executive Summary

**Why modularize:** oxlint-auto-configure currently lives in a single `go.mod` with 9 packages. While small, the project has a clear architectural seam between its **domain logic** (rule types, profiles, config generation, detection) and its **integration layer** (oxlint binary, go-finding pipeline, CLI). Splitting along this seam creates:

- Hard compile-time boundaries between domain and infrastructure
- A reusable core library (`pkg/rule`, `pkg/profile`, `pkg/detect`, `pkg/config`) that other tools could import without pulling in `cobra` or `go-finding`
- Independent versioning: the core library can evolve independently of the CLI
- Cleaner dependency story: domain packages have **zero external deps** today

**What changes:** Split into 3 modules — `core` (domain), `analysis` (go-finding integration), and the existing root module (CLI binary). No package moves; only `go.mod` boundaries and import paths change.

**Expected benefits:**

- Core library is publishable and reusable without CLI deps
- `go-finding` dependency isolated to `analysis` module only
- CLI module stays thin — wiring only
- Build times unchanged (project is small), but CI could run core tests independently in the future

---

## 2. Current State Analysis

### 2.1 Module Landscape

| Module                                         | Path        | Internal Deps  | External Deps                    | State    |
| ---------------------------------------------- | ----------- | -------------- | -------------------------------- | -------- |
| `github.com/larsartmann/oxlint-auto-configure` | `./` (root) | All 9 packages | `cobra`, `go-finding`, `testify` | Monolith |

### 2.2 Package Dependency Graph

```
                    ┌─────────────────┐
                    │ cmd/oxlint-auto │
                    │   -configure    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  internal/cli   │  ← imports ALL pkg/ packages
                    └────────┬────────┘
                             │
          ┌──────────┬───────┼───────┬──────────┬──────────┐
          │          │       │       │          │          │
    ┌─────▼────┐ ┌───▼───┐ ┌▼─────┐ │    ┌─────▼────┐ ┌──▼───────┐
    │pkg/config│ │pkg/   │ │pkg/  │ │    │pkg/oxlint│ │pkg/      │
    │          │ │detect │ │diff  │ │    │          │ │profile   │
    └────┬─────┘ └──┬────┘ └──┬───┘ │    └────┬─────┘ └────┬─────┘
         │          │         │     │         │            │
         │    ┌─────▼──┐      │     │         │      ┌─────▼────┐
         ├────┤pkg/    │      │     │         │      │pkg/rule  │
         │    │profile │      │     │         │      │  (leaf)  │
         │    └────┬───┘      │     │         │      └──────────┘
         │         │          │     │         │
         └────┬────▼──────────┘     │         │
              │  pkg/rule  ◄────────┘─────────┘
              │  (leaf)
              └───────────

    ┌──────────────┐
    │  pkg/format  │  ← leaf, no internal deps
    └──────────────┘
```

### 2.3 External Dependency Usage

| Package        | `cobra`            | `go-finding`     | `testify` | stdlib only |
| -------------- | ------------------ | ---------------- | --------- | ----------- |
| `internal/cli` | ✅ (all cmd files) | ✅ (cmd_analyze) | —         | —           |
| `pkg/oxlint`   | —                  | ✅ (detector)    | —         | —           |
| `pkg/rule`     | —                  | —                | —         | ✅          |
| `pkg/profile`  | —                  | —                | —         | ✅          |
| `pkg/detect`   | —                  | —                | —         | ✅          |
| `pkg/config`   | —                  | —                | —         | ✅          |
| `pkg/diff`     | —                  | —                | —         | ✅          |
| `pkg/format`   | —                  | —                | —         | ✅          |

**Key insight:** Only 2 packages need `go-finding`. Only 1 package needs `cobra`. 6 of 8 `pkg/` packages are pure stdlib.

### 2.4 Coupling Hotspots

1. **`internal/cli` is the god-package** — imports all 7 pkg/ packages. This is expected for a CLI orchestrator but means it's the hardest to modularize.
2. **`pkg/diff → pkg/config`** — diff compares configs, so it depends on the config model. This is a natural dependency.
3. **No circular dependencies** — the graph is a clean DAG already.

### 2.5 God-Package Detection

No single package qualifies as a god-package (15+ files or 30+ exported symbols):

- `pkg/rule` has the most exported symbols (~30) but is cohesive — all relate to rule types and registry
- `internal/cli` is the largest by line count but is split across multiple files already (`cmd_*.go`)
- All packages have a clear, single concern

### 2.6 Test Dependencies

| Test File                | Cross-Package Internal Imports          |
| ------------------------ | --------------------------------------- |
| `pkg/config/*_test.go`   | `pkg/detect`, `pkg/profile`, `pkg/rule` |
| `pkg/detect/*_test.go`   | `pkg/profile`, `pkg/rule`               |
| `pkg/profile/*_test.go`  | `pkg/rule`                              |
| `pkg/oxlint/*_test.go`   | `pkg/rule`                              |
| `pkg/diff/*_test.go`     | `pkg/config`                            |
| `internal/cli/*_test.go` | `pkg/config`, `pkg/profile`, `pkg/rule` |

All test dependencies follow the same DAG direction as production code.

---

## 3. Proposed Module Structure

### 3.1 Module Definitions

| Field                    | Module 1: Core                                                                            | Module 2: Analysis                                                                     | Module 3: CLI (root)                            |
| ------------------------ | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | ----------------------------------------------- |
| **Name**                 | `oxlint-core`                                                                             | `oxlint-analysis`                                                                      | `oxlint-auto-configure`                         |
| **Path**                 | `./core`                                                                                  | `./analysis`                                                                           | `./` (root)                                     |
| **Go module**            | `github.com/larsartmann/oxlint-auto-configure/core`                                       | `github.com/larsartmann/oxlint-auto-configure/analysis`                                | `github.com/larsartmann/oxlint-auto-configure`  |
| **Purpose**              | Domain types, rule registry, profiles, project detection, config generation, diff, format | Oxlint binary integration via go-finding                                               | CLI binary with cobra commands                  |
| **Packages**             | `rule`, `profile`, `detect`, `config`, `diff`, `format`                                   | `oxlint`                                                                               | `cmd/oxlint-auto-configure`, `internal/cli`     |
| **Prod deps (internal)** | None                                                                                      | `core`                                                                                 | `core`, `analysis`                              |
| **Prod deps (external)** | None                                                                                      | `go-finding`                                                                           | `cobra`                                         |
| **Test deps (external)** | `testify`                                                                                 | `testify`                                                                              | `testify`                                       |
| **Public API**           | All 6 pkg/ public APIs                                                                    | `Runner`, `Detector`, `NewDetector`, `Option`, `CheckVersion`, `CheckBinary`, `RunFix` | `Configure()`, `ConfigureOptions` (re-exported) |

### 3.2 Proposed Dependency Graph (DAG)

```
  ┌─────────────────────┐
  │   CLI (root module)  │
  │  cmd/ + internal/cli │
  └──────┬────────┬──────┘
         │        │
         │        │
  ┌──────▼──┐  ┌──▼──────────────┐
  │  Core   │  │   Analysis      │
  │ Module  │  │   Module        │
  │         │  │                 │
  │ rule    │  │ oxlint          │
  │ profile │  │ (go-finding)    │
  │ detect  │  └──┬──────────────┘
  │ config  │     │
  │ diff    │◄────┘
  │ format  │
  └─────────┘
```

### 3.3 DAG Verification

- **Core** → zero internal deps ✅
- **Analysis** → depends on Core only ✅
- **CLI** → depends on Core + Analysis ✅
- **No cycles** — Core has no downward deps; Analysis depends up to Core; CLI depends up to both ✅
- **No split brains** — each package lives in exactly one module ✅

---

## 4. Replace / Workspace Strategy

**Recommendation: `go.work` at repo root**

Rationale:

- 3 modules is enough to benefit from a workspace file
- No `replace` directives in any `go.mod` — cleaner
- `go.work` is ignored by consumers when the modules are published
- Developer experience: `go test ./...` just works from repo root

### go.work structure

```go
go 1.26.2

use (
    .
    ./core
    ./analysis
)
```

### go.mod files

**`core/go.mod`:**

```
module github.com/larsartmann/oxlint-auto-configure/core
go 1.26.2
require github.com/stretchr/testify v1.11.1
```

**`analysis/go.mod`:**

```
module github.com/larsartmann/oxlint-auto-configure/analysis
go 1.26.2
require (
    github.com/larsartmann/go-finding v0.3.0
    github.com/larsartmann/oxlint-auto-configure/core v0.0.0
    github.com/stretchr/testify v1.11.1
)
replace github.com/larsartmann/oxlint-auto-configure/core => ../core
```

**Root `go.mod` (simplified):**

```
module github.com/larsartmann/oxlint-auto-configure
go 1.26.2
require (
    github.com/larsartmann/oxlint-auto-configure/core v0.0.0
    github.com/larsartmann/oxlint-auto-configure/analysis v0.0.0
    github.com/spf13/cobra v1.10.2
    github.com/stretchr/testify v1.11.1
)
replace (
    github.com/larsartmann/oxlint-auto-configure/core => ./core
    github.com/larsartmann/oxlint-auto-configure/analysis => ./analysis
)
```

When using `go.work`, the `replace` directives in go.mod are technically redundant (go.work overrides), but keeping them ensures `go mod tidy` works without the workspace file — important for consumers and CI.

---

## 5. Test Dependency Isolation

| Module       | Production go.mod           | Test-only deps                          |
| ------------ | --------------------------- | --------------------------------------- |
| **Core**     | `testify`                   | All deps are test-only except stdlib    |
| **Analysis** | `go-finding`, `core`        | `testify` (also used in analysis tests) |
| **CLI**      | `cobra`, `core`, `analysis` | `testify`                               |

**Note:** `testify` appears in production `require` blocks because Go doesn't distinguish test-only requires in go.mod. This is standard Go practice and not a concern.

**Testhelpers:** Not applicable — no shared testhelpers package exists. Each module's tests are self-contained.

---

## 6. Interface Extraction

No interface extraction is needed for this modularization:

- `pkg/rule` already defines the `Registry` type used by `pkg/oxlint` — this is the natural interface between core and analysis
- `pkg/oxlint.Runner` is already an interface (seam for testing)
- The core module's public API is already the thin surface that analysis and CLI consume

If future consumers need thinner interfaces (e.g., a plugin system), interfaces can be extracted at that time. Premature interface extraction adds complexity without current benefit.

---

## 7. Versioning Strategy

**Recommendation: Root-only versioning**

Rationale:

- This is an internal CLI tool, not a published library
- No external consumers of the core or analysis modules
- Single git tag `vX.Y.Z` at repo root
- Sub-modules use `v0.0.0` with `replace` directives

If the core library is ever published for external consumption:

- Switch to independent semver: `core/v1.0.0`, `analysis/v1.0.0`
- Remove `replace` directives
- Add proper git tag format: `core/v1.2.3`

---

## 8. Migration Strategy

### Ordered Steps (each independently executable)

1. **Create `core/` directory and `go.mod`** — Move `pkg/rule`, `pkg/profile`, `pkg/detect`, `pkg/config`, `pkg/diff`, `pkg/format` into `core/`
2. **Update import paths in core packages** — Replace `github.com/larsartmann/oxlint-auto-configure/pkg/X` with `github.com/larsartmann/oxlint-auto-configure/core/X`
3. **Create `analysis/` directory and `go.mod`** — Move `pkg/oxlint` into `analysis/`
4. **Update import paths in analysis** — Point to `core/` imports
5. **Update root module** — Remove moved packages, add replace directives for `core` and `analysis`
6. **Update `internal/cli` imports** — Point to new module paths
7. **Create `go.work`** — Wire all three modules
8. **Verify build and tests** — `go build ./...`, `go test ./...`, `go vet ./...`
9. **Update `flake.nix`** — Adjust build to handle multi-module structure
10. **Update vendor/** — Re-vendor with new module structure
11. **Update documentation** — AGENTS.md, README.md

---

## 9. Risk Assessment

| Risk                                     | Likelihood | Impact | Mitigation                                                                                           |
| ---------------------------------------- | ---------- | ------ | ---------------------------------------------------------------------------------------------------- |
| Import path breaks across 30+ files      | High       | Medium | Use `sed`/`gofmt` for mechanical replacement; verify with `go build`                                 |
| Vendor directory needs full rebuild      | High       | Low    | `just vendor` already handles this; `GOWORK=off` required                                            |
| flake.nix build breaks                   | Medium     | High   | Update flake to reference new module paths; test with `nix build`                                    |
| Test discoverability changes             | Low        | Low    | `go test ./...` works identically in workspace mode                                                  |
| Circular import after move               | Very Low   | High   | DAG verified in proposal; no cycles exist                                                            |
| go.work interferes with parent workspace | Medium     | Medium | Parent `go.work` at `/home/lars/projects/go.work` already an issue; `GOWORK=off` pattern established |

---

## 10. Build System Impact

### flake.nix Changes

- Build step needs `go build ./cmd/oxlint-auto-configure` (unchanged)
- Vendor needs to include all 3 modules
- `nix run .#test` needs to test all 3 modules
- `makeWrapper` for oxlint binary unchanged

### CI Changes

- None required immediately — workspace mode makes `go test ./...` work as before
- Future: parallel per-module test jobs possible

---

## 11. Key Decisions

1. **3 modules, not 2** — Separating analysis from CLI isolates `go-finding` dep. A 2-module split (core + everything else) would leave `go-finding` in the CLI module, which is acceptable but less clean.

2. **Package paths stay flat** — No `core/pkg/rule`, just `core/rule`. The `pkg/` prefix was organizational; removing it simplifies import paths.

3. **No interface extraction** — Current public APIs are already clean module boundaries. Premature interfaces add complexity.

4. **Root-only versioning** — Tool is internal, no external consumers. Can upgrade later.

5. **`go.work` + `replace`** — Belt and suspenders. `go.work` for developer experience, `replace` for `go mod tidy` without workspace.

---

_Next: Phase 4 — Brutal Self-Review of this proposal_

---

## Self-Review (Phase 4)

### 4.1 Critical Questions Answered

| #   | Question                          | Answer                                                                                                                                                                                                                          |
| --- | --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | What did we forget?               | Embedded resources (`rules_data.json`, `rules_version.txt`) in `pkg/rule` — confirmed they move cleanly with the package. No issue.                                                                                             |
| 2   | What could be better?             | The `pkg/` prefix removal (`core/rule` not `core/pkg/rule`) is a judgment call. Keeping `pkg/` inside core would minimize import path changes but add redundancy. Removing it is cleaner.                                       |
| 3   | What could we still improve?      | Consider whether `pkg/format` belongs in core or analysis. Format views are populated from go-finding types in `cmd_analyze.go`, but the types themselves are pure data structs with no go-finding dependency. Core is correct. |
| 4   | Split brains?                     | None. Each package maps to exactly one module. No types are duplicated across modules.                                                                                                                                          |
| 5   | Granularity right?                | Yes. 3 modules is the minimum meaningful split for this codebase. Further splitting (e.g., rule as its own module) would be over-engineering for a 3K-line project.                                                             |
| 6   | Existing code reuse?              | All packages stay as-is. No new packages created. Only module boundaries change.                                                                                                                                                |
| 7   | Type model improvements?          | No changes needed. Current types are clean.                                                                                                                                                                                     |
| 8   | Leveraging established libs?      | Already using stdlib for core, go-finding for analysis. No improvements needed.                                                                                                                                                 |
| 9   | Replace/workspace strategy works? | `go.work` + `replace` directives verified. Parent workspace at `/home/lars/projects/go.work` is a known issue handled by `GOWORK=off` pattern.                                                                                  |
| 10  | Test deps isolated?               | Yes. `testify` in go.mod is standard Go practice. No test-only production deps.                                                                                                                                                 |
| 11  | CI actually faster?               | Marginal improvement — project is small. Main benefit is dependency isolation, not build speed.                                                                                                                                 |
| 12  | Versioning realistic?             | Root-only versioning is appropriate for an internal tool. No external consumers.                                                                                                                                                |

### 4.2 Key Risk: Import Path Churn

The biggest mechanical risk is updating ~30+ import paths across all files. This is:

- Entirely mechanical (find-and-replace)
- Verifiable by `go build ./...`
- Reversible in a single commit

Mitigation: Use `sed` for bulk replacement, then `go build` to verify.

### 4.3 Key Risk: Vendor Directory

The vendor directory must be regenerated after modularization:

- `just vendor` already handles this with `GOWORK=off`
- All 3 modules' deps must be vendored
- flake.nix uses `vendorHash = null` — needs update after re-vendoring

### 4.4 Revised Recommendation

The proposal stands as-is. No changes needed after self-review. The 3-module split is the right granularity for this project.

### 4.5 Alternative Considered: 2-Module Split

A 2-module split (core + CLI-with-analysis) was considered:

- **Pro:** Fewer modules, simpler structure
- **Con:** `go-finding` stays in the CLI module, defeating the primary benefit of isolation
- **Decision:** 3 modules. The extra module is worth the cleaner dependency story.

### 4.6 Alternative Considered: Keep Monolith

- **Pro:** No work needed, project is small
- **Con:** Misses the opportunity for compile-time enforced boundaries while the project is still small and easy to split
- **Decision:** Split now. The cost is low (~2 hours of mechanical work), the benefit compounds over time.
