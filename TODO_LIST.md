# TODO List

Short-term actionable work for oxlint-auto-configure.

Done items live in `CHANGELOG.md`, not here.

## Build & Tooling

- [ ] Resolve `go.mod` Go version mismatch with `encoding/json/v2` (gopls warnings). Options: bump to `go 1.27` or add a `toolchain` directive.
- [ ] Add CI check that `go mod vendor` produces no diff.
- [ ] Add BuildFlow to CI so the full local workflow runs on every PR.
- [ ] Run `golangci-lint run ./...` directly and record a baseline of findings.
- [ ] Run `hierarchical-errors` directly and record a baseline of findings.
- [ ] Investigate BuildFlow step-count discrepancy (42 vs 35) from the 2026-07-17 session.

## Code Quality

- [ ] Add tests for `cmd/oxlint-auto-configure/main.go` (entry point currently at 0% coverage).
- [ ] Add E2E integration test: configure → validate → report round-trip.
- [ ] Wire `flake.nix` ldflags for `commit`, `date`, and `builtBy` so nix builds show full version metadata.
- [ ] Increase `internal/cli` test coverage from ~74% toward 85%+.

## Documentation

- [ ] Create `ROADMAP.md` for long-term direction.
- [ ] Keep `README.md` and `AGENTS.md` in sync after every dependency or feature change.

## Decisions Needed

- [ ] Decide testify → ginkgo/gomega migration policy for this project.
- [ ] Decide whether to execute or archive the modularization proposal (docs already deleted; decision remains).
- [ ] Decide whether to add `gosec` to the CI security job.

---

Last reviewed: 2026-07-22
