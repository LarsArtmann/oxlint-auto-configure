# AGENTS.md - oxlint-auto-configure

## Project Overview

**oxlint-auto-configure** is a Go CLI that generates optimal `.oxlintrc.json` configurations for maximum type safety.

### SCOPE BOUNDARY — READ THIS FIRST

**This tool generates `.oxlintrc.json`. Nothing else.**

- ✅ **Our job:** Detect project type → pick profile → generate config → validate config
- ❌ **NOT our job:** Running oxlint, auto-fixing code, enforcing lint rules, replacing oxlint
- The `analyze` command is a **diagnostic aid** to help decide which profile to use — not a linter replacement
- `pkg/oxlint/fix.go` and `configure --fix` exist for convenience but are **not core purpose**
- The go-finding pipeline's fix+verify loop stays `DryRun=true` — we detect and report, we do not fix
- Do NOT add features that duplicate oxlint's job (auto-fix, watch mode, enforcement)

### Core Purpose

Oxlint has 716 rules across 7 categories and 15 plugins. Only 108 are enabled by default. This tool:

1. Discovers project type (React, Next.js, Vue, etc.)
2. Enables relevant plugins automatically
3. Sets optimal severity for every rule based on chosen profile
4. Generates a ready-to-use `.oxlintrc.json`

### Key Files

| File                                | Purpose                                                                                |
| ----------------------------------- | -------------------------------------------------------------------------------------- |
| `pkg/rule/rule.go`                  | Core types: Rule, Category, Plugin, FixCapability, SeverityDecision                    |
| `pkg/rule/registry.go`              | Rule registry loaded from embedded JSON (716 rules)                                    |
| `pkg/rule/rules_data.json`          | Embedded oxlint rules data (from `oxlint -f json --rules`)                             |
| `pkg/profile/profile.go`            | Profile presets, Categorizer engine, `DecideCategory()`, PluginConfig                  |
| `pkg/config/generator.go`           | .oxlintrc.json generator                                                               |
| `pkg/detect/detector.go`            | Project type detection from package.json                                               |
| `pkg/diff/differ.go`                | Config before/after comparison (all fields: plugins, categories, rules, env, settings) |
| `pkg/format/format.go`              | Rendering: FindingView, SummaryView, PrintSummary/PrintFindingsJSON/PrintFindingsTable |
| `pkg/oxlint/detector.go`            | go-finding Detector for oxlint; `Runner` interface seam                                |
| `pkg/oxlint/version.go`             | oxlint version check and binary verification                                           |
| `pkg/oxlint/fix.go`                 | oxlint --fix wrapper                                                                   |
| `internal/cli/cmd_root.go`          | Root command, shared constants (defaultConfigPath, defaultProfile, version)            |
| `internal/cli/cmd_configure.go`     | configure command + extracted `Configure(ctx, absRoot, opts)`                          |
| `internal/cli/cmd_analyze.go`       | analyze command with go-finding pipeline integration                                   |
| `internal/cli/cmd_validate.go`      | validate command                                                                       |
| `internal/cli/cmd_report.go`        | report command + format helpers (JSON, table, summary)                                 |
| `cmd/oxlint-auto-configure/main.go` | Entry point                                                                            |

### Nix

The project has a `flake.nix` for reproducible builds and dev shells.

```bash
nix build .                    # Build binary (tests run, oxlint included)
nix run . -- configure .      # Run with oxlint in PATH
nix develop .                  # Dev shell: go, oxlint, gopls, golangci-lint
```

- **Vendored deps** — `vendor/` committed for nix sandbox compatibility (private go-finding dep can't be fetched with `GOPROXY=off`)
- **`GOWORK=off go mod vendor`** — Re-vendor after `go.mod` changes, then commit (required for nix sandbox build)
- **Runtime dep** — `oxlint` is a runtime dependency; wrapped in `nix run` via `makeWrapper`
- **Git-derived version** — `self.rev or self.dirtyRev or "dev"` injected via ldflags
- **`lib.fileset`** — Precise source filtering (go.mod, go.sum, cmd/, internal/, pkg/, vendor/)
- **`nix-systems/default`** — System list via flake input instead of hardcoded
- **Checks** — `build` and `test` (reuses goModules from package)
- **Formatter** — `nixfmt` via `formatter` output

### Testing

```bash
just test        # Run tests with -race (pkg + internal)
just cover       # Coverage report
just vet         # Run go vet
just check       # All checks (fmt + vet + lint + test)
```

### Dependencies

- `github.com/larsartmann/go-finding` v1.0.0 — Unified static analysis model (private: `GOPRIVATE=github.com/LarsArtmann/*`; branded types `RuleName`/`ToolName`/`ID` in `NewFinding`)
- `github.com/spf13/cobra` — CLI framework
- `github.com/stretchr/testify` — Test assertions

### Design Principles

1. **Embedded rules data** — Rules are embedded via `go:embed` for zero-dependency startup
2. **Profile-driven** — All severity decisions flow from `DecideCategory()`; `Description()` derived from same source
3. **Project-aware** — Auto-detects frameworks to enable relevant plugins
4. **go-finding integration** — Detection and reporting only (DryRun=true, no auto-fix); Detector interface; FixStrategy from registry (for reporting, not for fixing); Range from oxlint labels; structured FindingError; Metrics/Retry/Callbacks; Report.PrettyJSON/ToSARIF/ToSARIFFiltered; SortByPosition/ActiveFindings; finding.Filter for --severity
5. **Config round-trip** — Generated configs can be parsed back and compared
6. **Self-describing types** — Plugin has `CLIFlag()`/`NeedsFlag()`; Registry has generic `Filter()`
7. **Decoupled rendering** — `pkg/format` accepts plain view structs, not go-finding types
8. **Testable commands** — `Configure()` extracted from cobra closure; independently callable

### Profiles

| Profile            | Description                                                   |
| ------------------ | ------------------------------------------------------------- |
| `maximal-typesafe` | ALL rules at error (nursery at warn)                          |
| `recommended`      | Correctness+suspicious+TS at error, rest at warn, nursery off |
| `strict`           | Correctness+suspicious at error, rest at warn, nursery off    |
| `minimal`          | Only correctness at error, rest uses defaults                 |

### Updating Rules

When oxlint adds new rules:

```bash
just update-rules  # Runs: oxlint -f json --rules > pkg/rule/rules_data.json
```

Then update `TestRegistryTotal` in `pkg/rule/registry_test.go` with the new count.

### Important Gotchas

- **Private go-finding** — `GOPRIVATE=github.com/LarsArtmann/*` required; v1.0.0 from GitHub (no local replace)
- **Plugin naming** — `FullName()` adds plugin prefix for all non-ESLint rules (e.g., `typescript/no-floating-promises`)
- **Oxlint config format** — Uses `categories` for category-level severity + `rules` for per-rule overrides
- **Version injected at build** — `internal/cli.version` via ldflags (default: "dev")
- **Per-command files** — Commands are in `internal/cli/cmd_*.go`, not a monolithic file
- **SARIF output** — analyze command defaults to summary format; SARIF is opt-in via `-f sarif`
- **Structured logging** — CLI uses `log/slog` for all diagnostic output with --verbose/--quiet flags
- **Global Flags** — `-v/--verbose` and `-q/--quiet` on root command control log level
- **PluginConfig** — `map[rule.Plugin]bool` (not a struct with bool fields)
- **Self-describing Plugin** — `CLIFlag()` and `NeedsFlag()` methods on Plugin type
- **Version-pinned rules** — `rules_version.txt` embedded; warns on mismatch
- **DecideCategory** — Returns `(SeverityDecision, bool)`; bool controls whether category appears in config (false = omit)
- **Differ completeness** — Compares all fields: Plugins, Categories, Rules, Env, Settings
- **Runner seam** — `pkg/oxlint.Runner` interface; `realRunner` (production), `mockRunner` (tests)
- **Configure extraction** — `Configure(ctx, absRoot, opts)` callable without cobra; malformed existing configs now log warnings
- **go-finding branded types** — `Finding.Rule` is `finding.RuleName`, `Finding.ToolName` is `finding.ToolName`. `NewFinding` requires branded conversions: `finding.RuleName(s)`, `finding.ToolName(s)`. Use `string(f.Rule)` when assigning to plain-string fields (e.g., `FindingView.Rule`).
- **go-finding PipelineResult.Stable** — `Stable()` is a method, not a field. Use `Reason: pipeline.ReasonStable` in struct literals.
- **go-finding Finding enrichment** — Detector populates Range (from oxlint label spans), FixStrategy (from registry FixCapability), Tag (plugin name), Snippet (label text), Metadata (url), Suggestion (help)
- **go-finding FindingError** — All detector errors use structured `finding.NewIOError`/`NewParseError` for `errors.Is()` support
- **Pipeline config** — Analyze command wires Metrics, Retry (2 retries, 100ms base), OnFinding/OnIteration callbacks; oxlint version in ToolInfo
- **Analyze formats** — `summary`, `json` (flat FindingView array), `report` (full go-finding Report JSON), `sarif`, `table`
- **WithRegistry option** — `oxlint.WithRegistry(reg)` enables FixStrategy lookup per-finding
- **Nix build** — `vendor/` committed (required for nix sandbox); `vendorHash = null` in flake; `GOWORK=off` for all go commands
- **Severity filter** — Analyze `-s/--severity` flag uses `finding.Filter(BySeverityAtLeast)` for json/table; `ToSARIFFiltered` for SARIF
- **Profile name dedup** — `profile.AllProfileNames()` is single source; no more `cliProfileNames`/`config.profileNames`
- **Detect logging** — `pkg/detect` logs warnings on malformed package.json (but not missing — that's normal for Go projects)
- **Summary enrichment** — `SummaryView.ByFixStrategy` populated from `Report.Summary.ByFixStrategy`
- **SummaryView.Findings** — Carries `[]FindingView` for top-rules/top-files computation in `PrintSummary`
- **Iteration logging** — OnIteration callback uses `slog.Debug` (only visible with `-v`); no more raw slog spam
- **ProjectTypeTest** — `vitest`/`jest` now detected as `ProjectTypeTest` (not `ProjectTypeNode`); enables both `PluginNode` AND the correct test plugin via `depPluginRules`
- **GOWORK=off** — Parent workspace at `/home/lars/projects/go.work` interferes; always use `GOWORK=off` for `go run`/`go test`
- **Test boilerplate is intentional** — `t.Parallel()` followed by `reg := loadTestRegistry(t)` in every test is idiomatic Go and **not** a duplication violation; `paralleltest` linter requires `t.Parallel()` directly in the test function. Do not fold it into the helper.
- **marshalConfigJSON helper** — `internal/cli/cmd_configure.go` extracts the shared `cfg.ToJSON()` + error wrap into `marshalConfigJSON`. The remaining 4-line preamble (`data, err := marshalConfigJSON(cfg); if err != nil { return err }`) in `writeConfig`/`writeDryRun` is idiomatic Go error propagation and intentionally not abstracted further.
- **`slices.Sorted(maps.Keys(m))`** — Prefer over manual `make+loop+sort.Strings` for sorted map-key enumeration in Go 1.23+.

---

_Assisted-by: Crush <crush@charm.land>_
