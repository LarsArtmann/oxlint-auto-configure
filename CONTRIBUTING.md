# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Prerequisites

### Option A: Nix (recommended)

Install [Nix](https://nixos.org/) with flakes enabled. Everything else is handled by `flake.nix`:

```bash
nix develop .    # Enter dev shell: go, oxlint, gopls, golangci-lint
```

No manual environment variable setup needed — the dev shell sets `GOPRIVATE`, `GOWORK=off`, and `GOEXPERIMENT=jsonv2` automatically.

### Option B: Manual Go setup

Requires Go 1.26+ and [oxlint](https://oxc.rs/) in PATH.

Three environment variables are **required** for all `go` commands:

| Variable | Value | Why |
| --- | --- | --- |
| `GOEXPERIMENT` | `jsonv2` | The codebase uses `encoding/json/v2`, still behind an experiment gate in Go 1.26. |
| `GOWORK` | `off` | A parent `go.work` at `/home/lars/projects/go.work` interferes; disable it. |
| `GOPRIVATE` | `github.com/larsartmann/*,github.com/LarsArtmann/*` | Private dependencies (`go-finding`, `go-atomic-write`, `gogenfilter`) require Git access. |

```bash
export GOEXPERIMENT=jsonv2
export GOWORK=off
export GOPRIVATE="github.com/larsartmann/*,github.com/LarsArtmann/*"
```

## Private Dependencies

This project depends on three private `github.com/LarsArtmann/*` repositories:

| Dependency | Purpose |
| --- | --- |
| `go-finding` v1.3.0 | Unified static analysis model used by the `analyze` command |
| `go-atomic-write` v0.3.0 | Crash-durable atomic file writes (temp + fsync + atomic rename) |
| `gogenfilter` | Code generation filter |

Git authentication (SSH key or token) for `github.com/LarsArtmann` is required to fetch them. The nix dev shell and build handle this via `GOPRIVATE`; the nix build sandbox injects them as local `replace` directives via `mkPreparedSource`.

## Atomic Write

Config output (`.oxlintrc.json`) is written via `atomicwrite.Write` from `go-atomic-write` v0.3.0 — temp file + fsync + atomic rename. A crash mid-write cannot truncate the user's existing config. Never use raw `os.WriteFile` for config output. See `internal/cli/cmd_configure.go` (`writeConfig`).

## Build, Test, and Lint

### Via Nix (canonical)

```bash
nix build .              # Build binary (runs tests, includes oxlint)
nix flake check .        # All checks: build, test, format
nix run . -- configure   # Run the tool (oxlint in PATH)
```

### Via Go (when iterating)

```bash
# All Go commands need GOEXPERIMENT=jsonv2 and GOWORK=off (see above)

GOWORK=off GOEXPERIMENT=jsonv2 go test -race ./...    # Tests with race detector
GOWORK=off GOEXPERIMENT=jsonv2 go test -cover ./...   # Coverage report
GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...           # Go vet
GOWORK=off GOEXPERIMENT=jsonv2 golangci-lint run ./...  # Linter
```

> **Note:** Local `golangci-lint` may report issues that CI does not, due to a version/config mismatch. CI is the source of truth for linting.

## Vendoring and vendorHash

`vendor/` is gitignored and regenerated locally. The nix build re-vendors via `buildGoModule` in a sandbox.

### After changing go.mod dependencies

1. Update `go.mod`/`go.sum` (add, update, or remove a dependency).
2. Regenerate local vendor directory:
   ```bash
   GOWORK=off go mod vendor
   ```
3. Rebuild via nix:
   ```bash
   nix build .#default
   ```
4. If nix reports a `vendorHash` mismatch, copy the `got:` sha256 from the error output.
5. Paste it into the `vendorHash` field in `flake.nix`:
   ```nix
   vendorHash = "sha256-<paste-the-got-hash-here>";
   ```
6. Run `nix build .#default` again to confirm it passes.

The `vendorHash` comment in `flake.nix` documents this workflow inline.

## Updating Embedded Rules

When oxlint adds new rules, regenerate the embedded data:

```bash
oxlint -f json --rules > pkg/rule/rules_data.json
```

Then update `TestRegistryTotal` in `pkg/rule/registry_test.go` with the new count, and update `pkg/rule/rules_version.txt` with the oxlint version.

## CI

CI runs three jobs on every push/PR (`.github/workflows/ci.yml`):

- **test** — `go test -race ./...`
- **security** — `govulncheck`
- **lint** — `golangci-lint run ./...`

CI is the source of truth. If local results differ from CI, the CI result wins.

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
