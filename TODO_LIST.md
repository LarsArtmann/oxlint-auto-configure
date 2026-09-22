# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> For shipped work, see `CHANGELOG.md`.
> Items here are verified, not assumed.

## Registry & Rules

| ID | Item                                                                                                                                                                                       | Evidence                                                               | Effort |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------- | ------ |
| R1 | Refresh embedded registry to current oxlint (`oxlint -f json --rules`), bump `rules_version.txt`, update `TestRegistryTotal`; runtime oxlint (1.82.x) is newer than the pinned 1.73.0 data | `pkg/rule/rules_version.txt`, `docs/status/2026-09-22_*.md` §c.4/§f.19 | M      |
| R2 | Pin the oxlint version in CI (`npm install -g oxlint` is unpinned; an upstream rule addition can break `TestRegistryTotal` silently)                                                       | `.github/workflows/ci.yml:42`                                          | S      |
| R3 | Add the `rules_version.txt` mismatch as a CI check (currently a runtime warning only)                                                                                                      | `internal/cli/cmd_configure.go` (warn path)                            | S      |

## External JS Plugins (@shadcn/lint) follow-ups

| ID | Item                                                                                                                                                                                                                                                                                                                                                                                                         | Evidence                                                               | Effort |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------- | ------ |
| S1 | Test `analyze` against a `jsPlugins` config (plugin installed + missing) and fix error surfacing if the oxlint failure becomes garbage findings                                                                                                                                                                                                                                                              | `internal/cli/cmd_analyze.go`, `docs/status/2026-09-22_*.md` §c.2/§f.1 | M      |
| S2 | External-plugin polish: single `oxlint --version` subprocess (currently `checkOxlintVersion` + `warnJsPluginsVersion` spawn twice); reuse the `package.json` read between `Detect()` and `DetectExternalPlugins()`; log "preserved N external plugins"; iterate the hint over all detected plugins (currently `detected[0]`); warn when preserving a jsPlugins entry whose package is no longer a dependency | `internal/cli/cmd_configure.go:199-236`, `pkg/detect/detector.go`      | M      |

## Release & CI hardening

| ID | Item                                                                                                                                                                                         | Evidence                                                                 | Effort |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ | ------ |
| C1 | Create `scripts/pre-release-check.sh` encoding the local gate + GoReleaser snapshot run (validated in the v0.5.0 session, never codified)                                                    | `docs/status/2026-09-11_12-22_*.md` §C1/C3, `2026-09-11_07-19_*.md` §c.8 | M      |
| C2 | Disabled-workflow canary: scheduled job (or doctor step) asserting every workflow in `.github/workflows/` has state `active` — the Jul–Sep billing block went unnoticed for ~2 months        | `docs/status/2026-09-11_12-22_*.md` §C7/E3                               | S      |
| C3 | Add a post-release smoke step inside `release.yml` (download own artifact + `--version`) so a broken release fails visibly                                                                   | `docs/status/2026-09-11_12-22_*.md` §f.25                                | M      |
| C4 | Pin `anchore/sbom-action/download-syft@v0` to a full SHA and pin `govulncheck@latest` in CI to a version (the two floating references left); drop the unused `HOMEBREW_TAP_GITHUB_TOKEN` env | `.github/workflows/release.yml:33,49`, `.github/workflows/ci.yml:79`     | S      |
| C5 | Migrate GoReleaser config off deprecated `brews`/`dockers` keys (`goreleaser check` warns)                                                                                                   | `.goreleaser.yaml`                                                       | M      |
| C6 | Sign container images with cosign + SBOM for the image (binaries are signed; the ghcr image is not)                                                                                          | `.goreleaser.yaml`, `docs/status/2026-09-11_07-19_*.md` §c.3             | M      |

## Public-repo hygiene

| ID | Item                                                                                                           | Evidence                           | Effort |
| -- | -------------------------------------------------------------------------------------------------------------- | ---------------------------------- | ------ |
| P1 | Add `SECURITY.md`, issue/PR templates, and `CODEOWNERS` (`.github/` holds only workflows + dependabot)         | `.github/`                         | S      |
| P2 | Run a full-history secret scan (gitleaks/trufflehog); only a working-tree grep ran before the repo went public | `docs/status/2026-09-09_*.md` §c.1 | S      |
| P3 | Delete the unused `SSH_PRIVATE_KEY` repo secret (repo settings; needs owner access)                            | `docs/status/2026-09-09_*.md` §c.3 | S      |

## Testing

| ID | Item                                                                                                                                                                                                                                                                                             | Evidence                                                                                           | Effort |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------- | ------ |
| T1 | Test the `ErrNotFound` branches added 2026-07-27: `configure` succeeds without oxlint in PATH (warn+continue); `analyze` fails with a clear error; restores `internal/cli` coverage toward 85%                                                                                                   | `internal/cli/cmd_configure.go:140`, `cmd_analyze.go:81`, `docs/status/2026-07-27_01-16_*.md` §c.1 | M      |
| T2 | Publish SDK v0.2.0 GitHub Release object (tag exists; `gh release create` never ran in the SDK repo); trigger pkg.go.dev fetch for the latest tag of both this repo and the SDK; run the clean-dir consumer test (`go get github.com/larsartmann/oxlint-auto-configure@v0.6.3` in a temp module) | `docs/status/2026-09-11_14-49_*.md` §b.2, `2026-09-11_12-22_*.md` §B4/C2                           | S      |

---

_Last reviewed: 2026-09-22_
