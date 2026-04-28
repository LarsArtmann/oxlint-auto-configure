# oxlint-auto-configure

Automatically configure [oxlint](https://oxc.rs/docs/guide/usage/linter.html) for maximum type safety. Uses [go-finding](https://github.com/larsartmann/go-finding) to run oxlint, collect findings, and auto-configure every available rule with the best severity setting.

## Why

Oxlint has **716 rules** across **7 categories** and **15 plugins**. Only 108 are enabled by default. Manually configuring each rule for maximum type safety is tedious and error-prone. This tool automates the entire process:

- Discovers your project type (React, Next.js, Vue, etc.)
- Enables relevant plugins automatically
- Sets optimal severity for every rule based on your chosen profile
- Generates a ready-to-use `.oxlintrc.json`

## Installation

```bash
go install github.com/larsartmann/oxlint-auto-configure/cmd/oxlint-auto-configure@latest
```

Requires Go 1.26+ and [oxlint](https://oxc.rs/docs/guide/usage/linter.html) in PATH.

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

| Profile | Correctness | Suspicious | TypeScript | Style | Perf | Pedantic | Restriction | Nursery |
|---------|------------|------------|------------|-------|------|----------|-------------|---------|
| **maximal-typesafe** | error | error | error | error | error | error | error | warn |
| **strict** | error | error | error | warn | warn | warn | warn | off |
| **recommended** | error | error | error | warn | warn | warn | warn | off |
| **minimal** | error | warn | warn | warn | warn | warn | off | off |

### Profile Details

- **maximal-typesafe**: Every single rule at `error`. Maximum type safety and correctness enforcement. Even nursery rules at `warn`.
- **strict**: Core correctness at `error`, everything else at `warn` except nursery.
- **recommended** (default): Correctness + suspicious + TypeScript rules at `error`. Style, perf, pedantic at `warn`. Restriction at `warn`. Nursery off.
- **minimal**: Only correctness at `error`. Everything else uses oxlint defaults.

## Project Detection

The tool auto-detects your project type and enables relevant plugins:

| Detected | Plugins Enabled |
|----------|----------------|
| React | `react`, `jsx-a11y`, `react-perf` |
| Next.js | `nextjs`, `react`, `jsx-a11y` |
| Vue | `vue` |
| Jest | `jest`, `node` |
| Vitest | `vitest` |
| TypeScript | `typescript` (always on) |

## Commands

### `configure`

Generate the optimal `.oxlintrc.json`:

```bash
oxlint-auto-configure configure [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-p, --profile` | `recommended` | Configuration profile |
| `-c, --config` | `.oxlintrc.json` | Output config file path |
| `-d, --dry-run` | false | Show changes without writing |
| `--fix` | false | Run oxlint --fix after writing config |
| `--root` | `.` | Project root directory |

### `analyze`

Run oxlint and show findings using the go-finding pipeline:

```bash
oxlint-auto-configure analyze [--root .] [-f summary|json|sarif|table]
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

| Category | Count | Default |
|----------|-------|---------|
| Correctness | 216 | 108 enabled |
| Style | 211 | mostly disabled |
| Pedantic | 115 | disabled |
| Restriction | 91 | disabled |
| Suspicious | 52 | disabled |
| Perf | 13 | disabled |
| Nursery | 18 | disabled |
| **Total** | **716** | **108 enabled** |

| Plugin | Rules |
|--------|-------|
| eslint | 173 |
| unicorn | 128 |
| typescript | 108 |
| react | 57 |
| jest | 56 |
| import | 32 |
| jsx_a11y | 31 |
| vitest | 23 |
| nextjs | 21 |
| jsdoc | 18 |
| vue | 17 |
| promise | 16 |
| oxc | 26 |
| node | 6 |
| react_perf | 4 |

## Development

```bash
just build       # Build CLI binary
just test        # Run tests with -race
just lint        # Run golangci-lint
just cover       # Coverage report
just check       # All checks
just update-rules # Refresh rules from oxlint
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
│   ├── oxlint/                 # go-finding Detector for oxlint
├── internal/cli/               # CLI commands (Cobra)
└── docs/                       # Documentation
```

## Tools Using This

- [go-finding](https://github.com/larsartmann/go-finding) — Unified static analysis data model
- [golangci-lint-auto-configure](https://github.com/larsartmann/golangci-lint-auto-configure) — Linter configuration for Go

## License

[MIT](LICENSE)
