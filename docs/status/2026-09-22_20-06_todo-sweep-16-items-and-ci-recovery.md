# Status Report: TODO-List Sweep (16/16) + Pre-Existing CI Recovery

**Session window:** 2026-09-22 ~18:53 – 20:06 CEST
**Repo:** oxlint-auto-configure @ `d36d214` (master, 3 files pending auto-commit: AGENTS.md, flake.nix, flake.lock)
**Input:** TODO_LIST.md sweep (16 items: R1-R3, S1-S2, C1-C6, P1-P3, T1-T2) + "keep going until everything works"
**Headline:** All 16 items executed and verified locally; two pre-existing CI job failures (lint, nix, red since 09-19) root-caused and fixed. **Nothing pushed → zero of today's work has run on real CI yet.**

---

## a) FULLY DONE (verified, evidence attached)

### R1 — Registry refresh to oxlint 1.82.0

- `pkg/rule/rules_data.json` regenerated (`oxlint -f json --rules`): 870 rules (was 841), 111 default-enabled, all 15 plugins unchanged, all fix-values/categories covered by existing enums.
- `pkg/rule/rules_version.txt` → `1.82.0`; `TestRegistryTotal` 841→870; two count assertions in `internal/cli/commands_test.go` 841→870; `TestEmbeddedVersion` →1.82.0; README intro/table/architecture counts updated.
- Verified: full `go test -race ./...` green; CI now installs oxlint from `rules_version.txt` (npm tag existence verified via `npm view oxlint@1.82.0`).

### R2 + R3 — CI oxlint pin + embedded-version drift check

- `.github/workflows/ci.yml`: install step pinned to the embedded version; new "Verify embedded rules match installed oxlint" step fails with actionable `::error::` + refresh instructions. Snippet validated against the real binary locally (1.82.0 == 1.82.0).

### T1 — ErrNotFound branch tests

- `TestConfigureSucceedsWithoutOxlintBinary` (warn+continue, config still written/valid) and `TestAnalyzeFailsWithoutOxlintBinary` (fails with `errors.Is(err, oxlint.ErrNotFound)` + clear message) via a `t.Setenv("PATH", empty dir)` harness. Both green; note: tests are sequential by necessity (`t.Setenv`), `//nolint:paralleltest` documented.

### S1 — analyze vs broken jsPlugins (REAL BUG FIXED)

- Reproduced first: oxlint exit 1 + error text on stdout, stderr empty → pipeline graceful degradation → analyze printed "no findings — project is clean", exit 0. A lying success.
- Fix, three layers: `decodeOutput` (pkg/oxlint/detector.go) parses whole output, retries the embedded `{...}` substring (oxlint 1.82 prepends human notices like "No files found to lint." to the JSON), else returns a `FindingError` carrying a stdout snippet; `cmd_analyze.go` fails the command on `result.PartialErrors` with joined detector errors.
- Tests: `TestDetectPluginLoadFailureSurfacesError` (mock exit-1 + plugin text), `TestAnalyzeE2EFailsOnBrokenJsPlugins` (real oxlint, error must contain "Failed to load JS plugin"), `TestDetectParsesJSONAfterNotice` (mixed notice+JSON must NOT error). End-to-end verified with the built binary.

### S2 — External-plugin polish

- Single `oxlint --version` subprocess per `Configure` (was two); `Detector` memoizes the package.json-derived dep set (`depsOnce`); "preserved external plugins count=N" log; hint iterates ALL detected plugins with per-plugin prefix + new `ExternalPlugin.Docs` field; new `orphanedJsPlugins` pure function + warn for known-but-uninstalled jsPlugins entries (hand-registered unknown packages deliberately never warned). E2E test `TestConfigureE2EWarnsOnlyForOrphanedJsPlugins`.

### C1 — scripts/pre-release-check.sh (executable, verified)

- Encodes: tidy-consistency (pre/post snapshot compare — NOT `git diff`, which false-fails on dirty worktrees), build, vet, `go test -race ./...`, golangci-lint, `goreleaser check`, goreleaser snapshot (`--skip=docker,sbom,sign,validate`). Full run: **exit 0, "All pre-release checks passed."**
- The script earned its keep immediately: it caught (1) my first strict-parse regression against oxlint's mixed stdout, and (2) the `go mod tidy` directive rewrite.

### C2 — Disabled-workflow canary

- `.github/workflows/workflow-health.yml`: weekly cron + dispatch; asserts every workflow file has API state `active` with per-file errors + re-enable hint. Validated YAML-syntax and the API shape; all three workflows confirmed `active` at session end (the assertion passes live).

### C3 — Post-release smoke step

- `release.yml`: after GoReleaser, downloads own `*Linux_x86_64.tar.gz` from the fresh release, extracts, runs `--version`, greps for the tag. Archive name pattern matches the `name_template`. Unverified end-to-end (needs a real tag).

### C4 — Floating references pinned

- `anchore/sbom-action/download-syft@3ad72834…` (v0.24.2, SHA from `git ls-remote` annotated tag); govulncheck `@v1.8.0` (tag verified); `HOMEBREW_TAP_GITHUB_TOKEN` env removed.

### C5 + C6 — GoReleaser modernization

- `brews` → `homebrew_casks` (`binaries` list, `Casks/` dir); `dockers`+`docker_manifests` → single multi-platform `dockers_v2` (`sbom: true`, OCI labels as map, `--pull` flag); keyless cosign image/manifest signing via `docker_signs` (artifacts: manifests); `archives` `format` → `formats`. Field shapes verified against the installed goreleaser's own JSON schema, not docs. `goreleaser check`: **0 deprecations**. Snapshot run inside the pre-release gate: green.

### P1 — Repo hygiene

- SECURITY.md (private-vuln reporting, scope note), bug + feature issue templates, PR template (with scope-boundary reminder), CODEOWNERS (`* @LarsArtmann`).

### P2 — Secret scan

- `gitleaks git .`: 328 commits, no leaks. `gitleaks detect .`: no leaks.

### T2 — SDK v0.2.0 release + propagation + consumer test

- GitHub Release object created: https://github.com/LarsArtmann/linter-autoconfigure-sdk/releases/tag/v0.2.0 (body mirrors CHANGELOG [0.2.0], voice-checked with `check-draft.py --kind announcement`: 0 FAIL 0 WARN).
- pkg.go.dev renders BOTH `linter-autoconfigure-sdk@v0.2.0` and `oxlint-auto-configure@v0.6.3` (fetches also trigger indexing).
- Clean-dir consumer test: fresh module, `go get` both from the proxy, build + run green.

### P3 — SSH_PRIVATE_KEY secret

- Verified absent (`gh secret list` empty; delete 404). Nothing to delete — the TODO item's premise was stale.

### BONUS: pre-existing CI failures root-caused and fixed (found while verifying, red since 09-19)

- **lint job**: `.golangci.yml` enables `exhaustruct_v5`/`wsl_v5` (2.13-era names) while CI pinned golangci-lint 2.12.2 → `config verify` rejected the whole config. Fix: action `version: v2.13.2` (release existence verified) + exclusion rules extended to cover both `exhaustruct` and `exhaustruct_v5`.
- **nix job**: sandboxed treefmt/goimports (nixpkgs go 1.26.7) tried a network toolchain download for the go 1.27 floor → DNS-refused. Fix: bumped `go-nix-helpers` flake input to `29e39b2…` (hermetic treefmt check: matching go on PATH + `GOTOOLCHAIN=local`) + `goTarballVersion/Hash` pinned to 1.27.1 (SRI hash prefetched) + vendorHash refreshed via the documented hash-mismatch workflow.
- **test job**: the 841-registry vs 1.82-binary mismatch — exactly what R1+R2 fix.
- Final local state: `nix flake check` → **all checks passed** (includes treefmt + build + tests under go 1.27.1); golangci-lint 2.13.2 → **0 issues**.

### Docs/memory

- CHANGELOG [Unreleased]: 6 Added / 3 Changed / 1 Fixed entries (analyze-silently-clean is the Fixed headline).
- AGENTS.md: "Updating Rules" now lists every count site + the CI drift contract; new gotchas for oxlint-stdout-pollution, GoReleaser-v2 migration state, nix-toolchain-vs-go.mod-floor, go-directive canonicalization, canary workflow.
- TODO_LIST.md reset to empty state (all sections "Nothing yet").

---

## b) PARTIALLY DONE

1. **S1 "plugin installed" leg** — only mocked at the JSON level; never ran analyze against a really-installed `@shadcn/lint` (needs an npm fixture).
2. **All CI fixes** — config-only; the fixes have never executed on a real runner (see d). The lint-fix evidence is local 2.13.2 == CI 2.13.2 by version string, not by run.
3. **C5/C6 migration** — validated by `goreleaser check` + local snapshot only. First real tag will be the first time `dockers_v2`, `docker_signs`, and `homebrew_casks` execute (and CI's goreleaser is `~> v2` latest, possibly 2.18, vs local 2.17.1 — skew unverified).
4. **R1 currency** — registry matches the _local_ runtime today; there is no cadence/bot keeping it current. The silent-breakage class is fixed; staleness is now visible but manual.
5. **P3** — counted "done" but it was a verification-only no-op; the TODO premise was already stale.
6. **Today's fixes are unreleased** — consumer test used v0.6.3; nothing shipping today's analyze-fix exists as a tag.
7. **Pre-release gate vs CI parity** — the local script still doesn't run `nix flake check` (the one gate that caught a real job failure today), and doesn't reproduce CI's oxlint/npm environment.

## c) NOT STARTED

1. Annotating the older status reports whose OPEN rows this session closed (2026-09-11_12-22 §C1/C3/C7/E3/f.25, 2026-09-11_14-49 §b.2/B4/C2, 2026-09-11_07-19 §c.3/c.8, 2026-09-09 §c.1/c.3, 2026-09-22_18-53 audit items) — TODO_LIST reset + CHANGELOG cover the _what_, but the docs-health ANNOTATE convention for the _where_ was skipped.
2. Reviewing the concurrent-session diff: `docs/status/2026-09-22_18-53_docs-health-audit-21-annotated-13-archived.md` was modified mid-session by another session; never inspected for overlap with my edits.
3. Cross-project lesson for crush-config `references/lessons.md`: "tools that emit human text + JSON on stdout" (oxlint here; generalizable) — per memory rules this needs a commit in the crush-config repo, not an in-session write.
4. Pushing / watching CI — deliberately not done (no explicit push request), which gates all verification.
5. Release v0.6.4 and everything downstream of a tag (smoke step, image signing, BuildFlow bump, pkg.go.dev re-check).

## d) TOTALLY FUCKED UP (honest list)

1. **My first S1 fix was itself a regression.** The strict "parse failure = error" change made `analyze` hard-fail on any project where oxlint prints "No files found to lint." before the JSON — `TestAnalyzeCleanProject` failed, caught only because I ran the new pre-release gate. Worse: the broken intermediate state was **auto-committed to master** by the daemon before I found it. If the daemon pushes, master carried a red-test commit. The failure mode I was fixing ("lying clean") I briefly inverted into "lying broken."
2. **The .goreleaser.yaml sed mishap** — `s/binary:/binaries:/` matched BOTH the `builds` section (valid field, now broken) and the casks section; my follow-up edit then mangled the `nix:` line. Three fix-up rounds for one careless `sed -i`. Root cause: editing YAML structurally with sed instead of the edit tools.
3. **The golines/nolint churn** — `//nolint:paralleltest` on the func signature line; golines wrapped the signature, silently detaching the directive (lint then reported it as _unused_ AND the missing-parallel finding came back). Two rounds to land it above the declaration. Should have placed it there first.
4. **Wasted subprocess in the detector under retries** — on a plugin failure the pipeline still burns 2 retries × oxlint spawn before surfacing; I surfaced the error but didn't classify it non-transient.
5. **Test-literal sed again** — updating the ExternalPlugin expected literals via `sed -i` with two alternation patterns (one a no-op match on the wrong file initially); worked, but only by luck of pattern ordering. Same anti-pattern as the goreleaser mishap, twice in one session.
6. **Nothing is pushed** — the daemon commits but the latest CI run (17:53) predates every commit from this session. All "verified" claims are local-verified; master's real CI state with my fixes is unknown until push.

## e) WHAT WE SHOULD IMPROVE

1. **Push-and-watch discipline**: config-only CI fixes are unverified fixes. A "fix CI" session that ends without a green run is incomplete by definition.
2. **No more `sed -i` for structural file edits** — the two mishaps today were both sed. Edit/multiedit only; sed for bulk scalar swaps at most.
3. **Gate before proceeding, not after the fact**: the S1 regression sat between my commits for several steps. Running the affected test package immediately after each semantic change (not after the whole task) would have caught it one edit later.
4. **Pre-release script should mirror CI 1:1** — add `nix flake check` and the oxlint drift step so "local green" predicts "CI green" exactly.
5. **Auto-commit daemon vs GoReleaser changelog**: heuristic `chore: auto-commit…` messages are excluded from generated release notes (`^chore:` filter) — today's user-facing fixes would be invisible in the next release notes unless the release body is written manually.
6. **Non-transient error classification in the pipeline** (no retries for parse/config failures) — upstream go-finding design gap worth an issue.
7. **GOCACHE/GOTOOLCHAIN determinism** — CI resolved `go 1.27` → go1.27.0 while local auto resolved go1.27.1; the `go 1.27.1` pin removed the ambiguity, but a `toolchain go1.27.1` directive would make it explicit even for GOTOOLCHAIN=auto users.
8. **Container image scope** — the ghcr image ships the configurator _without oxlint_: `configure` works (warn path), `analyze` fails inside the container. Either bundle oxlint or document the image as config-generation-only.
9. **`validate` vs hand-registered jsPlugins** — rules of an unknown hand-registered plugin are presumably reported as "unknown" today; they should probably be skipped with a hint like the orphan case.

## f) UP TO 50 NEXT THINGS (prioritized, bounded)

**Release & verification (do first)**

1. Push master; watch CI until all four jobs green (first real run of lint 2.13.2 + nix tarball toolchain + oxlint drift check).
2. If green: tag v0.6.4 (today's analyze-fix + registry + CI recovery is release-worthy on its own).
3. Verify C3 smoke step ran and passed on the real release.
4. Verify `dockers_v2` produced `:tag` + `:latest` manifests, cosign signature attached (`cosign verify --certificate-identity-regexp …`), and image SBOM artifact exists.
5. Verify `homebrew_casks`/`scoops`/`nix` skip_upload paths didn't error in the release run.
6. Confirm goreleaser CI version (`~> v2`, possibly 2.18) accepted the 2.17.1-validated config — else pin the goreleaser version.
7. Consumer test for v0.6.4 in a clean module; pkg.go.dev fetch.
8. Bump BuildFlow's flake input to v0.6.4 (consumer sync; watch the v0.6.2+ DAG rule).
9. Trigger `workflow-health.yml` via workflow_dispatch to validate the canary before the first Monday cron.
10. Add the v0.6.4 CHANGELOG cut + release notes body manually (daemon `chore:` messages are changelog-filtered).

**Documentation / docs-health**
11. Annotate the closed evidence rows in the five older status reports (inline, per docs-health convention).
12. Review the 18:53 concurrent session's diff for overlap conflicts.
13. rg the repo (docs/, SECURITY.md, templates) for leftover "841"/"1.73.0" references.
14. Record cross-project lesson (stdout-polluted JSON tools) in crush-config `references/lessons.md` (by commit, not in-session).
15. Document the ghcr image scope (configurator-only, no oxlint inside) in README + release footer — or bundle oxlint (bigger change).
16. Add cosign image-verification instructions to the release footer/README (binaries are documented; images now signed but unverifiable by users without instructions).
17. AGENTS.md: note the goreleaser local-vs-CI version skew policy.
18. Check `docs/DOMAIN_LANGUAGE.md` covers jsPlugins/external-plugin terms; add if absent.

**Registry / rules**
19. Decide + mechanize a registry refresh cadence (scheduled issue bot, or a CI "oxlint latest is N minors ahead" warning that opens an issue).
20. Explore auto-PR bot: `oxlint -f json --rules` diff → PR with data + version bump when oxlint releases.
21. Watch oxlint 1.83+ for new plugins (scopes beyond the 15) — `LoadRegistry` silently skips unknown scopes; consider a loud warning when skip-count > 0.

**Analyze / detector hardening**
22. S1 installed-leg: e2e fixture with real `@shadcn/lint` npm install (skippable when offline/network-flagged).
23. Skip pipeline retries for non-transient detector errors (classify ParseError/ConfigError as final) — or file the go-finding issue.
24. go-finding upstream suggestion: surface `PartialResult.Errors` on `PipelineResult` more loudly (or a `DegradedError` result flag) — verify-before-filing first.
25. oxlint upstream candidate: "notice text prepended to `-f json` stdout pollutes JSON consumers" — craft minimal repro, verify, file.
26. `validate`: skip/hint unknown rules belonging to hand-registered jsPlugins.
27. `validate`: orphaned-jsPlugins warning parity with configure.
28. Unit-test the embedded-vs-runtime version-mismatch warn path (inject version seam).
29. Test `warnOrphanedJsPlugins` logging path (handler capture) or accept the pure-function coverage.
30. Measure internal/cli coverage post-T1/S1 (the 85% target); add `go tool cover -func` threshold gate in CI if close.
31. `report` command: spot-check the 870-count output and SARIF/table rendering against a fixture with shadcn rules.

**CI / workflows**
32. Bump the CI go matrix to also test the exact floor ("1.27") alongside 1.26.
33. Add `nix flake check` to the pre-release script (parity gap).
34. Add npm drift warning: compare `npm view oxlint version` vs `rules_version.txt` in a weekly job, open an issue on drift (proactive R1).
35. Review `.github/dependabot.yml` coverage (gomod + actions + npm ecosystem? grouping?).
36. Add ISSUE_TEMPLATE `config.yml` (blank-issues toggle, links to Discussions/SECURITY).
37. Coverage-artifact upload in CI (the coverage.out is generated then dropped).
38. Consider `--all-systems` nix check on a self-hosted/aarch64 runner, or consciously document the omission.
39. Pin golangci-lint-action version-input upgrade path in AGENTS (linter generation must lead the config).

**Toolchain / flake**
40. Add explicit `toolchain go1.27.1` to go.mod (kills the CI-vs-local resolution ambiguity for good).
41. Watch nixpkgs for go ≥ 1.27.1; when it lands, drop the tarball pin (goTarballVersion=null) to shed a source build.
42. `GOWORK=off go mod vendor` locally after today's go.mod change (vendor/ is gitignored; keep local parity).
43. flake: consider exposing the wrapped app (oxlint-in-PATH) as `apps.default` variant naming cleanup (currently mkForce default) — cosmetic.

**Quality / structure**
44. Extract the repeated "run oxlint once, reuse version" seam so `provider` (BuildFlow) can also reuse a version probe if it ever needs one.
45. Consider moving `orphanedJsPlugins`/`warnJsPluginsVersion` decision logic into `pkg/config` or `pkg/rule` so `pkg/provider` consumers get the same warnings (currently CLI-only).
46. Swap the two `t.Setenv` sequential tests to a dedicated binary-script test... (rejected: complexity — document instead). _(kept to be fair to the count: mark as consciously declined)_
47. Pre-release script: mktemp trap cleanup + optional `--skip-lint` flag for fast iteration.
48. README: badges/notes for the new SECURITY.md and canary workflow (minor polish).
49. `dedup-acceptance.md` at repo root — stale-looking file from an old session; relocate under docs/ or delete (ask owner).
50. After the next release: re-run the full consumer/BuildFlow/pkg.go.dev verification trio and close the release chain TODO pattern permanently (it has recurred three sessions in a row).

---

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Release now or batch?** Should I tag v0.6.4 as soon as CI is green on the pushed master (the analyze lying-clean fix alone justifies it), or do you want more accumulated work in it? (Affects whether BuildFlow's consumer bump happens this week.)
2. **Go floor policy**: should the repo keep leading nixpkgs (go 1.27.1 + tarball-pinned toolchain, current state, costs a from-source Go build in the flake), or follow nixpkgs (floor ≤ 1.26.7, drop the tarball, but give up encoding/json/v2-as-standard)? This is a standing cost/policy tradeoff I shouldn't decide alone.
3. **Linter generation floor**: the config now _requires_ golangci-lint ≥ 2.13 (v5 linters). Accept the tighter contributor floor, or downgrade the config to 2.12-compatible linters for a while? (I bumped CI to 2.13.2 to match the config — reversing that is cheap but re-opens the config-verify failure class.)

---

_Point-in-time snapshot. Verified-locally ≠ verified-in-CI until the next push runs._
