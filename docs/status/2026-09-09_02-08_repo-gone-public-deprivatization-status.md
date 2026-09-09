# Status Report: Repo Went Public (De-Privatization)

**Date:** 2026-09-09 02:08 CEST
**Session scope:** "Should we make this public?" → full de-privatization of build/CI/docs → `gh repo edit --visibility public`
**Repo state at writing:** `LarsArtmann/oxlint-auto-configure` is **PUBLIC**; working tree clean; all session changes committed by the auto-commit daemon (`d8c1b94`, `7706545`, `bb88d57`).

---

## TL;DR

The repo is public and buildable by anyone. The perceived blocker ("private dependencies") turned out to be a documentation lie — all five `LarsArtmann/*` deps were already public. The real blockers were `git+ssh://` flake URLs, a silent flake/go.mod version drift (go-atomic-write v0.4.0 vs v0.5.1), a `go-nix-helpers` upgrade that silently broke the nix build, and SSH-secret plumbing threaded through every CI workflow. All fixed locally and verified via `nix flake check`; the repo is now public. CI has **not yet proven itself** post-cleanup (no push since). `docs/status/*` went public without an explicit decision.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                                                                                                                                                             | Evidence                                                                                     |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| 1  | Verified actual visibility of all 5 `LarsArtmann/*` deps + tags + module-proxy listing via `gh`/proxy.golang.org (go-finding v1.8.0, pipeline/v1.8.0, gogenfilter v3.4.0, go-atomic-write v0.5.1, go-error-family v0.10.0; go-nix-helpers public but **tagless**)                                                                                                | `gh repo view --json visibility`; `proxy.golang.org/.../@v/list`                             |
| 2  | `flake.nix`: 3× `git+ssh` inputs → `github:` tag/rev pins matching go.mod; `GOPRIVATE` dropped from `shellExtraEnv`; `subPackages = [ "cmd/oxlint-auto-configure" ]` added (required after go-nix-helpers upgrade); `vendorHash` updated; go-atomic-write pin v0.4.0 → v0.5.1                                                                                    | `nix build` + `./result/bin/oxlint-auto-configure --version` prints real version             |
| 3  | `nix flake check` green (build + test + treefmt), including the Go test suite with oxlint                                                                                                                                                                                                                                                                        | `all checks passed!`                                                                         |
| 4  | `ci.yml`: removed SSH-agent + `Configure GOPRIVATE` steps from test/security/lint jobs, all per-step `GOPRIVATE` envs, and the nix job's SSH-key step                                                                                                                                                                                                            | commit `bb88d57` (−44 lines); YAML validated via `yaml.safe_load`                            |
| 5  | `release.yml`: removed private-repo git config + ssh-agent steps; **fixed pre-existing YAML bug** — duplicate `with:` keys had orphaned `fetch-depth: 0` onto the ssh-agent step; restored it on checkout                                                                                                                                                        | commit `bb88d57` (−5 lines)                                                                  |
| 6  | Docs de-privatized: README (removed `export GOPRIVATE` + "requires private go-finding" note), CONTRIBUTING (3→2 required env vars, "Private Dependencies" → "Dependencies", stale versions corrected), AGENTS.md (5 entries: vendored-deps, mkPreparedSource, Dependencies, design-principle 9, "Private go-finding" gotcha → "Public deps" + dep-bump workflow) | commit `bb88d57` (+19/−77 across 5 files)                                                    |
| 7  | Confirmed `go install ...@latest` resolves: repo has tags (`v0.4.0` latest) and all deps are proxy-fetchable without auth                                                                                                                                                                                                                                        | `gh api repos/.../tags`; GitHub go_modules graph job completed successfully post-publication |
| 8  | Confirmed release path is public-safe: `LarsArtmann/homebrew-tap` is PUBLIC                                                                                                                                                                                                                                                                                      | `gh repo view`                                                                               |
| 9  | Repo flipped to PUBLIC and verified                                                                                                                                                                                                                                                                                                                              | `gh repo view` → `PUBLIC`                                                                    |
| 10 | Repo-wide leftover scan: no `GOPRIVATE`/`ssh://`/`insteadOf` outside historical `docs/status/`; Dockerfile and `.goreleaser.yaml` clean                                                                                                                                                                                                                          | grep this session                                                                            |

## b) PARTIALLY DONE

| # | Item                       | Works                                                                  | Missing                                                                                                                                                                                                                       | Effort             |
| - | -------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------ |
| 1 | CI without secrets         | Workflows edited, YAML valid                                           | **Zero post-cleanup CI runs observed** — the slimmed test/security/lint/nix jobs are unproven until the next push (proxy-fetch without GOPRIVATE is theoretically fine, theoretically)                                        | S (watch next run) |
| 2 | `release.yml` fix          | `fetch-depth: 0` back on checkout                                      | Never exercised — next `v*` tag is the first real test; prior releases ran with a **shallow checkout** (duplicate-key YAML meant checkout lost `fetch-depth`), so past goreleaser changelogs may have been silently truncated | S                  |
| 3 | `docs/status/*` triage     | Flagged pre-flip; grep found no live secrets                           | User never decided; 15 internal AI session reports + `dedup-acceptance.md` are now **public**, and **no git-history secret scan was run** before publishing                                                                   | S–M                |
| 4 | Public-ready consumability | Source + nix path work for strangers today                             | `v0.4.0` (what `@latest` resolves to) predates every fix in this session — tag-based consumers get the SSH-era docs; no new release cut                                                                                       | S                  |
| 5 | go-nix-helpers pin hygiene | Pinned to rev `a97742e` (master HEAD) so builds are reproducible _now_ | Repo has **zero tags**; pin is untagged and the master-HEAD default (`subPackages = [ "." ]`) silently broke this build — upstream fix or old-rev pin undecided                                                               | M                  |

## c) NOT STARTED

| # | Item                                                                                         | Why not started                                                     | Priority |
| - | -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- | -------- |
| 1 | Git-history secret scan (gitleaks/trufflehog) before search engines index the repo           | Only working-tree grep was done; session ended before history audit | Critical |
| 2 | New release tag shipping the public-ready state                                              | Waiting on user decision (see question 2)                           | High     |
| 3 | Deleting now-unused `SSH_PRIVATE_KEY` repo secret                                            | Cleanup, zero functional impact, just not done                      | Medium   |
| 4 | `docs-health` HARVEST of section (f) into TODO_LIST.md/ROADMAP.md                            | User said "then wait for instructions"                              | High     |
| 5 | Tagging go-nix-helpers / upstream fix for the `subPackages` default regression               | Ownership decision (question 3)                                     | Medium   |
| 6 | Public-repo hygiene: SECURITY.md, issue/PR templates, repo description/topics/social preview | Not in session scope                                                | Medium   |
| 7 | CHANGELOG.md entry for de-privatization + publication                                        | Daemon commits carry no semantic message to harvest from            | Medium   |

## d) TOTALLY FUCKED UP

1. **My opening analysis asserted the deps were private — without verifying.** I built the entire "hard blocker" framing on AGENTS.md's stale claims and `git+ssh` URLs, and the user had to correct me ("they are all public i think"). Severity: wrong premises in a publish/no-publish decision. Root cause: trusted memory/docs over a 5-second `gh repo view`. Mitigation: none needed downstream — the URL-pin work was necessary regardless — but the verification failure is exactly the anti-pattern AGENTS.md §"Independently verify tool output" bans.
2. **I introduced a silent nix-build breakage and exit-code 0 masked it.** Pinning go-nix-helpers to master HEAD switched `subPackages` default to `[ "." ]`; `nix build` "succeeded" with an **empty output directory** and no error. I only caught it because `--version` was chained onto the build command. Severity: had I trusted the green build, a publicly-unbuildable flake would have shipped. Root cause: upstream default change + no guard (installPhase doesn't assert the binary exists; `enableCheck = false`). Mitigation: explicit `subPackages` (done); real fix belongs upstream (item f-13).
3. **The flake built go-atomic-write v0.4.0 while go.mod demanded v0.5.1 — silently.** The `mkPreparedSource` replace directive overrides any version, so the "reproducible" nix build compiled stale source for an unknown period. Found by accident while editing the input, not by any check. Severity: reproducibility lie; could bite any dep. Mitigation: pins now match go.mod; systemic fix = consistency check (item f-12).
4. **release.yml shipped a malformed duplicate-`with:` YAML block (pre-existing).** Parsers resolve duplicate keys last-wins: ssh-agent got `fetch-depth: 0`, checkout got nothing → every prior release ran goreleaser on a shallow clone. Severity: degraded (not broken) release artifacts. Fixed this session; effect unverifiable until the next tag.
5. **The repo went public with 16 internal files exposed and no history audit.** `docs/status/*` (15 AI session reports with `/home/lars` paths and tooling narratives) + `dedup-acceptance.md`. Working-tree grep shows no secrets, but `git log` history was never scanned, and the triage decision was left open when the user said "do it". Severity: medium (privacy/embarrassment, not credentials — as far as a shallow scan shows). Mitigation: none yet — needs question 1 answered.

## e) WHAT WE SHOULD IMPROVE

1. **"Verify external claims" must include our own docs.** AGENTS.md said "private"; reality said public. Stale memory files actively misled a publish decision. Rule going forward: any claim used for an irreversible decision gets checked against the live system (`gh`, proxy, actual build), not against AGENTS.md.
2. **Green build ≠ working build.** `nix build` exiting 0 with an empty output should be impossible. Until upstream guards it, every flake input bump here needs a `result/bin/<binary> --version` smoke test chained into the build command — as a habit, not luck.
3. **go.mod ↔ flake pin drift needs a machine check.** The v0.4.0/v0.5.1 drift survived multiple sessions because only humans compare the two files. A tiny test/CI script diffing `go.mod` requires against `flake.nix` pins would have caught it instantly.
4. **Secrets plumbing should be deleted the day it becomes unnecessary, not documented around.** Three CI jobs + release + docs all carried GOPRIVATE/SSH ceremony for deps that were public the whole time. The ceremony made the "private" story self-perpetuating.
5. **Floating references in a public repo's supply chain** — `pnpm add -g oxlint` (unpinned), `govulncheck@latest`, `anchore/sbom-action@v0` (tag, not SHA; every other action is SHA-pinned). CI green-ness and SBOM integrity currently depend on other people's moving tags. Also: unpinned oxlint versions can break `TestRegistryTotal` the day oxlint ships a new rule.
6. **Personal-machine details in public docs** — `/home/lars/projects/go.work` appears in CONTRIBUTING.md and AGENTS.md as if it were universal. Public repos need generic phrasing ("if a parent `go.work` exists").
7. **Auto-daemon heuristic commit messages erase semantic history.** This session's work is spread across `chore: auto-commit N file(s)` commits, so neither CHANGELOG harvesting nor `git log` archaeology can reconstruct _why_. Meaningful milestones deserve manual, well-written commits.
8. **Pre-publication checklist gap.** Going public should have been: history secret scan → internal-content triage → floating-ref pin audit → flip. Two of four happened.

## f) Up to 50 things to get done next

Ranked by impact. (HARVEST input — route to TODO_LIST.md; starred = ROADMAP fuel.)

| #   | Task                                                                                                                                                          | Impact   | Effort | Category      |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1   | Run full-history secret scan (gitleaks) on the now-public repo                                                                                                | Critical | S      | Quality       |
| 2   | Decide + execute docs/status/* and dedup-acceptance.md triage (delete from HEAD vs history rewrite; needs Q1 answer)                                          | Critical | S–L    | Cleanup       |
| 3   | Watch first post-cleanup CI run end-to-end (all 4 jobs, no secrets)                                                                                           | Critical | S      | Quality       |
| 4   | Cut release tag (v0.4.1 or v0.5.0) shipping de-privatized state; verify release.yml works with restored fetch-depth                                           | High     | S      | Release       |
| 5   | Verify `go install github.com/larsartmann/oxlint-auto-configure/cmd/oxlint-auto-configure@latest` from a clean, unauthenticated machine                       | High     | S      | Quality       |
| 6   | Verify pkg.go.dev indexes the module and renders docs                                                                                                         | High     | S      | Quality       |
| 7   | Verify `nix run github:LarsArtmann/oxlint-auto-configure -- configure` works for an anonymous stranger                                                        | High     | S      | Quality       |
| 8   | Pin `oxlint` version in CI (pnpm) and devshell — unpinned oxlint can break `TestRegistryTotal` rule-count on upstream releases                                | High     | S      | Bug           |
| 9   | Pin `govulncheck@latest` to a version                                                                                                                         | High     | S      | Quality       |
| 10  | SHA-pin `anchore/sbom-action/download-syft@v0` (only floating action left)                                                                                    | High     | S      | Quality       |
| 11  | Add CI/test check: go.mod dep versions == flake.nix input pins (would have caught the v0.4.0/v0.5.1 drift)                                                    | High     | M      | Quality       |
| 12  | Add postInstall assertion (or `enableCheck`) so an empty nix build output fails loudly; file upstream against go-nix-helpers' `subPackages = [ "." ]` default | High     | M      | Bug           |
| 13  | Tag go-nix-helpers (zero tags) and pin this flake to a tag instead of master-HEAD rev (needs Q3 answer)                                                       | Medium   | M      | Cleanup       |
| 14  | Delete unused `SSH_PRIVATE_KEY` secret from repo settings                                                                                                     | Medium   | S      | Cleanup       |
| 15  | CHANGELOG.md entry: de-privatization + publication                                                                                                            | Medium   | S      | Documentation |
| 16  | Run docs-health HARVEST on this report's section (f)                                                                                                          | High     | S      | Documentation |
| 17  | Replace `/home/lars/projects/go.work` references in CONTRIBUTING/AGENTS with generic phrasing                                                                 | Medium   | S      | Documentation |
| 18  | Add SECURITY.md (public repo now has a vulnerability-reporting surface)                                                                                       | Medium   | S      | Documentation |
| 19  | Add issue + PR templates                                                                                                                                      | Medium   | S      | Documentation |
| 20  | Set GitHub repo description/topics/social preview for public presence                                                                                         | Medium   | S      | Documentation |
| 21  | Update ROADMAP.md/FEATURES.md to reflect public launch + remaining public-hardening items                                                                     | Medium   | S      | Documentation |
| 22  | Add CI badge + license badge to README                                                                                                                        | Medium   | S      | Documentation |
| 23  | Configure Dependabot/Renovate (Go deps + GitHub Actions) — public repo now has external eyes; stale pins get flagged                                          | Medium   | S      | Quality       |
| 24  | Verify goreleaser brew/scoop/nix publish targets all point at public repos (homebrew-tap confirmed; check scoop/nix steps in full)                            | Medium   | S      | Release       |
| 25  | Smoke-test a full `goreleaser release --snapshot` locally to prove the slimmed release env has no hidden GOPRIVATE assumptions                                | Medium   | M      | Release       |
| 26  | Document the dep-bump workflow (go.mod + flake tag + vendorHash) as a script or CONTRIBUTING section, not just an AGENTS gotcha                               | Medium   | M      | Documentation |
| 27  | Audit `docs/status/*.html` reports for personal info (paths, emails) beyond the secret-pattern grep                                                           | Medium   | M      | Cleanup       |
| 28  | Consider history rewrite ONLY if scan (item 1) finds something (requires force-push approval per policy)                                                      | High*    | L      | Cleanup       |
| 29* | Evaluate re-enabling `enableCheck` in flake now that deps are public (checks were possibly disabled for private-fetch friction)                               | Medium   | S      | Quality       |
| 30* | Announce publication (post/README badge/awesome-lists) — user-call                                                                                            | Low*     | S      | Feature       |
| 31* | docs-site / website-launch for the tool (landing page, demo)                                                                                                  | Low*     | L      | Feature       |
| 32* | Homebrew formula freshness: confirm tap auto-updates on next release                                                                                          | Low      | S      | Release       |
| 33  | Confirm `nix flake check --systems x86_64-linux` (CI syntax) still valid on CI's Nix version — local Nix rejected the flag; version skew unexamined           | Medium   | S      | Bug           |
| 34  | Add a CI job running the flake build for a second system (aarch64-darwin) to catch arch breakage                                                              | Low      | M      | Quality       |
| 35  | Kill stale local artifacts: `reports/coverage.out`, `coverage/`, stray `result` symlinks (gitignored, but clutter)                                            | Low      | S      | Cleanup       |

## g) Questions I cannot figure out myself

1. **`docs/status/*` + `dedup-acceptance.md` are now public.** Do you want them stripped (forward-delete from HEAD is clean; a history rewrite would need your force-push approval), and do you know of anything _inside_ those 15 session reports that my secret-pattern grep wouldn't have caught?
2. **Release timing:** tag `v0.4.1`/`v0.5.0` now so `@latest` serves the de-privatized state, or wait until the slimmed CI proves itself on the next push?
3. **go-nix-helpers:** it has zero tags and its current master silently changed `subPackages` defaults. Do you want to (a) tag it and pin this flake to the tag, (b) pin the older known-good rev here, or (c) fix/restore auto-detection upstream and re-pin?

---

_Report generated per `status-report` skill; Markdown written per explicit user instruction (skill default is HTML)._
