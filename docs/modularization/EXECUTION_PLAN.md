# Execution Plan — oxlint-auto-configure Modularization

**Date:** 2026-05-13
**Status:** Draft
**Estimated Total Effort:** 2–3 hours

---

## Task Overview

| #   | Task                               | Tier      | Effort | Depends On |
| --- | ---------------------------------- | --------- | ------ | ---------- |
| 1   | Create core module skeleton        | 1% → 51%  | 15 min | —          |
| 2   | Move core packages to `core/`      | 1% → 51%  | 20 min | 1          |
| 3   | Update core internal imports       | 1% → 51%  | 15 min | 2          |
| 4   | Create analysis module skeleton    | 4% → 64%  | 10 min | 1          |
| 5   | Move oxlint package to `analysis/` | 4% → 64%  | 10 min | 4          |
| 6   | Update analysis imports            | 4% → 64%  | 15 min | 3, 5       |
| 7   | Update root go.mod and CLI imports | 4% → 64%  | 20 min | 3, 6       |
| 8   | Create go.work                     | 20% → 80% | 5 min  | 7          |
| 9   | Verify build and tests             | 20% → 80% | 10 min | 8          |
| 10  | Re-vendor dependencies             | 20% → 80% | 10 min | 9          |
| 11  | Update flake.nix                   | 20% → 80% | 15 min | 10         |
| 12  | Update documentation               | Remaining | 10 min | 9          |
| 13  | Final verification                 | Remaining | 10 min | 11, 12     |

---

## Detailed Steps

### Task 1: Create core module skeleton

**What:** Create `core/` directory and `core/go.mod` with the core module definition.

**Steps:**

```bash
mkdir -p core
cat > core/go.mod << 'EOF'
module github.com/larsartmann/oxlint-auto-configure/core

go 1.26.2

require github.com/stretchr/testify v1.11.1
EOF
```

**Verify:** `cat core/go.mod` shows correct content.

**Rollback:** `rm -rf core/`

---

### Task 2: Move core packages to `core/`

**What:** Move 6 packages from `pkg/` to `core/` using `git mv`.

**Steps:**

```bash
mkdir -p core
git mv pkg/rule core/rule
git mv pkg/profile core/profile
git mv pkg/detect core/detect
git mv pkg/config core/config
git mv pkg/diff core/diff
git mv pkg/format core/format
```

**Files affected:**

- `pkg/rule/*` → `core/rule/*` (3 files + 2 embedded)
- `pkg/profile/*` → `core/profile/*` (1 file)
- `pkg/detect/*` → `core/detect/*` (1 file)
- `pkg/config/*` → `core/config/*` (3 files)
- `pkg/diff/*` → `core/diff/*` (1 file)
- `pkg/format/*` → `core/format/*` (1 file)

**Verify:** `ls core/` shows 6 directories. `pkg/` only contains `oxlint/`.

**Rollback:** `git mv` each back to `pkg/`.

---

### Task 3: Update core internal imports

**What:** Replace all internal import paths within core packages.

**Import path changes:**

```
github.com/larsartmann/oxlint-auto-configure/pkg/rule     → github.com/larsartmann/oxlint-auto-configure/core/rule
github.com/larsartmann/oxlint-auto-configure/pkg/profile   → github.com/larsartmann/oxlint-auto-configure/core/profile
github.com/larsartmann/oxlint-auto-configure/pkg/detect    → github.com/larsartmann/oxlint-auto-configure/core/detect
github.com/larsartmann/oxlint-auto-configure/pkg/config    → github.com/larsartmann/oxlint-auto-configure/core/config
github.com/larsartmann/oxlint-auto-configure/pkg/diff      → github.com/larsartmann/oxlint-auto-configure/core/diff
github.com/larsartmann/oxlint-auto-configure/pkg/format    → github.com/larsartmann/oxlint-auto-configure/core/format
```

**Files to update (within core/):**

| File                            | Current Internal Imports                |
| ------------------------------- | --------------------------------------- |
| `core/config/generator.go`      | `pkg/detect`, `pkg/profile`, `pkg/rule` |
| `core/config/configure.go`      | `pkg/detect`, `pkg/profile`, `pkg/rule` |
| `core/config/validate.go`       | `pkg/rule`                              |
| `core/detect/detector.go`       | `pkg/profile`, `pkg/rule`               |
| `core/diff/differ.go`           | `pkg/config`                            |
| `core/profile/profile.go`       | `pkg/rule`                              |
| `core/config/generator_test.go` | `pkg/detect`, `pkg/profile`, `pkg/rule` |
| `core/config/configure_test.go` | `pkg/detect`, `pkg/profile`, `pkg/rule` |
| `core/config/validate_test.go`  | `pkg/rule`                              |
| `core/detect/detector_test.go`  | `pkg/profile`, `pkg/rule`               |
| `core/diff/differ_test.go`      | `pkg/config`                            |
| `core/profile/profile_test.go`  | `pkg/rule`                              |
| `core/rule/registry_test.go`    | (none internal)                         |
| `core/format/format_test.go`    | (none internal)                         |

**Steps:**

```bash
find core/ -name '*.go' -exec sed -i \
  's|github.com/larsartmann/oxlint-auto-configure/pkg/|github.com/larsartmann/oxlint-auto-configure/core/|g' {} +
```

**Verify:**

```bash
cd core && GOWORK=off go build ./...
cd core && GOWORK=off go test ./...
```

**Rollback:** `git checkout -- core/`

---

### Task 4: Create analysis module skeleton

**What:** Create `analysis/` directory and `analysis/go.mod`.

**Steps:**

```bash
mkdir -p analysis
cat > analysis/go.mod << 'EOF'
module github.com/larsartmann/oxlint-auto-configure/analysis

go 1.26.2

require (
	github.com/larsartmann/go-finding v0.3.0
	github.com/larsartmann/oxlint-auto-configure/core v0.0.0
	github.com/stretchr/testify v1.11.1
)

replace github.com/larsartmann/oxlint-auto-configure/core => ../core
EOF
```

**Verify:** `cat analysis/go.mod` shows correct content.

**Rollback:** `rm -rf analysis/`

---

### Task 5: Move oxlint package to `analysis/`

**What:** Move `pkg/oxlint` to `analysis/oxlint` using `git mv`.

**Steps:**

```bash
git mv pkg/oxlint analysis/oxlint
rmdir pkg  # should be empty now
```

**Files affected:**

- `pkg/oxlint/detector.go` → `analysis/oxlint/detector.go`
- `pkg/oxlint/detector_test.go` → `analysis/oxlint/detector_test.go`
- `pkg/oxlint/version.go` → `analysis/oxlint/version.go`
- `pkg/oxlint/version_test.go` → `analysis/oxlint/version_test.go`
- `pkg/oxlint/fix.go` → `analysis/oxlint/fix.go`
- `pkg/oxlint/fix_test.go` → `analysis/oxlint/fix_test.go`
- `pkg/oxlint/errors.go` → `analysis/oxlint/errors.go`

**Verify:** `ls analysis/oxlint/` shows all files. `pkg/` no longer exists.

**Rollback:** `git mv analysis/oxlint pkg/oxlint`

---

### Task 6: Update analysis imports

**What:** Update import paths in analysis module files.

**Import path changes:**

```
github.com/larsartmann/oxlint-auto-configure/pkg/rule → github.com/larsartmann/oxlint-auto-configure/core/rule
```

**Files to update:**

- `analysis/oxlint/detector.go` — imports `pkg/rule`
- `analysis/oxlint/detector_test.go` — imports `pkg/rule`

**Steps:**

```bash
find analysis/ -name '*.go' -exec sed -i \
  's|github.com/larsartmann/oxlint-auto-configure/pkg/|github.com/larsartmann/oxlint-auto-configure/core/|g' {} +
```

**Verify:**

```bash
cd analysis && GOWORK=off go build ./...
cd analysis && GOWORK=off go test ./...
```

**Rollback:** `git checkout -- analysis/`

---

### Task 7: Update root go.mod and CLI imports

**What:** Update root `go.mod` to depend on core + analysis, update all CLI imports.

**Root go.mod becomes:**

```
module github.com/larsartmann/oxlint-auto-configure

go 1.26.2

require (
	github.com/larsartmann/oxlint-auto-configure/analysis v0.0.0
	github.com/larsartmann/oxlint-auto-configure/core v0.0.0
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/larsartmann/go-finding v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/larsartmann/oxlint-auto-configure/core => ./core
	github.com/larsartmann/oxlint-auto-configure/analysis => ./analysis
)
```

**Import path changes in CLI:**

| File                            | Old Imports                                                                     | New Imports                                                                               |
| ------------------------------- | ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| `internal/cli/cmd_analyze.go`   | `pkg/format`, `pkg/oxlint`, `pkg/rule`                                          | `core/format`, `analysis/oxlint`, `core/rule`                                             |
| `internal/cli/cmd_configure.go` | `pkg/config`, `pkg/detect`, `pkg/diff`, `pkg/oxlint`, `pkg/profile`, `pkg/rule` | `core/config`, `core/detect`, `core/diff`, `analysis/oxlint`, `core/profile`, `core/rule` |
| `internal/cli/cmd_report.go`    | needs check                                                                     | may reference pkg/ types                                                                  |
| `internal/cli/cmd_validate.go`  | needs check                                                                     | may reference pkg/ types                                                                  |
| `internal/cli/cmd_root.go`      | needs check                                                                     | may reference pkg/ types                                                                  |
| `internal/cli/commands_test.go` | `pkg/config`, `pkg/profile`, `pkg/rule`                                         | `core/config`, `core/profile`, `core/rule`                                                |

**Steps:**

```bash
# Update all imports in internal/cli and cmd/
find internal/ cmd/ -name '*.go' -exec sed -i \
  's|github.com/larsartmann/oxlint-auto-configure/pkg/|github.com/larsartmann/oxlint-auto-configure/core/|g' {} +

# Fix oxlint-specific imports (should point to analysis module)
find internal/ cmd/ -name '*.go' -exec sed -i \
  's|github.com/larsartmann/oxlint-auto-configure/core/oxlint|github.com/larsartmann/oxlint-auto-configure/analysis/oxlint|g' {} +

# Run go mod tidy
GOWORK=off go mod tidy
```

**Verify:**

```bash
GOWORK=off go build ./...
GOWORK=off go vet ./...
```

**Rollback:** `git checkout -- go.mod go.sum internal/ cmd/`

---

### Task 8: Create go.work

**What:** Create `go.work` at repo root.

**Steps:**

```bash
cat > go.work << 'EOF'
go 1.26.2

use (
    .
    ./core
    ./analysis
)
EOF
```

**Verify:** `go work sync` completes without errors.

**Rollback:** `rm go.work`

---

### Task 9: Verify build and tests

**What:** Full verification that everything works.

**Steps:**

```bash
# Workspace mode (developer experience)
go build ./...
go test ./...
go vet ./...

# Per-module verification
cd core && go build ./... && go test ./... && go vet ./... && cd ..
cd analysis && go build ./... && go test ./... && go vet ./... && cd ..

# Without workspace (consumer experience)
GOWORK=off go build ./...
```

**Verify:** All commands exit 0.

**Rollback:** N/A (verification only).

---

### Task 10: Re-vendor dependencies

**What:** Regenerate vendor directory with multi-module structure.

**Steps:**

```bash
GOWORK=off go mod tidy
GOWORK=off go mod vendor
```

**Note:** The vendor directory layout changes in multi-module projects. Each module's dependencies are vendored under `vendor/`. Verify flake.nix can still build.

**Verify:** `ls vendor/` shows all expected dependencies including `github.com/larsartmann/oxlint-auto-configure/core/` and `github.com/larsartmann/oxlint-auto-configure/analysis/`.

**Rollback:** `git checkout -- vendor/`

---

### Task 11: Update flake.nix

**What:** Update flake.nix to handle multi-module structure.

**Key changes:**

- Build command may need adjustment if vendor layout changes
- Test command (`nix run .#test`) must work with workspace
- Verify `vendorHash` or `vendorPath` is correct after re-vendoring
- The `makeWrapper` for oxlint binary is unchanged

**Verify:** `nix build .` and `nix run .#test` pass.

**Rollback:** `git checkout -- flake.nix`

---

### Task 12: Update documentation

**What:** Update AGENTS.md and README.md to reflect new structure.

**AGENTS.md changes:**

- Update "Key Files" table with new paths (`core/rule`, `analysis/oxlint`)
- Update "Testing" section with per-module test commands
- Update "Nix" section if build commands changed
- Add note about `go.work` and multi-module structure
- Update import path examples

**Verify:** Read through AGENTS.md — all paths and commands are accurate.

**Rollback:** `git checkout -- AGENTS.md README.md`

---

### Task 13: Final verification

**What:** End-to-end verification of the complete modularization.

**Checklist:**

- [ ] `go build ./...` passes at root
- [ ] `go test ./...` passes at root
- [ ] `go vet ./...` passes at root
- [ ] `go build ./...` passes in `core/`
- [ ] `go test ./...` passes in `core/`
- [ ] `go build ./...` passes in `analysis/`
- [ ] `go test ./...` passes in `analysis/`
- [ ] `go mod tidy` changes nothing in any module
- [ ] `go mod verify` passes in all modules
- [ ] `go work sync` succeeds
- [ ] `GOWORK=off go build ./...` passes (consumer experience)
- [ ] `nix build .` passes
- [ ] `nix run .#test` passes
- [ ] No circular imports in `go mod graph`
- [ ] Documentation is accurate

---

## Dependency Order Diagram

```
Task 1 (core skeleton)
  └── Task 2 (move core packages)
        └── Task 3 (update core imports) ──────────┐
                                                    │
Task 4 (analysis skeleton)                         │
  └── Task 5 (move oxlint package)                 │
        └── Task 6 (update analysis imports) ──────┤
                                                    │
                      Task 7 (update root + CLI) ◄──┘
                        └── Task 8 (create go.work)
                              └── Task 9 (verify build)
                                    ├── Task 10 (re-vendor)
                                    │     └── Task 11 (update flake.nix)
                                    ├── Task 12 (update docs)
                                    └── Task 13 (final verification)
                                              (depends on 11 + 12)
```

---

## Commit Plan

| After Task | Commit Message                                                                         |
| ---------- | -------------------------------------------------------------------------------------- |
| 3          | `refactor(core): extract core module with rule, profile, detect, config, diff, format` |
| 6          | `refactor(analysis): extract analysis module with oxlint go-finding integration`       |
| 7+8        | `refactor: wire root module with core + analysis, add go.work`                         |
| 9          | `chore: verify multi-module build and tests pass`                                      |
| 10         | `chore: re-vendor dependencies for multi-module structure`                             |
| 11         | `build(nix): update flake.nix for multi-module structure`                              |
| 12         | `docs: update AGENTS.md and README.md for modular structure`                           |
| 13         | `chore: final multi-module verification`                                               |

Each commit leaves the project in a buildable, testable state.

---

## Pareto Impact Summary

| Tier                         | Tasks       | Impact                                                      |
| ---------------------------- | ----------- | ----------------------------------------------------------- |
| **1% → 51%** (Foundational)  | Tasks 1-3   | Core module with zero external deps — publishable, reusable |
| **4% → 64%** (High leverage) | Tasks 4-7   | go-finding isolated to analysis module                      |
| **20% → 80%** (Broad value)  | Tasks 8-11  | go.work, vendor, nix — complete working system              |
| **Remaining** (Polish)       | Tasks 12-13 | Documentation and final verification                        |
