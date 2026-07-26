# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> For shipped work, see `CHANGELOG.md`.
> Items here are verified, not assumed.

## Build and Tooling

All previously listed build/tooling tasks have been completed:

- ✅ Embedded rules updated from oxlint `1.59.0` to `1.73.0` (716 → 841 rules, 108 → 113 enabled)
- ✅ `go.mod` Go version mismatch resolved (`go 1.26.5` + `GOEXPERIMENT=jsonv2`)
- ✅ `nix flake check` failure fixed (passes all checks)
- ✅ CI check for `go mod tidy` consistency added (`.github/workflows/ci.yml`)
- ✅ Reproducible `golangci-lint` baseline established: 0 issues local + CI, version pinned to `v2.12.2`
- ✅ Nix flake check CI job added (`nix` job in `.github/workflows/ci.yml`)
- ✅ gosec confirmed covered by golangci-lint (already in enabled linters list)
- ✅ flake.nix ldflags wired for `commit`, `date`, `builtBy` (full version metadata in nix builds)

## Testing

- ✅ Entry-point tests added for `cmd/oxlint-auto-configure/main.go` (`main_test.go`)
- ✅ E2E integration test: `configure` → parse output → verify round-trip via `config.FromJSON` (`e2e_test.go`)
- ✅ Atomic-write contract test: verify `writeConfig` leaves no `.tmp` files (`atomic_write_test.go`)
- ✅ `internal/cli` coverage increased from 74.2% → 82.7% (remaining gap is oxlint integration code)
- ✅ Coverage tests for `renderFindings`, `printSARIF`, `printReportJSON`, `sortedByPosition`, `resolveConfigPath`, `logDiffIfExisting`, `marshalConfigJSON` (`coverage_test.go`)

## Maintenance

- ✅ `hierarchical-errors` skill run: **0 findings** (codebase already uses `errors.AsType[E]` exclusively)
- ✅ `naming-review` skill run: **0 findings** (no vague names, no Manager/Handler/Helper, no Impl suffixes)
- ✅ `code-quality-scan` skill run: **0 issues** (build, lint, vet all clean)
- ✅ `deduplicate-code` skill run: **0 clone groups** at threshold 5
- ✅ Broken `#resolution` anchor link fixed in `docs/status/2026-07-17_*.md`

---

_Last reviewed: 2026-07-26_
