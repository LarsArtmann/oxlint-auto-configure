# AGENTS.md - oxlint-auto-configure

## Project Overview

**oxlint-auto-configure** is a Go CLI that automatically generates optimal `.oxlintrc.json` configurations for maximum type safety. It uses [go-finding](https://github.com/larsartmann/go-finding) as its core library for the detect→triage→fix→verify pipeline.

### Core Purpose

Oxlint has 716 rules across 7 categories and 15 plugins. Only 108 are enabled by default. This tool:

1. Discovers project type (React, Next.js, Vue, etc.)
2. Enables relevant plugins automatically
3. Sets optimal severity for every rule based on chosen profile
4. Generates a ready-to-use `.oxlintrc.json`

### Key Files

| File | Purpose |
|------|---------|
| `pkg/rule/rule.go` | Core types: Rule, Category, Plugin, FixCapability, SeverityDecision |
| `pkg/rule/registry.go` | Rule registry loaded from embedded JSON (716 rules) |
| `pkg/rule/rules_data.json` | Embedded oxlint rules data (from `oxlint -f json --rules`) |
| `pkg/profile/profile.go` | Profile presets, Categorizer engine, PluginConfig |
| `pkg/config/generator.go` | .oxlintrc.json generator |
| `pkg/detect/detector.go` | Project type detection from package.json |
| `pkg/diff/differ.go` | Config before/after comparison |
| `pkg/oxlint/detector.go` | go-finding Detector for oxlint |
| `pkg/oxlint/version.go` | oxlint version check and binary verification |
| `pkg/oxlint/fix.go` | oxlint --fix wrapper |
| `internal/cli/cmd_root.go` | Root command, shared constants (defaultConfigPath, defaultProfile, version) |
| `internal/cli/cmd_configure.go` | configure command + helpers (writeDryRun, profileNames) |
| `internal/cli/cmd_analyze.go` | analyze command with go-finding pipeline integration |
| `internal/cli/cmd_validate.go` | validate command |
| `internal/cli/cmd_report.go` | report command + format helpers (JSON, table, summary) |
| `cmd/oxlint-auto-configure/main.go` | Entry point |

### Testing

```bash
just test        # Run tests with -race (pkg + internal)
just cover       # Coverage report
just vet         # Run go vet
just check       # All checks (fmt + vet + lint + test)
```

### Dependencies

- `github.com/larsartmann/go-finding` — Unified static analysis model (local replace)
- `github.com/spf13/cobra` — CLI framework
- `github.com/stretchr/testify` — Test assertions

### Design Principles

1. **Embedded rules data** — Rules are embedded via `go:embed` for zero-dependency startup
2. **Profile-driven** — All severity decisions flow from the chosen profile
3. **Project-aware** — Auto-detects frameworks to enable relevant plugins
4. **go-finding integration** — Uses Detector interface for oxlint integration
5. **Config round-trip** — Generated configs can be parsed back and compared
6. **Self-describing types** — Plugin has `CLIFlag()`/`NeedsFlag()`; Registry has generic `Filter()`

### Profiles

| Profile | Description |
|---------|-------------|
| `maximal-typesafe` | ALL rules at error (nursery at warn) |
| `recommended` | Correctness+suspicious+TS at error, rest at warn, nursery off |
| `strict` | Correctness+suspicious at error, rest at warn, nursery off |
| `minimal` | Only correctness at error, rest uses defaults |

### Updating Rules

When oxlint adds new rules:

```bash
just update-rules  # Runs: oxlint -f json --rules > pkg/rule/rules_data.json
```

Then update `TestRegistryTotal` in `pkg/rule/registry_test.go` with the new count.

### Important Gotchas

- **Local replace** for go-finding — `go.mod` has `replace` directive pointing to `/home/lars/projects/go-finding`
- **Plugin naming** — `FullName()` adds plugin prefix for all non-ESLint rules (e.g., `typescript/no-floating-promises`)
- **Oxlint config format** — Uses `categories` for category-level severity + `rules` for per-rule overrides
- **Version injected at build** — `internal/cli.version` via ldflags (default: "dev")
- **Per-command files** — Commands are in `internal/cli/cmd_*.go`, not a monolithic file
- **SARIF output** — analyze command defaults to summary format; SARIF is opt-in via `-f sarif`
- **Structured logging** — CLI uses `log/slog` for all diagnostic output (not fmt.Fprintf)
- **PluginConfig** — `map[rule.Plugin]bool` (not a struct with bool fields)
- **Self-describing Plugin** — `CLIFlag()` and `NeedsFlag()` methods on Plugin type
- **Version-pinned rules** — `rules_version.txt` embedded; warns on mismatch

---

_Assisted-by: Crush <crush@charm.land>_
