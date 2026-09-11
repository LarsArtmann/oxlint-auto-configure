# Status Report: GitHub Metadata Session — Brutal Self-Review

**Date:** 2026-09-11 07:27 CEST
**Scope:** This session only ("Give this repo a proper GitHub Metadata, description and co!") plus things noticed in passing. No unrelated research performed.
**Format note:** User explicitly requested `.md`, overriding the status-report skill's HTML default.

---

## Session Summary

| Step | Action | Result |
| ---- | ------ | ------ |
| 1 | Loaded `website-launch` skill Phase 6 (GitHub Metadata) | Trigger matched "set up GitHub metadata" |
| 2 | Gathered facts: `gh repo view`, README, workflows, `.goreleaser.yaml`, website/ check | Repo had NO description, NO topics, NO homepage, NO badges |
| 3 | `gh repo edit` — description + 15 topics | Applied, verified via `gh repo view` |
| 4 | README badge bar (CI \| Docker \| MIT) at README.md:3-5 | Applied (application badge set per skill) |
| 5 | AGENTS.md memory entry for metadata decisions | Applied |
| 6 | Verification: release run, ghcr anonymous pull, license detection | All green (details below) |

README.md + AGENTS.md edits were auto-committed by the daemon (241537f).

---

## a) FULLY DONE

1. **Description set + verified.** "Generate the optimal `.oxlintrc.json` — auto-detects your framework, enables the right plugins, and sets severity for all 841 oxlint rules". Confirmed live via `gh repo view`.
2. **15 topics set + verified.** `go`, `golang`, `cli`, `oxlint`, `oxc`, `linter`, `linting`, `static-analysis`, `code-quality`, `code-analysis`, `typescript`, `javascript`, `developer-tools`, `sarif`, `nix`. All genuinely apply (nix flake exists; SARIF output exists; no framework-lying topics).
3. **README badge bar.** CI (actions/workflows/ci.yml/badge.svg — workflow exists, name `CI` confirmed) | Docker (ghcr.io link) | MIT. Correctly used the **application** set — no Go Reference / Go Report Card, because this is a CLI, not an importable package.
4. **Docker badge honesty verified.** Mapped v0.5.0 → its release run (`success`, the "provide ghcr.io credentials" run), then confirmed the image is **anonymously pullable** via the ghcr registry API (200, tags: `v0.5.0`, `v0.5.0-amd64`, `v0.5.0-arm64`, `latest`). The badge does not lie.
5. **Homepage deliberately left empty** — no `website/` exists; setting one would 404. Decision recorded in AGENTS.md.
6. **AGENTS.md updated** — metadata decisions + badge rationale + "add homepage via website-launch skill if docs site ever exists" so future sessions don't re-derive.
7. **License** — MIT already present and auto-detected by GitHub; nothing to do (verified, not assumed).

## b) PARTIALLY DONE

1. **"Metadata and co" is 90% done, not 100%.** Missing: **social preview / OG image** (GitHub renders a default card from README badges — functional but bland; no asset was created or uploaded). Needs image generation + web UI upload (API cannot set it).
2. **CI badge render unverified.** I confirmed the workflow file exists but never fetched the `badge.svg` endpoint itself to confirm HTTP 200 / non-"no status" state. Probability of failure: near zero. Verification completeness: not zero-effort-done.
3. **Repo settings beyond metadata untouched.** Wiki (off), Issues (on) noted but Discussions, branch protection, merge strategy, labels, CODEOWNERS — all unchecked. "and co" arguably includes them.

## c) NOT STARTED

1. Social preview image creation/upload.
2. GitHub Discussions enablement (a product decision — see questions).
3. `.github/` community health files: SECURITY.md, issue templates, PR template, CODEOWNERS (`.github/` currently holds only `ci.yml` + `release.yml`).
4. Documentation website (website-launch skill) — deliberately deferred, noted in AGENTS.md.
5. README polish beyond badges: "Who is this for?" / "When NOT to use this" sections, Docker usage section, Releases/binary-download links.
6. pkg.go.dev listing check (module is public; `pkg/*` are importable — the badge decision followed the skill for apps, but whether pkg.go.dev actually lists the module was never checked).

## d) TOTALLY FUCKED UP

Nothing destructive. But two honest process failures:

1. **I violated read-before-edit on AGENTS.md.** Attempted the edit without viewing the file first in this conversation; the tool guard rejected it ("you must read the file before editing"). A guardrail caught me, not my discipline. Sloppy. Correct sequence followed afterwards (viewed lines 155-175, then exact-match edit succeeded).
2. **Ambiguous package-page fetch handled reactively, not proactively.** The anonymous HTML fetch of the ghcr package page returned a login-interstitial soup — inconclusive. I recovered correctly (registry API with anonymous token = ground truth), but the right order would have been registry API first, page second. Wasted one round trip on a signal I should have known was noisy.

## e) WHAT WE SHOULD IMPROVE (self-review answers)

1. **What did you forget?** Social preview image; Discussions/settings sweep; fetching the CI badge SVG; checking the sibling repo (`golangci-lint-auto-configure`) for house badge/metadata conventions before inventing mine — consistency across LarsArtmann repos beats per-revo originality.
2. **What's stupid that we do anyway?** The Docker badge is a *static* shields.io badge — it can never show live data (ghcr has no pull-count shield) and must be manually recolored if conventions change. Acceptable, but it's decoration, not signal.
3. **What could you have done better?** Verified the CI badge endpoint; batched the AGENTS.md view with the first edit attempt (would have avoided the guard rejection); checked sibling-repo conventions first.
4. **What can you still improve?** Everything in section (f), top items first.
5. **Did you lie to you?** No. Every claim above is backed by a command run this session (`gh repo view`, `gh release view`, `gh run list`, ghcr registry API). One claim leaned on inference: "package page may render login wall for logged-out users" — the registry pull is the verified fact; page rendering is cosmetic and unverified.
6. **How can we be less stupid?** For any future metadata task: checklist = description, topics, homepage-if-site, badges (verify each target URL returns 200), social preview, community files, settings sweep, sibling-repo consistency pass. This checklist is now in AGENTS.md-adjacent memory via this report; fold the durable parts into AGENTS.md when acted on.
7. **Ghost systems?** None created. The AGENTS.md entry and the live metadata agree (single source: this session's decisions).
8. **Scope creep?** No — stayed inside Phase 6, explicitly skipped the website launch phases. The "not started" list is deferred work, not creeped work.
9. **Removed something useful?** No.
10. **Split brains?** One tiny risk: README tagline ("not a linter, a configurator") vs the new description ("Generate the optimal...") are two phrasings of the same value prop. Intentional (description = feature-forward for search; tagline = scope boundary). Not a split brain, but worth remembering when either changes.
11. **Tests?** N/A — zero code changed this session. The only "test" was metadata verification, which was done with real API calls, not assumptions.

## f) Up to 50 things to get done next

Impact-sorted. Items 1-10 are actionable soon; 11+ are brainstorm / ROADMAP fuel (do NOT treat all 50 as commitments — harvest selectively).

**Metadata & presence (this session's thread):**
1. Fetch `actions/workflows/ci.yml/badge.svg` → confirm 200 (close the verification gap).
2. Diff badge/metadata against sibling `golangci-lint-auto-configure` → align conventions.
3. Generate + upload a social preview OG image (1200×640; HyperFrames or simple render).
4. Decide Discussions on/off; enable via `gh repo edit --enable-discussions` if yes.
5. Add `SECURITY.md` (security policy) — repo ships a linter-config tool, policy is cheap trust.
6. Add CODEOWNERS (`@LarsArtmann`).
7. Issue templates (bug/config-request) + PR template.
8. Add README version badge (`v0.5.0` via shields GitHub release endpoint — live data, unlike Docker).
9. Add "Docker" usage section to README with `docker run ghcr.io/larsartmann/oxlint-auto-configure:v0.5.0 configure` (images exist; README never mentions them).
10. Verify the Docker image actually contains `oxlint` in PATH (the binary needs it at runtime — GoReleaser Dockerfile builds the Go binary; oxlint inclusion unverified).

**README / docs quality:**
11. Add "Who is this for?" and "When NOT to use this" sections (website-launch content pattern, retrofittable without a site).
12. Link CONTRIBUTING.md and CHANGELOG.md from README (both exist, neither is linked).
13. Link the GitHub Releases page from README (binary downloads are advertised in release notes, not README).
14. README usage GIF/asciinema demo (configure → diff output).
15. Add `:::tip`-style callouts / tidy the scope-boundary blockquote.
16. Check whether pkg.go.dev lists the module; if yes, reconsider Go Reference badge for `pkg/` importables (decision currently: no).
17. Sync README rule-statistics table against the embedded registry count (claims 841; registry test enforces it — verify test count matches README).
18. Clarify `nix profile install` + `nix develop` flake-ref examples resolve against `v0.5.0` tag.
19. Move/absorb stray root-level `dedup-acceptance.md` into `docs/`.
20. README "Tools Using This" section is one link deep — either grow or cut it.

**Repo plumbing (noticed in passing):**
21. Branch protection on `master` (CI required check) — repo settings appear default.
22. Dependabot: Go modules + GitHub Actions version updates.
23. Consider CodeQL alongside govulncheck (CI security job currently govulncheck-only).
24. Uncommitted work-in-progress sits in the tree: `flake.nix`, `go.mod`, `go.sum`, and the 07-19 status report are modified (dep bump in flight by user/another session — NOT touched, per safety rules). Needs an owner.
25. Release notes for v0.5.0: audit quality (existence verified, content not reviewed).
26. `goreleaser check` deprecation warnings (`brews`, `dockers`) — plan migration to `brews`→`homebrew_casks`/new syntax or accept.
27. Add `--discussion-category` metadata to releases if Discussions gets enabled.

**Bigger arcs (ROADMAP fuel):**
28. Documentation website via website-launch skill (Astro + Starlight + Firebase, lars.software subdomain) — fills the empty homepage field.
29. Demo video for said website (HyperFrames HTML→MP4).
30. Docs-health HARVEST: route items from this list into TODO_LIST.md / ROADMAP.md (this report is the input; do on instruction).
31. Check TODO_LIST.md / FEATURES.md / ROADMAP.md freshness against actual state.
32. `analyze` command UX pass (exit codes, CI integration story, SARIF upload docs for GitHub Code Scanning).
33. Version-pin warning UX: `rules_version.txt` mismatch flow.
34. Consider `configure --fix` deprecation messaging (AGENTS.md says non-core).
35. Rules-data refresh workflow: document/automate `oxlint -f json --rules > pkg/rule/rules_data.json` + registry-test update as a script.
36. Add `report` command markdown output format for pasting into issues/PRs.
37. Test coverage audit for `pkg/diff` and `pkg/format` edge paths.
38. Consider a `doctor` command (oxlint presence, version match, config validity in one shot).
39. Ship v0.5.1 once items 1-10 land (metadata + README polish release).
40. Check whether `docker run` example needs volume mounts for project access; document if so.

**Nice-to-have / cosmetic:**
41. Star-history chart in README (only after organic stars exist — premature otherwise).
42. GitHub Sponsors/Funding metadata (only if you want it).
43. Repo labels triage set (bug/config/detection/profile).
44. Merge queue evaluation (probably overkill — single-maintainer repo).
45. Release RSS mention in README for version-watchers.
46. Verify README pipe-table renders correctly on github.com after badge insert (visual check).
47. Consider topic `config-generator` / `linter-configuration` (discoverability vs topic noise — cap ~15-20).
48. Align description em-dash usage with house style if a rule emerges for social copy (source-code rule doesn't apply to repo metadata).
49. Archive the two failed v0.5.0 release-run attempts in a short "release pipeline history" note (AGENTS.md already carries the lesson — check for duplication before writing).
50. Ask for the social-preview branding direction (colors/logo) — blocked on your input (see questions).

## g) Questions I cannot figure out myself

1. **Social preview image:** generate one automatically (I'd pick a dark card with the repo name + "841 rules, one command" line), or do you have brand assets / a preferred look? (The API can't upload it — you'd click it in repo Settings either way, but I can produce the PNG.)
2. **GitHub Discussions:** enable for Q&A/config-help, or keep Issues-only? (It's your community-surface preference; I can flip it via `gh repo edit` once decided.)
3. **Docs website:** should I kick off the website-launch flow (Astro + Starlight + Firebase, `oxlint-auto-configure.lars.software`) so the empty homepage field gets a real target, or is GitHub-README-only the intended end state for this tool?

---

**Unowned in-flight changes noticed (NOT touched, NOT mine):** `flake.nix` (2 lines), `go.mod`/`go.sum` (dep bump pattern), `docs/status/2026-09-11_07-19_v0.5.0-release-pipeline-repair.md` (30 lines). Owner: user or a parallel session.

**Now waiting for instructions.**
