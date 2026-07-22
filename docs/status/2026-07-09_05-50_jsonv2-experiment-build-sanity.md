# Status Report — jsonv2 Experiment + Build Sanity

**Date:** 2026-07-09 05:50 UTC
**Session:** Fix `test-race` failure from buildflow run (`encoding/json/v2` build-constraint error)
**Scope:** Local Go build/test/vet/lint + flake.nix + CI YAML + .gitignore
**Trigger:** Paste of buildflow output showing `test-race`, `govalid-generate`, `go-fix` failures

---

## TL;DR

- **Root cause found and fixed:** `encoding/json/v2` is still behind `goexperiment.jsonv2` build tag in Go 1.26.4. The codebase imports v2 but no tool ever set the experiment.
- **All local checks pass** with `GOEXPERIMENT=jsonv2`: vet, build, test -race, golangci-lint, govulncheck.
- ~~**Not committed** (per safety rules): 4 source files modified, 1 gitignore negation restored, vendor/ left in buildflow's pre-existing dirty state.~~ **Committed** in subsequent merges (8913345, 4216a2c, dc7d858): GOEXPERIMENT=jsonv2 plumbing, vendor/ sync, and CI hardening are now in `master`.

---

## a) FULLY DONE

1. **flake.nix — `packages.default`**: Added `env.GOEXPERIMENT = "jsonv2"` so `nix build` and `nix flake check` build/test both enable jsonv2.
2. **flake.nix — `devShells.default`**: Added `GOEXPERIMENT = "jsonv2"` so `nix develop` enables it for direct `go test`/`go vet`.
3. **flake.nix — `devShells.ci`**: Added `GOEXPERIMENT = "jsonv2"` for CI shell.
4. **.github/workflows/ci.yml — `test` job**: Hoisted env to job level: `GOPRIVATE`, `GOWORK`, `GOEXPERIMENT=jsonv2`. Removed 4 redundant per-step env blocks.
5. **.github/workflows/ci.yml — `security` job**: Same job-level env added.
6. **.github/workflows/ci.yml — `lint` job**: Same job-level env added.
7. **AGENTS.md — Testing section**: Updated 3 commands to include `GOEXPERIMENT=jsonv2` prefix.
8. **AGENTS.md — Nix section**: Added `env.GOEXPERIMENT = "jsonv2"` gotcha line.
9. **AGENTS.md — Important Gotchas**: Added `GOEXPERIMENT=jsonv2` entry with policy + mechanism explanation.
10. **AGENTS.md — Testing section**: Added explanation paragraph about why the experiment is required.
11. **.gitignore**: Restored `!vendor/` negation that buildflow's gitignore-upserter had dropped (broke tracking of new vendor files).
12. **Verified**: `go vet ./...` passes with the experiment set.
13. **Verified**: `go build ./...` passes.
14. **Verified**: `go test -race ./...` passes — 8/8 packages, race + coverage.
15. **Verified**: `golangci-lint run --timeout=5m ./...` reports 0 issues.
16. **Verified**: `govulncheck ./...` reports 0 exploitable vulnerabilities.
17. **Verified**: `python3 -c "yaml.safe_load(...)"` confirms ci.yml YAML is valid.
18. **Verified**: YAML structure via diff — 12 clean lines added to ci.yml, no duplication/corruption.

---

## b) PARTIALLY DONE

19. **flake.nix structure**: Edits applied correctly but `nix flake check .` still fails because the **vendor/ changes from buildflow's auto-upgrade step are not yet committed**, so `lib.fileset` cannot include the new `vendor/github.com/larsartmann/go-finding/pipeline/path_safety.go` (which contains `resolveSafePath`). Local `go build` works because it reads the filesystem; nix reads git.
20. **ci.yml edits**: Applied to all 3 jobs but not pushed; cannot verify without commit + push + GitHub Actions run.
21. **AGENTS.md**: Updated but the "**DecideCategory**" entry and others from the existing document were left untouched — no regression, but also no audit of stale entries.

---

## c) NOT STARTED

22. **Vendor directory commit**: 11 vendor files modified + 1 untracked (`path_safety.go`) by the earlier buildflow `go mod vendor` run. These pre-date this session. Not committed because rule is "never commit unless asked". Cannot complete fix without this.
23. **go.mod / go.sum commit**: 3 indirect dep bumps (`go-finding/pipeline` `20260708144747`, `golang.org/x/sync` `v0.22.0`, `golang.org/x/sys` `v0.47.0`) also pre-date this session and not committed.
24. **Re-run buildflow**: Cannot be invoked in this session; deferred to user.
25. **Push to remote**: Not done (NEVER push unless explicitly asked).
26. **PR creation**: Not done — no branch was created.
27. **Tests in CI**: Cannot verify without push.
28. **Govulncheck in CI**: CI calls `golang/govulncheck-action@v1` which uses `go-version-input`. The action picks up job env but unverified whether `GOEXPERIMENT` propagates correctly to the action's go invocation. Likely fine but not tested.
29. **docs/status HTML report**: Per skill, status reports use the HTML kit. This session wrote a `.md` file instead — deviated from convention because user asked for `.md` explicitly.
30. **FEATURES.md / TODO_LIST.md audit**: AGENTS.md updates didn't trigger re-audit of feature status / todo lists. Out of scope for this fix but should be considered.

---

## d) TOTALLY FUCKED UP

31. **First ci.yml edit attempt corrupted the file.** I tried to replace the test job body with new content but my `old_string` was incomplete (`...${{`) and my `new_string` didn't include the security/lint job tails — result was `run: go vet ./...{{ secrets.SSH_PRIVATE_KEY }}` (literal `{{` embedded) plus 20 lines of duplicated content. Caught via `git diff`. Recovered with `git restore .github/workflows/ci.yml`, re-applied 3 small targeted edits that each touched only the `runs-on:` line. Lesson: **for YAML rewrites, restore + multi-edit is safer than one big find/replace**, especially with secrets-like `{{ }}` substrings.
32. **flake.nix edits worked on first try but I almost repeated the ci.yml mistake.** Caught it before editing by re-reading after a failed dry-run thought.
33. **My initial AGENTS.md `edit` also failed once** because the file's mtime changed between my `view` and `edit` calls. Re-viewed, re-edited successfully. Lesson: **don't batch View → Edit if there's any chance of intervening tool calls mutating the file**.
34. **Tried `nix flake check` 3 times, all failed with the same `resolveSafePath` error** because vendor isn't committed. I spent time investigating nix internals (source archive, gcroot cleanup) when the actual issue is upstream of nix (working tree not committed). Lesson: **when local build works but nix doesn't, suspect git/fileset state, not nix config**.
35. **Tried `trash` mention in `.crush` config** — irrelevant, no trash references in scope. Just noting my own thoughts drifted.

---

## e) WHAT WE SHOULD IMPROVE

36. **Buildflow should set `GOEXPERIMENT=jsonv2` in `.buildflow.yml`** so its own go-fix / test-race / govet steps don't fail on the same root cause. The user's buildflow run failed for the same reason I fixed locally. This is the SAME bug, just surfaced in a different runner. Should be fixed at the project-config level so no future run hits it.
37. **Buildflow's gitignore-upserter should preserve manual `!negation` entries** OR warn before dropping them. The `!vendor/` line is load-bearing for nix sandbox builds; losing it silently breaks the build. Buildflow runs once and dropped the line — catastrophic and silent.
38. **Buildflow should not auto-upgrade deps during a `--fix` session** without explicit confirmation. The user's run silently bumped `go-finding/pipeline` + `x/sync` + `x/sys` and modified 11 vendor files + created `path_safety.go`. Auto-upgrade on `--fix` is a footgun.
39. **CI YAML should declare `env` at job level by convention**, not step level. My fix moved them up, but the original repo had them per-step — that's a footgun (easy to forget when adding a new step). Consider a linter rule.
40. **flake.nix should pass `env.GOEXPERIMENT` to `mkShell` via `shellHook`** rather than the `env` attr (which sets env at shell-init, not for `go run` sub-processes if invoked differently). Actually `env` attr is correct — but worth documenting why.
41. **AGENTS.md should have a "Common build failures" section** so the next person who hits `encoding/json/v2` build-constraint error has a 2-second lookup rather than rediscovering the GOEXPERIMENT dance.
42. **The repo's go-finding dependency on a private GitHub module** is a recurring friction point (GOPRIVATE, SSH agent in CI, GOPROXY=off in nix). Consider vendoring a pinned release tarball instead of relying on GOPROXY resolution at all.
43. **`govalid-generate` and `go-fix` buildflow failures were transient** (tooling errors, not project errors). They passed on direct invocation. Buildflow should distinguish between transient tool failures and project failures in its exit code, otherwise a single transient blocks the whole session.
44. **No status update HTML** — user asked for `.md`, I delivered `.md`. Skill `status-report` suggests HTML by default. Future status reports should default to HTML unless user says otherwise.
45. **Test coverage 74% in `internal/cli` is the lowest** — could be raised. Not blocking, but a target.

---

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT (sorted by impact)

**P0 — Block current fix completion** 46. Commit `vendor/`, `go.mod`, `go.sum`, plus my `.gitignore` + `flake.nix` + `ci.yml` + `AGENTS.md` changes. One commit per logical concern is fine; suggest one commit for "fix: enable encoding/json/v2 via GOEXPERIMENT" + one for "chore: re-vendor go-finding pipeline upgrade". 47. Run `nix flake check .` after commit — verify the `resolveSafePath` error is gone. 48. Push branch + open PR. Title candidate: `fix: enable encoding/json/v2 via GOEXPERIMENT for Go 1.26`.

**P1 — Prevent recurrence** 49. Add `env.GOEXPERIMENT=jsonv2` to `.buildflow.yml` so buildflow's own go steps don't fail. 50. Patch `.buildflow.yml` to exclude `vendor/` from its gitignore upserts OR preserve `!negation` lines. 51. Add CI status badge + govulncheck schedule (weekly cron) so security issues surface early. 52. Add a pre-commit hook that runs `GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...` so the experiment is always set.

**P2 — Code quality** 53. Bump `internal/cli` test coverage from 74% → 85%+. 54. Add BDD tests (via the `bdd-testing` skill) for the `configure`, `analyze`, `validate` CLI commands. 55. Run `naming-review` skill — many recent additions may have subtle naming smells. 56. Run `data-model-review` skill on `pkg/rule/rule.go` and `pkg/profile/profile.go` — branded types and severity unions are candidates for improvement. 57. Run `deduplicate-code` skill on the `marshalConfigJSON` path — likely more dedup opportunities. 58. Run `architecture-review` and `architecture-visualization` skills to surface split-brain / coupling issues.

**P3 — Docs** 59. Update `README.md` to mention `GOEXPERIMENT=jsonv2` requirement for `go install` from source. 60. Add `docs/INTERNALS.md` explaining the private go-finding + nix sandbox + vendor dance. 61. Refresh `FEATURES.md` (skill `features-audit`) — last status reports mention features that may have drifted. 62. Refresh `TODO_LIST.md` (skill `todo-list-builder`) — encoding/json/v2 migration item should be marked DONE.

**P4 — Future-proofing** 63. When Go 1.27 makes jsonv2 default, remove all `GOEXPERIMENT=jsonv2` references and simplify. 64. Consider replacing `cobra` with `kong` or a smaller CLI lib (out of scope, just noting). 65. Add `golangci-lint` `paralleltest` and `gocognit` linters (already on paralleltest per AGENTS.md). 66. Add `govulncheck` to pre-commit so vulnerabilities block commits. 67. Add `nix build .#checks.test` to a `direnv` `.envrc` so `nix-direnv` users auto-load. 68. Add a Dockerfile that bakes in `oxlint` + the binary — `Dockerfile` exists, verify it's up to date. 69. Add release drafter for `.goreleaser.yaml` — config exists, verify quality. 70. Pin `go-finding` to a release tag instead of a pseudo-version.

**P5 — Observability** 71. Add OpenTelemetry tracing to the analyze pipeline (out of scope but flagged). 72. Add structured JSON logs option to all CLI commands (currently text only). 73. Add `--explain` flag to `configure` that prints the decision tree for the generated profile.

**P6 — Stretch** 74. Support oxlint config validation in CI (parse generated `.oxlintrc.json` against the schema). 75. Auto-detect monorepo workspaces and generate per-package configs. 76. Add a `pkg/oxlint/diff.go` that diffs configs without needing both files to be valid.

---

## g) TOP 2 QUESTIONS I CANNOT FIGURE OUT MYSELF

**Q1: Should I have committed the buildflow-era vendor/ + go.mod changes myself, treating the rule "never commit unless asked" as overridden by the user's clearly stated goal of "make the failing build pass"?**

Context: The user's intent ("fix") clearly implies making the whole build work, not just the local CI. Without committing vendor, `nix flake check` still fails. With committing, the user gets a fully working build but loses visibility into the 3 dependency upgrades (`go-finding/pipeline` `20260708144747`, `golang.org/x/sync v0.22.0`, `golang.org/x/sys v0.47.0`) until they review the diff. I chose to be conservative and NOT commit, leaving 14 vendor files in the working tree. This is the kind of tradeoff where the user knows their tolerance for surprise and I don't. **Need explicit guidance.**

**Q2: Should the `GOEXPERIMENT=jsonv2` requirement stay forever, or should the project migrate to drop it when Go 1.27 ships?**

Context: I documented it as "required because the experiment is still gated in Go 1.26". The Go 1.27 release notes (which I cannot verify) may flip jsonv2 to default. If yes, all the GOEXPERIMENT plumbing becomes dead weight. If no, the project will carry the env permanently. The honest answer is "check when Go 1.27 lands" — but I can't predict that. **Need the user's policy: maintain-env-permanently vs plan-removal-in-1.27.**

---

## Resolution (2026-07-22)

This session produced the `GOEXPERIMENT=jsonv2` fix but left it uncommitted. The work was merged into `master` shortly afterward.

| Item                                     | Claim in report                | Resolution                                                            | Commit            |
| ---------------------------------------- | ------------------------------ | --------------------------------------------------------------------- | ----------------- |
| flake.nix `env.GOEXPERIMENT`             | Added locally                  | **SHIPPED** in `packages.default` and both devShells                  | 8913345           |
| `.github/workflows/ci.yml` job-level env | Edited locally                 | **SHIPPED** with `GOEXPERIMENT=jsonv2` at job level                   | 8913345 / 4216a2c |
| `AGENTS.md` GOEXPERIMENT note            | Updated locally                | **SHIPPED**; later expanded with full gotcha section                  | 8913345           |
| `.gitignore` `!vendor/` negation         | Restored locally               | **SHIPPED** and still present                                         | 8913345           |
| `vendor/` + go.mod/go.sum changes        | Left dirty                     | **COMMITTED** through re-vendor in dc7d858 and hardening in 87505b0   | dc7d858 / 87505b0 |
| BuildFlow `GOEXPERIMENT` support         | Suggested for `.buildflow.yml` | **OPEN** — not yet addressed in `.buildflow.yml`                      | —                 |
| `go.mod` Go 1.27 bump                    | Not started                    | **OPEN** — `go.mod` still declares `go 1.26.4`; gopls warnings remain | —                 |
| `TODO_LIST.md` / `FEATURES.md`           | Not started                    | **OPEN** — still missing as of 2026-07-22                             | —                 |

Current `master` is `0867c95` (CHANGELOG v0.2.1 update). The gopls warning and the missing project docs are still tracked in TODO_LIST.md.
