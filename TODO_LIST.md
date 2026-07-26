# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> For shipped work, see `CHANGELOG.md`.
> Items here are verified, not assumed.

## Build and Tooling

| Task                                                                                                                                                                       | Impact | Effort | Evidence                                                                 | Source                          |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------ | ------------------------------- |
| Update embedded rules from oxlint `1.59.0` to current (`1.73.0`): regenerate `rules_data.json`, bump `rules_version.txt`, update `TestRegistryTotal` count                 | High   | 30min  | `pkg/rule/rules_version.txt` says `1.59.0`; produces `WARN` on every run | 2026-07-26 f.11-13              |
| Resolve `go.mod` Go version mismatch: `go 1.26.4` triggers gopls `stdversion` warnings (`json.Marshal` requires go1.27). Bump to `go 1.27` or add `toolchain` directive | High   | 15min  | `go.mod:3`; active LSP `stdversion` warnings                             | 2026-07-22 c.2, 2026-07-26 f.15 |
| Fix `nix flake check` failure: `go 1.26.4` directive (changed in `ec08705`) breaks `mkPreparedSource` go-modules derivation ("go: updates to go.mod needed") | High   | 1h     | `nix build .#default` fails; `go build`/`go test`/`go vet` all pass        | `ec08705`, verified this session |
| Add CI check that `go mod vendor` produces no diff (root cause of 2026-07-17 BuildFlow failure)                                                                            | High   | 1h     | No such check in `.github/workflows/ci.yml`                              | 2026-07-22 c.3, e.1             |
| Establish reproducible `golangci-lint` baseline: local reports ~116 issues (depguard, varnamelen, mnd, tagliatelle, err113, forbidigo) while CI passes                     | Med    | 2h     | `FEATURES.md` gap; `.golangci.yml` exists but local/CI mismatch          | 2026-07-22 d.2                  |
| Add BuildFlow to CI so the full local workflow runs on every PR                                                                                                            | Med    | 1h     | Not in CI                                                                | 2026-07-22 c.4, f.5             |
| Decide whether to add `gosec` to the CI security job                                                                                                                       | Low    | 15min  | `.github/workflows/ci.yml` has govulncheck but not gosec                 | 2026-07-22 c.13                 |

## Testing

| Task                                                                                                                     | Impact | Effort | Evidence                                                          | Source                          |
| ------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | ----------------------------------------------------------------- | ------------------------------- |
| Add entry-point tests for `cmd/oxlint-auto-configure/main.go` (currently 0% coverage)                                    | High   | 1h     | `go test -cover ./...` shows 0.0% for `cmd/oxlint-auto-configure` | 2026-07-22 c.8, f.10            |
| Add E2E integration test: `configure` -> parse output -> verify round-trip via `config.FromJSON`                         | Med    | 2h     | No such test exists                                               | 2026-07-22 c.9, 2026-07-26 f.18 |
| Add dedicated test for atomic-write contract: verify `writeConfig` leaves no `.tmp` files and always produces valid JSON | Med    | 1h     | No `atomicwrite`/`.tmp` test found in `*_test.go`                 | 2026-07-26 c.2, f.2-3           |
| Wire `flake.nix` ldflags for `commit`, `date`, and `builtBy` so nix builds show full version metadata                    | Low    | 30min  | `flake.nix` only injects `version`; rest default to `unknown`     | 2026-07-22 c.10                 |
| Increase `internal/cli` test coverage from 74.3% toward 85%+                                                             | Low    | 3h     | `go test -cover ./...` shows 74.3%                                | 2026-07-22 c.11                 |

## Maintenance

| Task                                                                                                                       | Impact | Effort | Evidence                                                                              | Source                            |
| -------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------------------- | --------------------------------- |
| Run `hierarchical-errors` skill and baseline findings                                                                      | Low    | 1h     | Not yet run as a dedicated pass                                                       | 2026-07-22 c.6                    |
| Run dedicated skill passes (`naming-review`, `code-quality-scan`, `deduplicate-code`) and baseline findings across codebase | Low    | 3h     | Listed in multiple reports; never executed                                             | 2026-07-22 f.15, 2026-07-26 f.18  |
| Fix broken `#resolution` anchor links in `docs/status/` (07-17 `.md` + 2 `.html` files) — GitHub generates `#resolution-2026-07-22`, not `#resolution` | Low    | 20min  | `grep -rn '#resolution' docs/status/` shows 3 broken links; headings have date suffix | 2026-07-26 f.3, d.1              |

---

_Last reviewed: 2026-07-26_
