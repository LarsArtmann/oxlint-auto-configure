# PUBLIC OR PRIVATE? — oxlint-auto-configure

**Decision Date:** 2026-05-04
**Current Status:** Private
**Recommendation:** **Make public CONDITIONALLY** (see prerequisites)

---

## Verdict

**Make public after resolving the private dependency blocker.** This is a high-quality, genuinely useful tool with no real competitors, clean code, and no sensitive data. The only hard blocker is the private `go-finding` dependency — once resolved, this project should be public.

---

## Project Summary

| Aspect | Detail |
|--------|--------|
| **What** | Go CLI that generates optimal `.oxlintrc.json` for oxlint |
| **Maturity** | 84 commits, ~1 week old, actively developed |
| **Code quality** | Clean architecture, tests, CI, nix flakes, linting |
| **License** | MIT |
| **Unique?** | No direct competitor exists (Sourcegraph search confirms) |

---

## PRO — Make It Public

### 1. Solves a real, unsolved problem

Oxlint has 716 rules, only 108 enabled by default. No tool exists to auto-configure them. This fills a genuine gap in the oxc ecosystem.

### 2. High code quality

- Clean Go architecture with proper separation of concerns (`pkg/rule`, `pkg/profile`, `pkg/config`, `pkg/detect`, `pkg/diff`, `pkg/format`, `pkg/oxlint`)
- Test suite with `-race` flag, coverage reports, `golangci-lint` zero-errors
- Reproducible builds via `flake.nix`
- CI with SSH-based private dep resolution
- Proper semantic versioning via ldflags
- Well-documented README with profiles, commands, architecture

### 3. No sensitive data

- No secrets, tokens, API keys, or private URLs anywhere in the codebase
- `.gitignore` properly excludes IDE files, coverage, build artifacts
- `.gitattributes` marks `vendor/` as `linguist-generated`
- CI uses GitHub Secrets for SSH keys (not committed)
- No internal infrastructure references

### 4. Ecosystem alignment

- MIT license — fully open source friendly
- References `golangci-lint-auto-configure` as a sibling project (same pattern, different linter)
- Could attract contributions from the growing oxc/oxlint community
- `nix run github:larsartmann/oxlint-auto-configure` already documented in README (expecting public)

### 5. Marketing & credibility

- Public portfolio piece demonstrating Go architecture, nix, CI/CD, testing
- First-mover advantage in the oxlint configuration space
- README already references the repo as if it's public (`nix run github:larsartmann/...`)

### 6. No competitive risk

- The "secret sauce" is the profile decision engine + rule categorization — not something competitors can steal meaningful value from
- The value is in *maintenance* (keeping rules_data.json updated with oxlint releases) — that requires ongoing effort regardless of visibility

### 7. Vendored go-finding code is NOT committed

- `git ls-tree` shows 0 files from `go-finding` in the committed vendor tree
- The vendor directory is stale/inconsistent locally but not in git
- No private library source code would be leaked

---

## CONTRA — Keep It Private

### 1. HARD BLOCKER: Private `go-finding` dependency

`go-finding` (v0.3.0) is a **private** repository. This means:

- `go install github.com/larsartmann/oxlint-auto-configure@latest` will **fail** for anyone without access
- README already acknowledges this: "requires access to the private go-finding dependency"
- The nix method works (vendored) but the Go method won't

**Resolution options:**
1. **Best:** Make `go-finding` public too (it's a generic static analysis model — no reason to be private)
2. **Workaround:** Replace `go-finding` with a minimal internal implementation (only uses a subset: Finding, Report, SARIF, Filter, Severity)
3. **Acceptable:** Document nix-only installation, accept that `go install` is restricted

### 2. Very young project (1 week old)

- Created 2026-04-28, only 84 commits
- No external users yet, no issues/PRs from community
- API might change significantly

### 3. README references nix run as if public

The README already uses `github:larsartmann/oxlint-auto-configure` URLs that only work for public repos — this is either intentional (preparing for public) or an oversight.

### 4. Vendor directory inconsistency

- `go.mod` requires `go-finding v0.3.0` but `vendor/modules.txt` has `v0.2.0`
- This is a known issue that needs fixing before public release (`just vendor`)

### 5. AGENTS.md is committed

- Contains detailed internal architecture notes, design decisions, and gotchas
- Useful for AI assistants but also exposes internal reasoning
- **Mitigation:** AGENTS.md is harmless — it's documentation, not secrets. Many public projects have similar files.

### 6. `docs/status/` contains session reports

- Detailed status reports from development sessions
- References Crush (AI assistant) usage
- Not harmful but reveals development workflow

---

## Conditions for Going Public

### Must (hard blockers)

- [ ] **Resolve go-finding dependency** — Either make `go-finding` public, or remove the dependency, or accept nix-only installation
- [ ] **Fix vendor inconsistency** — Run `just vendor` to sync `go.mod` v0.3.0 with vendor directory
- [ ] **Verify no private references** — Audit commit history for any accidentally committed secrets or private URLs

### Should (quality gates)

- [ ] **Add GitHub description** — Currently empty: `""`
- [ ] **Add topics/tags** — `oxlint`, `oxc`, `linter`, `configuration`, `go`, `nix`
- [ ] **Clean up stale docs** — `docs/status/` session reports are development artifacts; consider `.gitignore` or move to wiki
- [ ] **Verify CI works publicly** — CI currently uses SSH key for private go-finding; if made public without go-finding public, CI needs adjustment

### Nice to have

- [ ] Add CONTRIBUTING.md
- [ ] Add issue templates
- [ ] Add release tags / GitHub release
- [ ] Set up branch protection rules

---

## Decision Matrix

| Factor | Public | Private |
|--------|--------|---------|
| Community adoption | +++ | — |
| Portfolio value | +++ | — |
| Ecosystem contribution | +++ | — |
| Maintenance burden | ++ (issues/PRs) | — |
| Dependency complexity | — (private dep) | + |
| Competitive exposure | negligible | + |
| First-mover advantage | +++ | — |
| **Net** | **Strong positive** | Neutral |

---

## Final Recommendation

> **Make public, but resolve the `go-finding` dependency first.**
>
> The strongest path: make `go-finding` public too. It's a generic static analysis data model (Finding, Report, SARIF, Severity, Filter) — there's no competitive advantage in keeping it private. Making both public creates a coherent ecosystem and enables the `go install` path.
>
> If `go-finding` must stay private, the nix-only installation path works for public distribution — but document this clearly and expect a smaller user base.
>
> This project has no secrets, no sensitive data, no real competitive risk, and fills an unsolved gap. There is no good reason to keep it private long-term.

---

_Analyzed by Crush <crush@charm.land>_
