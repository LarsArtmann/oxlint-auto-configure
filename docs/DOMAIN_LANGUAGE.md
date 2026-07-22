# Domain Language

A **Unified Language** for **oxlint-auto-configure** — shared across Customer, Product Owner, Developer, and AI.

Every term below should mean the **same thing** to everyone who reads it.
If a word means something different to a developer than to a customer, define it here.

## Glossary

| Term                      | Definition                                                                         | Context       |
| ------------------------- | ---------------------------------------------------------------------------------- | ------------- |
| **oxlint**                | The fast linter this tool configures.                                              | External tool |
| **oxlint-auto-configure** | This Go CLI. Generates `.oxlintrc.json` files.                                     | Product name  |
| **Rule**                  | A single lint rule (e.g., `no-unused-vars`).                                       | Core domain   |
| **Plugin**                | A named set of rules from a source (e.g., `react`, `typescript`).                  | Core domain   |
| **Category**              | A rule classification (e.g., `correctness`, `style`, `nursery`).                   | Core domain   |
| **Profile**               | A preset severity policy (`maximal-typesafe`, `recommended`, `strict`, `minimal`). | Core domain   |
| **Severity**              | How a rule is enforced (`error`, `warn`, `off`).                                   | Core domain   |
| **Project Type**          | Detected framework/ecosystem (e.g., `react`, `nextjs`, `vue`, `jest`).             | Detection     |
| **PluginConfig**          | Map of plugins to enable for a detected project.                                   | Detection     |
| **.oxlintrc.json**        | The generated oxlint configuration file.                                           | Output        |
| **go-finding**            | The static-analysis pipeline library used by `analyze`.                            | Dependency    |

## Entities

Objects with identity and lifecycle.

| Term              | Definition                                                             | Context                   |
| ----------------- | ---------------------------------------------------------------------- | ------------------------- |
| **Rule Registry** | The embedded collection of 716 oxlint rules loaded at startup.         | `pkg/rule/registry.go`    |
| **Categorizer**   | The engine that maps categories and rules to severities for a profile. | `pkg/profile/profile.go`  |
| **Generator**     | Creates the `.oxlintrc.json` structure from profile decisions.         | `pkg/config/generator.go` |
| **Detector**      | Discovers project type from `package.json` and filesystem.             | `pkg/detect/detector.go`  |
| **Report**        | A go-finding `Report` produced by the `analyze` command.               | `pkg/oxlint/detector.go`  |

## Value Objects

Immutable objects defined by attributes.

| Term                 | Definition                                                 | Context                  |
| -------------------- | ---------------------------------------------------------- | ------------------------ |
| **Profile Name**     | One of the four supported profile strings.                 | CLI flag, config         |
| **SeverityDecision** | The resolved severity for a category or rule.              | `pkg/rule/rule.go`       |
| **FixCapability**    | Whether a rule is fixable (`none`, `manual`, `automatic`). | `pkg/rule/rule.go`       |
| **ProjectType**      | A detected project type string constant.                   | `pkg/detect/detector.go` |

## Events

Things that happen in the domain.

| Term                  | Definition                                                     | Context                 |
| --------------------- | -------------------------------------------------------------- | ----------------------- |
| **Config Generated**  | The `.oxlintrc.json` file has been written or printed.         | `configure` command     |
| **Project Detected**  | The detector has identified project types from `package.json`. | `configure` / `analyze` |
| **Findings Reported** | The `analyze` command has emitted findings.                    | `analyze` command       |
| **Config Validated**  | The `validate` command has confirmed the config is parseable.  | `validate` command      |

## Commands

Actions the system can perform.

| Term          | Definition                                                       | Context     |
| ------------- | ---------------------------------------------------------------- | ----------- |
| **Configure** | Generate an `.oxlintrc.json` for the detected project.           | CLI command |
| **Analyze**   | Run oxlint through the go-finding pipeline and report findings.  | CLI command |
| **Validate**  | Check an existing `.oxlintrc.json` for unknown rules/severities. | CLI command |
| **Report**    | List all rules and their recommended severities for a profile.   | CLI command |

## Bounded Contexts

No overlapping vocabulary currently requires separate contexts; all terms above are used consistently across the CLI, docs, and code.

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
