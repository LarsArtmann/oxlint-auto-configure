# oxlint-auto-configure

**Generate the optimal `.oxlintrc.json` — not a linter, a configurator.**

This tool's one job: inspect your project and write the best possible [oxlint](https://oxc.rs/docs/guide/usage/linter.html) config. It does **not** lint, fix, or replace oxlint — it configures oxlint so you don't have to.

## Why

Oxlint has **716 rules** across **7 categories** and **15 plugins**. Only 108 are enabled by default. Manually configuring each rule for maximum type safety is tedious and error-prone. This tool automates the entire process:

- Discovers your project type (React, Next.js, Vue, etc.)
- Enables relevant plugins automatically
- Sets optimal severity for every rule based on your chosen profile
- Generates a ready-to-use `.oxlintrc.json`

> **Scope boundary:** This tool generates `.oxlintrc.json`. Running oxlint, auto-fixing code, and enforcing lint rules are oxlint's job. The `analyze` command is a diagnostic aid to help you decide which profile to use — not a replacement for running oxlint itself.

## Installation

### Nix (recommended)

```bash
# Build and run directly (oxlint included)
nix run github:larsartmann/oxlint-auto-configure -- configure

# Or install to your nix profile
nix profile install github:larsartmann/oxlint-auto-configure

# Dev shell with go, oxlint, gopls, golangci-lint
nix develop github:larsartmann/oxlint-auto-configure
```

### Go

```bash
export GOPRIVATE=github.com/LarsArtmann/*
go install github.com/larsartmann/oxlint-auto-configure/cmd/oxlint-auto-configure@latest
```

Requires Go 1.26+ and [oxlint](https://oxc.rs/docs/guide/usage/linter.html) in PATH.

> **Note:** The Go install method requires access to the private
> [go-finding](https://github.com/larsartmann/go-finding) dependency.
> The nix method works without any Go setup or Git authentication.
>
> **Note:** When building or running from source, set `GOEXPERIMENT=jsonv2`.
> The project uses `encoding/json/v2`, which is still behind the jsonv2 experiment
> in Go 1.26. Nix builds and dev shells set this automatically.

## Quick Start

```bash
# Generate config with recommended profile (default)
oxlint-auto-configure configure

# Maximum type safety — ALL rules at 'error'
oxlint-auto-configure configure --profile maximal-typesafe

# Dry run to see what would change
oxlint-auto-configure configure --dry-run

# Validate your existing config
oxlint-auto-configure validate

# Analyze project and show findings (SARIF output)
oxlint-auto-configure analyze

# Report of all 716 rules and their recommended severity
oxlint-auto-configure report
```

## Profiles

| Profile              | Correctness | Suspicious | Style   | Perf    | Pedantic | Restriction | Nursery |
| -------------------- | ----------- | ---------- | ------- | ------- | -------- | ----------- | ------- |
| **maximal-typesafe** | error       | error      | error   | error   | error    | error       | warn    |
| **strict**           | error       | error      | warn    | warn    | warn     | warn        | off     |
| **recommended**      | error       | error      | warn    | warn    | warn     | warn        | off     |
| **minimal**          | error       | default    | default | default | default  | default     | default |

### Profile Details

- **maximal-typesafe**: Every single rule at `error`. Maximum type safety and correctness enforcement. Even nursery rules at `warn`.
- **strict**: Core correctness at `error`, everything else at `warn` except nursery.
- **recommended** (default): Correctness + suspicious at `error`. Style, perf, pedantic, and restriction at `warn`. Nursery off.
- **minimal**: Only correctness at `error`. Everything else uses oxlint defaults.

## Project Detection

The tool auto-detects your project type and enables relevant plugins:

| Detected   | Plugins Enabled                             |
| ---------- | ------------------------------------------- |
| React      | `react`, `jsx-a11y`, `react-perf`           |
| Next.js    | `nextjs`, `react`, `jsx-a11y`, `react-perf` |
| Vue        | `vue`                                       |
| Jest       | `jest`, `node`                              |
| Vitest     | `vitest`, `node`                            |
| TypeScript | `typescript` (always on)                    |

## Commands

### `configure`

Generate the optimal `.oxlintrc.json`:

```bash
oxlint-auto-configure configure [flags]
```

| Flag            | Default          | Description                           |
| --------------- | ---------------- | ------------------------------------- |
| `-p, --profile` | `recommended`    | Configuration profile                 |
| `-c, --config`  | `.oxlintrc.json` | Output config file path               |
| `-d, --dry-run` | false            | Show changes without writing          |
| `--fix`         | false            | Run oxlint --fix after writing config |
| `--root`        | `.`              | Project root directory                |

### `analyze`

Run oxlint and show findings using the go-finding pipeline:

```bash
oxlint-auto-configure analyze [--root .] [-f summary|json|report|sarif|table]
```

### `validate`

Validate an existing `.oxlintrc.json`:

```bash
oxlint-auto-configure validate [-c .oxlintrc.json]
```

### `report`

Generate a report of all rules and recommended severities:

```bash
oxlint-auto-configure report [-p recommended] [-f table|json|summary]
```

## Rule Statistics

| Category    | Count   | Default         |
| ----------- | ------- | --------------- |
| Correctness | 216     | 108 enabled     |
| Style       | 211     | mostly disabled |
| Pedantic    | 115     | disabled        |
| Restriction | 91      | disabled        |
| Suspicious  | 52      | disabled        |
| Perf        | 13      | disabled        |
| Nursery     | 18      | disabled        |
| **Total**   | **716** | **108 enabled** |

| Plugin     | Rules |
| ---------- | ----- |
| eslint     | 173   |
| unicorn    | 128   |
| typescript | 108   |
| react      | 57    |
| jest       | 56    |
| import     | 32    |
| jsx_a11y   | 31    |
| vitest     | 23    |
| nextjs     | 21    |
| jsdoc      | 18    |
| vue        | 17    |
| promise    | 16    |
| oxc        | 26    |
| node       | 6     |
| react_perf | 4     |

## Development

```bash
GOWORK=off GOEXPERIMENT=jsonv2 go test -race ./...  # Run tests with -race
GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...         # Run go vet
GOWORK=off GOEXPERIMENT=jsonv2 golangci-lint run ./...  # Run linter
nix build .                                          # Build via nix (runs tests)
nix run . -- configure                               # Run via nix (oxlint included)
oxlint -f json --rules > pkg/rule/rules_data.json  # Refresh rules from oxlint
GOWORK=off go mod vendor                             # Re-vendor deps (needed after go.mod changes)
```

## Architecture

```
oxlint-auto-configure/
├── cmd/oxlint-auto-configure/   # CLI entry point
├── pkg/
│   ├── rule/                   # Rule types, registry (716 embedded rules)
│   ├── profile/                # Profiles, categorization engine
│   ├── config/                 # .oxlintrc.json generator
│   ├── detect/                 # Project type detection
│   ├── diff/                   # Config comparison
│   ├── format/                 # Findings rendering (summary, JSON, table)
│   ├── oxlint/                 # go-finding Detector for oxlint
├── internal/cli/               # CLI commands (Cobra)
└── docs/                       # Documentation
```

## Tools Using This

- [go-finding](https://github.com/larsartmann/go-finding) — Unified static analysis data model
- [golangci-lint-auto-configure](https://github.com/larsartmann/golangci-lint-auto-configure) — Linter configuration for Go

## License

[MIT](LICENSE)
