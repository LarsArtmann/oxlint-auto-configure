# Status Report — Docs Health + Update Old Docs Pass

**Date:** 2026-07-22 09:28:25 CEST (Wednesday)  
**Branch:** `master`  
**Commit:** `4a64b0d` — `docs(status): improve resolution report navigation and readability`  
**Report file:** `docs/status/2026-07-22_09-28_docs-health-and-update-old-docs.md`  
**Reporter:** Crush (this session)

---

## TL;DR

- **Four 2026-07-`*` status reports were read and verified.** The update-old-docs annotations were already committed in `0ff370f` and `4a64b0d`; they pass the specificity and fresh-open tests.
- **A full docs-health audit was run** on every living doc. Found drift in `README.md`, `AGENTS.md`, `CONTRIBUTING.md`, `docs/DOMAIN_LANGUAGE.md`, `internal/cli/cmd_configure.go`, and `pkg/profile/profile.go`.
- **All identified drift was fixed.** `FEATURES.md` and `TODO_LIST.md` were created. `CHANGELOG.md` Unreleased section was updated.
- **All quality gates pass:** `go test -race ./...`, `go vet ./...`, `nix flake check .`.

---

## a) FULLY DONE

1. **Read all four `/2026-07-`*`status files** in`docs/status/`.
   - `2026-07-07_16-42_build-fix-and-cleanup-session.html`
   - `2026-07-07_23-03_full-cleanup-and-hardening-session.html`
   - `2026-07-09_05-50_jsonv2-experiment-build-sanity.md`
   - `2026-07-17_12-25_buildflow-vendor-fix-status.md`
2. **Verified existing update-old-docs annotations** in `0ff370f` and `4a64b0d`.
   - Inline/appendix placement is correct, no top-of-file banners.
   - Annotations cite specific commits (`8913345`, `4216a2c`, `804d117`, `87505b0`, etc.).
   - Resolution sections distinguish shipped vs. open items.
3. **Ran a full docs-health audit** on every living doc.
   - `README.md`, `AGENTS.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `docs/DOMAIN_LANGUAGE.md`, `FEATURES.md`, `TODO_LIST.md`.
4. **Fixed `README.md` drift.**
   - Removed the non-existent `TypeScript` category from the Profiles table.
   - Aligned `strict` and `recommended` rows with `pkg/profile/profile.go`.
   - Added `node` to the Vitest detection row.
   - Added `GOEXPERIMENT=jsonv2` to Development commands.
   - Added a `GOEXPERIMENT=jsonv2` note to the Go install section.
5. **Fixed `AGENTS.md` drift.**
   - Updated `go-finding` version from `v1.2.0` to `v1.2.1`.
   - Corrected the `recommended` profile description.
6. **Fixed `CONTRIBUTING.md` commands.**
   - Added `GOWORK=off` and `GOEXPERIMENT=jsonv2` to local commands.
   - Added Nix alternatives.
7. **Rewrote `docs/DOMAIN_LANGUAGE.md`.**
   - Replaced template placeholders with actual project domain terms (Rule, Plugin, Category, Profile, Severity, Project Type, Detector, Categorizer, Generator, etc.).
8. **Created `FEATURES.md`.**
   - Honest feature inventory with status and code evidence for every item.
9. **Created `TODO_LIST.md`.**
   - Only open, short-term actionable work; no completed items retained.
10. **Updated `CHANGELOG.md` Unreleased section.**
    - Documented the new docs, the fixed commands, and the corrected profile table.
11. **Fixed user-facing code comments.**
    - `internal/cli/cmd_configure.go` help text no longer claims `TypeScript` is a severity category.
    - `pkg/profile/profile.go` comments now match the actual `profileSpecs` table.
12. **Verified all counts and commands against code.**
    - Rule totals (716), category counts, plugin counts confirmed via the registry.
    - Strict and recommended profiles confirmed to produce identical output.
13. **All quality gates passed.**
    - `GOWORK=off GOEXPERIMENT=jsonv2 go test -race ./...` — 9 packages pass.
    - `GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...` — clean.
    - `nix flake check .` — all checks passed.

---

## b) PARTIALLY DONE

1. **golangci-lint baseline.**
   - I ran `GOWORK=off GOEXPERIMENT=jsonv2 golangci-lint run ./...` locally.
   - It reports 117 issues (depguard, err113, varnamelen, mnd, etc.), many of which appear suppressed or configured differently in CI.
   - A clean, reproducible baseline has not been established.
2. **docs-health "fitness" score.**
   - All identified drift was fixed. ~~Some docs (e.g., `ROADMAP.md`) remain uncreated~~ `ROADMAP.md` created in `4c9ea2f` (2026-07-26). `TODO_LIST.md` was new at this time.
3. **Profile differentiation.**
   - `strict` and `recommended` are functionally identical in the current code.
   - The docs now reflect this reality; a product decision on whether to differentiate them is still open.
4. **Internal markdown link audit.**
   - Only one internal link exists (`README.md` → `LICENSE`); it resolves.
   - External links were not verified live.

---

## c) NOT STARTED

1. ~~Push the current working-tree changes to `origin/master`.~~ DONE: `4a64b0d` is on `origin/master`;
2. Resolve the `go.mod` Go version mismatch with `encoding/json/v2` (gopls warnings).
3. Add a CI check that `go mod vendor` produces no diff.
4. Add BuildFlow to CI.
5. Establish a clean `golangci-lint run ./...` baseline.
6. Run `hierarchical-errors` directly and establish a baseline.
7. Investigate the BuildFlow step-count discrepancy (42 vs 35) from the 2026-07-17 session.
8. Add tests for `cmd/oxlint-auto-configure/main.go` (entry point at 0% coverage).
9. Add an E2E integration test: configure → validate → report round-trip.
10. Wire `flake.nix` ldflags for `commit`, `date`, and `builtBy` so nix builds show full version metadata.
11. Increase `internal/cli` test coverage toward 85%+.
12. ~~Create `ROADMAP.md` for long-term direction.~~ DONE: `4c9ea2f`;
13. Decide whether to add `gosec` to the CI security job.
14. Decide testify → ginkgo/gomega migration policy for this project.
15. Decide whether to execute or archive the modularization proposal (docs already deleted; decision remains).
16. Add a `go mod verify` CI step.
17. Add a reproducibility check for `nix build`.
18. Add a Go version matrix in CI (1.26 + 1.27 + tip).
19. Add a pre-commit hook that runs `go mod vendor` and `nix fmt --check`.
20. Add `--explain` flag to `configure` that prints the decision tree.
21. Add shell completion tests for the CLI.
22. Add `--version` flag tests.
23. Add tests for error message formatting and user-facing output.
24. Add integration tests for the `configure` command.
25. Add integration tests for the `analyze` command.
26. Add integration tests for the `validate` command.
27. Add integration tests for the `report` command.
28. Add end-to-end tests that run against a real `oxlint` binary.
29. Add unit tests for `pkg/config` edge cases.
30. Add unit tests for `pkg/detect` edge cases.
31. Add unit tests for `pkg/oxlint` (currently 81.9%).
32. Add property-based tests for `.oxlintrc.json` config generation.
33. Add fuzz tests for config parsing and validation.
34. Add benchmark tests for rule registry loading.
35. Add tests for `--dry-run` behavior in `configure`.
36. Add tests for `--fix` behavior in `configure`.
37. Add tests for profile selection logic.
38. Add tests for plugin detection edge cases (no package.json, missing deps, etc.).
39. Add a `nix flake check` CI job.
40. Add dependency vulnerability scanning verification (`govulncheck` already runs; ensure `GOEXPERIMENT` propagation).
41. Add `--config` flag for custom config path.
42. Add `--output` flag to `configure` for custom output path.
43. Add `--profile` flag to `analyze` command.
44. Add Cobra completion subcommand.
45. Add benchmark tests for `parseOutput`.
46. Add benchmark tests for `Registry.Filter`.
47. Review and refactor generic `error` returns flagged by `hierarchical-errors`.
48. Review disabled `golangci-lint` linters and decide which to enable.
49. Add typed errors for `detect`, `config`, and `oxlint` packages.
50. Improve separation between CLI rendering and business logic.

---

## d) TOTALLY FUCKED UP!

1. **I initially re-edited the four 2026-07-`*` status files before realizing they were already annotated.** The edit tool reported success, but the changes were identical to HEAD, so git showed no diff. This wasted time and could have introduced noise if I had not caught it via `git hash-object`.
2. **`golangci-lint run ./...` locally reports 117 issues**, while CI and `AGENTS.md` claim it reports 0. This implies the local linter version, config, or environment differs from CI. Relying on the "0 issues" claim without a reproducible local baseline is misleading.
3. **`strict` and `recommended` profiles are functionally identical** in the code, yet `README.md`, `AGENTS.md`, and `cmd_configure.go` help text historically presented them as different. Fixing the docs exposed the underlying design question rather than resolving it.
4. **`CONTRIBUTING.md` had broken commands** (`go test ./... -race` and `golangci-lint run ./...` without `GOWORK=off` or `GOEXPERIMENT=jsonv2`). This would fail for any new contributor following the file literally.
5. **`docs/DOMAIN_LANGUAGE.md` was a placeholder template** with `.` and "Example Term". It provided no actual domain vocabulary and would confuse a new contributor or AI.
6. **`FEATURES.md` and `TODO_LIST.md` did not exist**, despite being referenced as canonical in previous status reports and the global documentation model.
7. **`AGENTS.md` cited the wrong `go-finding` version** (`v1.2.0` vs. `v1.2.1` in `go.mod`), which could mislead dependency audits.

---

## e) WHAT WE SHOULD IMPROVE!

1. **Add a CI `go mod vendor` consistency check.** This was the root cause of the 2026-07-17 BuildFlow failure; it should be caught before tooling runs.
2. **Add entry-point tests for `cmd/oxlint-auto-configure`.** 0% coverage for the main entry point is unacceptable.
3. **Resolve the `go.mod` Go version mismatch.** Either bump to `go 1.27` or add a `toolchain` directive so gopls stops warning about `encoding/json/v2`.
4. **Decide whether `strict` and `recommended` should differ.** If they should, change the code; if not, remove one or document the equivalence explicitly.
5. **Establish a reproducible local linter baseline.** Pin `golangci-lint` version in `flake.nix` or document the exact CI version so local results match CI.
6. **Add BuildFlow to CI.** The local BuildFlow workflow is not currently exercised in PRs.
7. **Improve `AGENTS.md` "Important Gotchas"** to mention the profile identity issue and the local linter mismatch.
8. **Add a `ROADMAP.md`** to capture long-term direction (e.g., monorepo support, public docs website, BDD tests).
9. **Add E2E round-trip test** so configure → validate → report is continuously verified.
10. **Wire flake.nix version metadata** so `nix run . -- --version` shows commit/date/builtBy, not "unknown".
11. **Consider running `hierarchical-errors` and `naming-review` skills** as dedicated passes, not just as CI side effects.
12. **Keep `FEATURES.md` and `TODO_LIST.md` current** after every future feature or dependency change.

---

## f) Up to 50 Things We Should Get Done Next (sorted by impact)

### P0 — Block current work from shipping

1. ~~Push the current docs-health changes to `origin/master` (or confirm review before pushing).~~ DONE: `4a64b0d` is on `origin/master`;
2. Resolve `go.mod` Go version / `encoding/json/v2` mismatch.
3. Add CI check that `go mod vendor` produces no diff.
4. Establish a reproducible local `golangci-lint` baseline.

### P1 — Prevent recurrence

5. Add BuildFlow to CI.
6. Add a `nix flake check` CI job.
7. Add a Go version matrix in CI (1.26 + 1.27 + tip).
8. Add a pre-commit hook that runs `go mod vendor` and `nix fmt --check`.
9. Pin `golangci-lint` version in `flake.nix`.

### P2 — Code quality

10. Add entry-point tests for `cmd/oxlint-auto-configure`.
11. Add E2E integration test: configure → validate → report round-trip.
12. Increase `internal/cli` coverage from ~74% toward 85%+.
13. Add BDD tests (via `bdd-testing` skill) for `configure`, `analyze`, `validate`, `report`.
14. Run `hierarchical-errors` directly and fix or baseline findings.
15. Run `naming-review` skill across the codebase.
16. Add typed errors for `detect`, `config`, and `oxlint` packages.
17. Review disabled `golangci-lint` linters and decide which to enable.
18. Add a `Runner` mock test for the real oxlint execution path.
19. Add documentation for `pkg/diff` and `pkg/format` packages.
20. Add mutation testing to evaluate test quality.

### P3 — Features

21. Differentiate `strict` and `recommended` profiles or document equivalence.
22. Add `--explain` flag to `configure` that prints the decision tree.
23. Add `--config` flag for custom config path.
24. Add `--output` flag to `configure` for custom output path.
25. Add `--profile` flag to `analyze` command.
26. Add Cobra completion subcommand.
27. Add structured JSON logs option to all CLI commands.
28. Support oxlint config validation in CI.
29. Auto-detect monorepo workspaces and generate per-package configs.
30. Add a `pkg/oxlint/diff.go` that diffs configs without both files being valid.

### P4 — Coverage & testing

31. Add unit tests for `pkg/config` edge cases.
32. Add unit tests for `pkg/detect` edge cases.
33. Add unit tests for `pkg/oxlint` edge cases.
34. Add property-based tests for `.oxlintrc.json` config generation.
35. Add fuzz tests for config parsing and validation.
36. Add benchmark tests for rule registry loading.
37. Add tests for `--dry-run` behavior in `configure`.
38. Add tests for `--fix` behavior in `configure`.
39. Add tests for profile selection logic.
40. Add tests for plugin detection edge cases (no package.json, missing deps, etc.).
41. Add tests for `--version` flag output.
42. Add shell completion tests for the CLI.
43. Add tests for error message formatting and user-facing output.
44. Add benchmark tests for `parseOutput`.
45. Add benchmark tests for `Registry.Filter`.
46. Add a coverage threshold check to CI (≥80%).

### P5 — Docs & process

47. ~~Create `ROADMAP.md` for long-term direction.~~ DONE: `4c9ea2f`;
48. Update `AGENTS.md` with the profile identity gotcha and linter baseline mismatch.
49. Add `docs/INTERNALS.md` explaining the private go-finding + nix sandbox + vendor dance.
50. Consider a public documentation website via the `website-launch` skill.

---

## g) Top 3 Questions I Cannot Figure Out Myself

1. **Should I push the current working-tree changes to `origin/master` now, or do you want to review the docs-health diff first?**
   Context: The working tree has 7 modified files and 2 new files (`FEATURES.md`, `TODO_LIST.md`). All quality gates pass. I did not commit because the rule is "never commit unless asked".

   **Resolved:** Changes committed as `4a64b0d` and pushed to `origin/master`.

2. **Should `strict` and `recommended` profiles remain functionally identical, or should we differentiate them?**
   Context: `pkg/profile/profile.go` currently gives both profiles the exact same `profileSpecs` map. The docs now reflect this reality. If they should differ, the code should change; if not, we could remove one profile or document the equivalence explicitly.

3. **Should I create `ROADMAP.md` now, or wait until more long-term direction is defined?**
   Context: The docs-health model flags `ROADMAP.md` as optional for a library/package, but previous status reports called it out as missing. The current `TODO_LIST.md` captures short-term work; a `ROADMAP.md` would be helpful for long-term vision (e.g., monorepo support, public docs).

   **Resolved:** `ROADMAP.md` created in `4c9ea2f` (2026-07-26) with themes, open questions, and non-goals.

---

## Appendix: Raw Verification Output

### Git

```
Branch: master
Commit: 4a64b0d [origin/master]
Working tree: 7 modified, 2 untracked
```

### Modified files

- `AGENTS.md`
- `CHANGELOG.md`
- `CONTRIBUTING.md`
- `README.md`
- `docs/DOMAIN_LANGUAGE.md`
- `internal/cli/cmd_configure.go`
- `pkg/profile/profile.go`

### New files

- `FEATURES.md`
- `TODO_LIST.md`

### Go Build & Test

```
GOWORK=off GOEXPERIMENT=jsonv2 go test -race ./...  # 9 packages pass
GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...         # clean
```

### Nix

```
nix flake check .  # all checks passed
```

### Local golangci-lint (not clean; CI mismatch noted)

```
GOWORK=off GOEXPERIMENT=jsonv2 golangci-lint run ./...  # 117 issues
```

---

## Resolution (2026-07-26)

| Item  | Claim in report                    | Resolution                  | Commit    |
| ----- | ---------------------------------- | --------------------------- | --------- |
| §b.3  | `ROADMAP.md` remain uncreated      | Created                     | `4c9ea2f` |
| §c.1  | Push docs-health changes to origin | Pushed; `4a64b0d` on origin | `4a64b0d` |
| §c.12 | Create `ROADMAP.md`                | Created                     | `4c9ea2f` |
| §f.1  | Push docs-health changes           | Pushed; `4a64b0d` on origin | `4a64b0d` |
| §f.47 | Create `ROADMAP.md`                | Created                     | `4c9ea2f` |
| §g.1  | Should I push?                     | Resolved: pushed            | `4a64b0d` |
| §g.3  | Should I create `ROADMAP.md`?      | Resolved: created           | `4c9ea2f` |

**Subsequent work not anticipated by this report:** A later session (2026-07-26) migrated config writes to `go-atomic-write` v0.3.0 (crash-durable: temp + fsync + atomic rename). Upgraded `go-finding` from v1.2.1 to v1.3.0. Removed `vendor/` from git tracking. See the 2026-07-26 status report for details.

---

_Report generated by Crush. Waiting for instructions._
