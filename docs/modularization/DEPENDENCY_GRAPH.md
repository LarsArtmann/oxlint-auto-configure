# Dependency Graph — oxlint-auto-configure

**Date:** 2026-05-13

---

## Current State (Single Module)

### Production Dependencies (Internal)

```
cmd/oxlint-auto-configure
  └── internal/cli
        ├── pkg/config
        │     ├── pkg/detect
        │     │     ├── pkg/profile
        │     │     │     └── pkg/rule
        │     │     └── pkg/rule
        │     ├── pkg/profile
        │     │     └── pkg/rule
        │     └── pkg/rule
        ├── pkg/detect (see above)
        ├── pkg/diff
        │     └── pkg/config (see above)
        ├── pkg/format (no deps)
        ├── pkg/oxlint
        │     └── pkg/rule
        ├── pkg/profile
        │     └── pkg/rule
        └── pkg/rule (no deps)
```

### Production Dependencies (External)

```
github.com/larsartmann/oxlint-auto-configure
  ├── github.com/spf13/cobra@v1.10.2         ← internal/cli (all cmd files)
  ├── github.com/larsartmann/go-finding@v0.3.0 ← pkg/oxlint, internal/cli/cmd_analyze
  ├── github.com/stretchr/testify@v1.11.1      ← test-only (all _test.go)
  ├── golang.org/x/sync@v0.20.0                ← transitive (go-finding)
  ├── golang.org/x/tools@v0.44.0               ← transitive (go-finding)
  └── gopkg.in/yaml.v3@v3.0.1                  ← transitive (go-finding, testify)
```

### Dependency Matrix (Production)

```
                       │ cli │ config │ detect │ diff │ format │ oxlint │ profile │ rule │
  cmd/oxlint-auto-conf │  ●  │        │        │      │        │        │         │      │
  internal/cli         │     │   ●    │   ●    │  ●   │   ●    │   ●    │    ●    │  ●   │
  pkg/config           │     │        │   ●    │      │        │        │    ●    │  ●   │
  pkg/detect           │     │        │        │      │        │        │    ●    │  ●   │
  pkg/diff             │     │   ●    │        │      │        │        │         │      │
  pkg/format           │     │        │        │      │        │        │         │      │
  pkg/oxlint           │     │        │        │      │        │        │         │  ●   │
  pkg/profile          │     │        │        │      │        │        │         │  ●   │
  pkg/rule             │     │        │        │      │        │        │         │      │
```

### External Dep Heat Map

```
                       │ cobra │ go-finding │ testify │ stdlib-only │
  internal/cli         │  ●    │     ●      │  (test) │             │
  pkg/oxlint           │       │     ●      │  (test) │             │
  pkg/rule             │       │            │  (test) │     ●       │
  pkg/profile          │       │            │  (test) │     ●       │
  pkg/detect           │       │            │  (test) │     ●       │
  pkg/config           │       │            │  (test) │     ●       │
  pkg/diff             │       │            │  (test) │     ●       │
  pkg/format           │       │            │  (test) │     ●       │
```

---

## Proposed State (3 Modules)

### Module Boundaries

```
┌─────────────────────────────────────────────────────────────┐
│  CLI Module (root)                                          │
│  github.com/larsartmann/oxlint-auto-configure               │
│                                                             │
│  ┌──────────────────────┐  ┌─────────────────────────────┐ │
│  │  cmd/oxlint-auto-    │  │  internal/cli               │ │
│  │    configure         │  │  cmd_root, cmd_configure,   │ │
│  │                      │  │  cmd_analyze, cmd_validate,  │ │
│  │                      │  │  cmd_report                  │ │
│  └──────────────────────┘  └─────────────────────────────┘ │
│                                                             │
│  External: cobra, go-finding (via analysis)                 │
└────────────────────────┬──────────────┬─────────────────────┘
                         │              │
            ┌────────────▼──┐    ┌──────▼────────────────────┐
            │  Core Module  │    │  Analysis Module           │
            │  ./core       │    │  ./analysis                │
            │               │    │                            │
            │  rule/        │    │  oxlint/                   │
            │  profile/     │    │    detector.go             │
            │  detect/      │    │    version.go              │
            │  config/      │    │    fix.go                  │
            │  diff/        │    │    errors.go               │
            │  format/      │    │                            │
            │               │    │  External: go-finding      │
            │  External:    │    │  Internal: core             │
            │    (none)     │    │                            │
            └───────────────┘    └────────────────────────────┘
```

### Proposed Dependency Flow

```
CLI ──► Core
CLI ──► Analysis ──► Core
```

Single direction. No cycles. No upward dependencies from Core.

### Proposed External Dependency Isolation

| Module | Production External Deps |
|--------|------------------------|
| Core | **None** (stdlib only) |
| Analysis | `go-finding` |
| CLI | `cobra` |

`go-finding` is isolated to the Analysis module. Consumers who only need the core rule engine (profiles, config generation, detection) can import the Core module with zero heavy dependencies.

### Test Cross-Module Dependencies

| Module | Tests import from |
|--------|------------------|
| Core | Core only |
| Analysis | Core + Analysis |
| CLI | Core + Analysis + CLI |

All follow the DAG direction. No bidirectional test deps.

---

## Coupling Analysis

### Tightest Coupling Points

1. **`pkg/rule` → consumed by 6 other packages** — This is the foundational type layer. Correctly placed as the leaf of the graph. Moving it to core is the natural choice.

2. **`pkg/config` → consumed by 4 packages (including diff)** — The OxlintConfig type is the central domain model. Core module is the right home.

3. **`pkg/oxlint` → depends only on `pkg/rule`** — Clean seam. Only external dep is go-finding. Perfect analysis module candidate.

4. **`pkg/format` → zero internal deps** — Pure rendering. Could go anywhere. Placing in core keeps format views co-located with the types they render.

### No Coupling Concerns

- No circular dependencies
- No god-packages (all packages < 15 files)
- No test-only production deps leaking
- No bidirectional dependencies

---

## Size Metrics

| Package | Files | Lines (prod) | Lines (test) | Exported Symbols |
|---------|-------|--------------|--------------|-----------------|
| `pkg/rule` | 3 | ~390 | ~334 | ~30 |
| `pkg/profile` | 1 | ~272 | ~227 | ~12 |
| `pkg/detect` | 1 | ~250 | ~216 | ~12 |
| `pkg/config` | 3 | ~343 | ~377 | ~8 |
| `pkg/diff` | 1 | ~228 | ~141 | ~8 |
| `pkg/format` | 1 | ~173 | ~157 | ~5 |
| `pkg/oxlint` | 4 | ~383 | ~635 | ~10 |
| `internal/cli` | 5 | ~862 | ~423 | ~2 |
| `cmd/...` | 1 | ~12 | 0 | 0 |
| **Total** | **20** | **~2913** | **~2510** | **~87** |

**Core module:** ~1656 prod lines, 6 packages, ~75 exported symbols
**Analysis module:** ~383 prod lines, 1 package, ~10 exported symbols
**CLI module:** ~874 prod lines, 2 packages, ~2 exported symbols
