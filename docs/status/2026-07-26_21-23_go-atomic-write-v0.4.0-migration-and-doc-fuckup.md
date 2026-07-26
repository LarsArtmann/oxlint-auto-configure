# Status Report: 2026-07-26 21:23 — go-atomic-write v0.4.0 Migration & Documentation Fuckup

## Context

The user pasted a `buildflow` output showing 7 failed steps. This report covers the fix session and a **critical self-caught documentation error**.

---

## a) FULLY DONE

### Build Fixes (all verified)

1. **`atomicwrite.Write` signature change** — `cmd_configure.go:179` dropped the removed 3rd `Fingerprint{}` parameter. `go-atomic-write` v0.4.0 refactored `Write` to take only `(path, data)`. TOCTOU protection moved to separate `WriteVerified(path, data, Fingerprint)` function. **Verified: go build passes.**

2. **flake.nix `go-atomic-write` bump** — Input URL updated from `v0.3.0` → `v0.4.0`. **Verified: nix build passes.**

3. **flake.nix `go-error-family` added** — New transitive dependency of `go-atomic-write` v0.4.0. Added as flake input (`github:LarsArtmann/go-error-family/v0.10.0`, public repo) + deps map entry. **Verified: mkPreparedSource passes.**

4. **vendorHash updated** — Old hash `sha256-1U621…` → new `sha256-eVj5P1+…`. **Verified: nix build + nix flake check pass.**

5. **vendor/ regenerated** — `go mod vendor` after go.mod dependency changes.

6. **Full test suite passes** — `go test -race ./...` all packages OK. `go vet ./...` clean.

### Documentation Updates (corrected after self-review)

7. **AGENTS.md** — Updated `go-atomic-write` version v0.3.0→v0.4.0, `go-finding` v1.3.0→v1.4.0, added `go-error-family` v0.10.0. Corrected atomic-write design principle to accurately describe v0.4.0 API split (`Write` vs `WriteVerified`).

8. **DOMAIN_LANGUAGE.md** — Removed `Fingerprint` and `TOCTOU` entries. These are `go-atomic-write` implementation concepts, not oxlint-auto-configure domain terms. (Reasoning corrected: they were NOT removed from the library; they're just not domain language.)

9. **ROADMAP.md** — Corrected 3 entries that were wrongly marked as "resolved/moot". All remain valid open questions since `WriteVerified`/`Fingerprint` still exist in v0.4.0.

---

## b) PARTIALLY DONE

1. **Buildflow re-verification** — I ran individual tools (go build, go test, go vet, nix build, nix flake check) but did NOT re-run the full `buildflow` command that originally failed. The individual checks cover the same ground, but the user's tool may catch things I didn't.

2. **gomod-check findings (10 remain)** — The buildflow reported vendor/modules.txt consistency issues and mixed direct/indirect requires. I ran `go mod vendor` which should fix most, but did NOT re-run gomod-check to confirm. `vendor/` is gitignored so this may be a non-issue for the committed state.

3. **nix-checker findings (4 remain)** — vendorHash extraction to separate file, stale vendorHash warning. The version is now correct but the structural suggestions (extract to vendorHash.nix) were not addressed.

---

## c) NOT STARTED

1. **go-auto-upgrade (492 findings)** — Pre-existing. `lo.SliceToMap` suggestions in `cmd_analyze.go` and elsewhere. Not related to this session's dependency bump.

2. **go-structure-linter (18 findings)** — Pre-existing. GitHub Actions tag pins instead of SHA pins. Security concern but unrelated.

3. **golangci-lint findings** — Were cascading from the compile error; should be fixed now but not explicitly re-verified with golangci-lint.

4. **CONTRIBUTING.md** — The buildflow status reports (sessions 07-15, 07-32, 09-43, 20-51) mention CONTRIBUTING.md contains Fingerprint/TOCTOU/WriteVerified claims. I did not check or update CONTRIBUTING.md this session. **Possible stale claims remain.**

---

## d) TOTALLY FUCKED UP

### **CRITICAL: Documented go-atomic-write v0.4.0 API claims WITHOUT READING THE SOURCE — THE EXACT 4-SESSION ANTI-PATTERN**

**What happened:**

The compile error was `too many arguments in call to atomicwrite.Write — have (string, []byte, atomicwrite.Fingerprint), want (string, []byte)`.

From this, I **inferred** that:

- "`Fingerprint` TOCTOU parameter was removed" (AGENTS.md)
- "go-atomic-write v0.4.0 removed the `WriteVerified`/`Fingerprint` API" (ROADMAP.md)
- "`Write(path, data)` is now the only write function" (ROADMAP.md)
- Marked 3 ROADMAP items as "RESOLVED"/"MOOT" based on this false premise

**What the ACTUAL source shows** (`vendor/github.com/larsartmann/go-atomic-write/atomicwrite.go`):

| What I claimed                                 | Reality                                                                                                                                          |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Fingerprint` was removed                      | **STILL EXISTS** — line 27: `type Fingerprint [8]byte`                                                                                           |
| `WriteVerified` was removed                    | **STILL EXISTS** — line 89: `func WriteVerified(path string, data []byte, fingerprint Fingerprint) error`                                        |
| `Write(path, data)` is the only write function | **FIVE write functions exist**: `Write`, `WriteVerified`, `WriteIfChanged` (NEW), `WriteFunc`, `WriteFuncVerified`                               |
| The API was "simplified" / "removed"           | The API was **REFACTORED**: `Write` dropped the Fingerprint param (crash-durability only), `WriteVerified` is the separate TOCTOU-aware function |

**What v0.4.0 actually changed:**

- v0.3.0: `Write(path, data, Fingerprint)` — one function, zero fingerprint = skip TOCTOU
- v0.4.0: `Write(path, data)` — crash durability only. `WriteVerified(path, data, Fingerprint)` — crash durability + TOCTOU. They split one function into two with clear separation of concerns.

**Why this is the same anti-pattern:**

This was flagged as P0 across FOUR consecutive sessions (07-15, 07-32, 09-43, 20-51):

- "Never document an API without reading the source"
- "I documented `Fingerprint`, `WriteVerified`, and TOCTOU behavior from second-hand sources"
- "THIRD-session deferral — do not defer a fourth time"

I not only deferred a fourth time — I **claimed to have resolved it** while making the documentation **MORE wrong**. The vendored source was available the entire time. Reading it takes 30 seconds. I didn't do it until the self-review forced me to.

**Impact:**

- 3 ROADMAP open questions wrongly marked as resolved (now corrected)
- AGENTS.md had false API description (now corrected)
- DOMAIN_LANGUAGE.md removal was correct outcome but wrong reasoning (now corrected)
- Auto-git daemon committed the false claims to git history before I caught them

**Root cause:** Inferring API design from a compile error message instead of reading the source. The compile error only says "too many arguments for Write" — it says nothing about whether Fingerprint/WriteVerified were removed from the package.

---

## e) WHAT WE SHOULD IMPROVE

1. **The verify-external-claims skill is not being loaded.** This is the FOURTH session it would have prevented the exact same class of error. The skill description triggers on "verify external claims" and "unverified claims." Every time we touch an external API, we should load it. We don't.

2. **Read vendored source before documenting APIs.** The source is RIGHT THERE in `vendor/github.com/larsartmann/go-atomic-write/atomicwrite.go`. 30 seconds. No excuses.

3. **Don't infer library-wide API changes from a single compile error.** "Write() dropped a parameter" ≠ "Fingerprint was removed from the package." These are completely different claims.

4. **The ROADMAP resolution pattern is dangerous.** Marking something "RESOLVED" without verification creates false confidence. Future sessions will trust the resolution and skip the verification entirely.

5. **Auto-git daemon commits false claims to history.** My first round of false documentation was committed before I caught it in self-review. The corrected version will be a separate commit. History now contains both the false claim AND the correction, with no clear signal which is authoritative.

6. **CHANGELOG.md still references v0.3.0 and the old Fingerprint/TOCTOU vocabulary** (line 14-15). Not updated this session.

---

## f) NEXT TASKS (up to 50)

### Immediate (this session's fallouts)

1. **Verify CONTRIBUTING.md** for stale Fingerprint/WriteVerified/TOCTOU claims — update if present
2. **Update CHANGELOG.md** — add v0.4.0 migration entry; fix line 14-15 stale references
3. **Re-run full `buildflow`** to confirm all 7 original failures are resolved
4. **Re-run `gomod-check`** to verify vendor/modules.txt consistency issues are resolved
5. **Re-run `nix-checker`** to confirm vendorHash is no longer flagged as stale
6. **Check flake.lock** — was auto-updated by nix build; verify it's in a good state

### go-atomic-write v0.4.0 API opportunities (NOW with correct knowledge)

7. **Evaluate `WriteIfChanged` for `configure`** — v0.4.0 added `WriteIfChanged(path, data) (bool, error)` which skips writes when content is unchanged. This would prevent spurious diffs/mtime bumps on re-runs of `configure`. Natural fit.
8. **Evaluate `WriteVerified` for `configure`** — Open question #2 in ROADMAP. Now we know the API exists, the evaluation is concrete.
9. **Document the full v0.4.0 API surface in AGENTS.md** — Currently only mentions `Write`. The library offers `Write`, `WriteVerified`, `WriteIfChanged`, `WriteFunc`, `WriteFuncVerified`, `FingerprintFile`, `FingerprintFromBytes`.

### Pre-existing buildflow findings (not from this session)

10. **Fix go-auto-upgrade findings (492)** — `lo.SliceToMap` idiomatic replacements in `cmd_analyze.go` and other files
11. **Fix go-structure-linter findings (18)** — Pin GitHub Actions to SHA commits instead of tags in `.github/workflows/ci.yml`
12. **Extract vendorHash to `vendorHash.nix`** — nix-checker suggestion for cleaner diffs
13. **Fix gomod-check: separate direct/indirect requires** in go.mod
14. **Run `golangci-lint`** explicitly to confirm no remaining findings after compile fix

### Documentation debt

15. **ROADMAP Open Question #1** — Adopt `linter-autoconfigure-sdk` for `validate`?
16. **ROADMAP Open Question #3** — Should `strict` and `recommended` profiles differ?
17. **ROADMAP Open Question #4** — testify to ginkgo/gomega migration policy
18. **ROADMAP Open Question #5** — Modularization proposal: execute or archive?
19. **ROADMAP Open Question #6** — Markdown or HTML for status reports?
20. **ROADMAP Open Question #7** — Fingerprint/TOCTOU in DOMAIN_LANGUAGE.md (now resolved: keep OUT as implementation detail)
21. **ROADMAP Open Question #2** — Should `configure` use `WriteVerified`? (Still open — needs evaluation)
22. **CHANGELOG.md audit** — Verify all entries are factually accurate against current code
23. **Update all historical status reports** that reference the "3-session deferral" — now resolved (but with a fuckup along the way)
24. **AGENTS.md: document `WriteIfChanged` as a potential future enhancement** for idempotent config writes

### Testing

25. **Add test for `writeConfig`** that verifies the atomic write behavior (temp file + rename, not raw os.WriteFile)
26. **Add test for `Configure()` with existing config** — verify it overwrites correctly with v0.4.0 `Write`
27. **Coverage threshold** — ROADMAP item: gate PRs on >=80%
28. **BDD tests for all commands** — ROADMAP item via bdd-testing skill

### Architecture

29. **Extract write logic into `pkg/config`** — ROADMAP item: `ConfigWriter` interface
30. **Typed errors across packages** — `detect`, `config`, `oxlint` still return generic `error`
31. **Shell completions** — Cobra completion subcommand
32. **Structured JSON logs** — `--log-format json` flag
33. **Custom output paths** — `--output` flexibility

### CI / Build

34. **Pin all GitHub Actions to SHAs** — 5+ actions in ci.yml use tag pins
35. **Add govulncheck to buildflow** — currently only in GitHub Actions CI
36. **Add coverage reporting** to CI
37. **Flake updates automation** — `nix flake update` on schedule

### Code Quality

38. **Run deduplicate-code skill** — check for duplication
39. **Run naming-review skill** — audit naming quality
40. **Run full-code-review skill** — comprehensive review
41. **Run data-model-review skill** — review types
42. **Check `cmd_report.go` stdversion warnings** — uses go1.27 APIs but project is go1.26
43. **Check `commands_test.go` stdversion warnings** — same issue
44. **Run architecture-review skill** — review modularity
45. **Run code-quality-scan skill** — full quality audit

### Domain / Rules

46. **Update rules_data.json** — check if oxlint has new rules since last update
47. **Update `TestRegistryTotal`** if rule count changed
48. **Verify `rules_version.txt`** matches installed oxlint version
49. **Review profile specs** — are severity decisions still optimal?
50. **Review restriction denylist** — are there new rules that should be denied?

---

## g) QUESTIONS (that I CANNOT figure out myself)

1. **Should `configure` use `WriteIfChanged` instead of `Write`?** v0.4.0 added `WriteIfChanged(path, data) (bool, error)` which skips writes when content is identical — no spurious diffs, no mtime bump, no file-watcher trigger. This seems strictly better for a config generator, but I don't know if you want `configure` to always touch the file (e.g., for "last generated" timestamp semantics).

2. **Should I fix the committed false claims in git history, or leave them with the correction commit?** The auto-git daemon committed my false "Fingerprint was removed" claims. I then corrected them. History now has both. I cannot rewrite history without `git rebase` (which violates safety rules), so the correction stays as a separate commit — unless you want something different.

3. **The `cmd_report.go` and `commands_test.go` files use `encoding/json` APIs that gopls flags as requiring go1.27, but `go.mod` says `go 1.26.5`.** Is the project targeting go1.27, or is this an accidental use of future APIs? I didn't touch these files but the warnings were visible in diagnostics.
