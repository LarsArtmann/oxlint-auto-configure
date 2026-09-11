# Status Report: BuildFlow Integration via toolsdk (pkg/provider)

**Date:** 2026-09-11 14:18 CEST
**Scope:** Single session in `oxlint-auto-configure` (+ surgical BuildFlow wiring) — task: "I just want to better integrate with /home/lars/projects/BuildFlow"
**Verification state at writing:** `go test -race ./...` green (all packages incl. new `pkg/provider`), `go vet` clean, `golangci-lint run` 0 issues, `nix build` + `nix flake check` all checks passed, binary runs (`--version` OK), BuildFlow `tools/providers` full suite green (26.4s). All changes committed by the auto-commit daemon; both working trees clean.

---

## Session Summary

Adopted `github.com/larsartmann/go-finding/toolsdk v1.10.0` and shipped `pkg/provider`: a self-registering `toolsdk.Spec` (Detect = missing `.oxlintrc.json` finding `OXLOPT_CONFIG_MISSING`; Repair = generate recommended-profile config, dry-run aware, never overwrites). BuildFlow's 132-line hand-written `NewOxlintAutoConfigureProvider` glue was deleted and replaced by a blank import (`sdk_imports.go`), following the established provider pattern (dependabot-auto-configure, go-structure-linter, licenseforge). Also fixed a stale flake pin (`go-finding` v1.8.0 → v1.10.0, required by toolsdk) and refreshed `vendorHash`.

## a) FULLY DONE

| # | Item | Evidence |
| --- | --- | --- |
| 1 | `pkg/provider` package: Spec (name `oxlint-auto-configure`, JS/TS trigger, `Inputs: [package.json, .oxlintrc.json]`, `DependsOn: [oxlint]`, nil HealthCheck) registered at import time | `pkg/provider/provider.go` |
| 2 | Detect: warning finding when a recognizable JS/TS project lacks `.oxlintrc.json`; existing configs and non-JS dirs never flagged | `detectMissingConfig`; guarded by `hasConfig` + `hasKnownProjectType` |
| 3 | Repair: detect → `ProfileRecommended` → generate → atomic write; honors `toolsdk.DryRunFromContext`; refuses to overwrite existing config (glue overwrote unconditionally and ignored dry-run — both glue bugs fixed by the migration) | `repairConfig`, `generateConfig`, `writeConfigFile` |
| 4 | 8 contract tests: registration uniqueness, detect semantics (react/known type, existing config, bare package.json = Node, no markers), dry-run hold-back, round-trip via `config.FromJSON`, no-overwrite | `pkg/provider/provider_test.go`, all green under `-race` |
| 5 | Dependency + vendor: `go-finding/toolsdk v1.10.0` in go.mod; `vendor/` regenerated | `go.mod:9` |
| 6 | Stale flake pin fixed: `go-finding` input v1.8.0 → v1.10.0 (go.mod required v1.10.0 since an earlier session; additive API had kept the stale pin compiling); vendorHash refreshed via the documented mismatch workflow (fail → copy `got:` → pass) | `flake.nix:28`, `flake.nix:54`; `nix build` + `nix flake check` green |
| 7 | BuildFlow glue deleted: `NewOxlintAutoConfigureProvider` + `oxlintAutoConfigureSDKDetector` + `oxlintAutoConfigureSDKRepairer` (-132 lines), registration line removed | BuildFlow `tools/providers/sdk_external_tools.go`, `tools/providers/registration.go` |
| 8 | BuildFlow blank import added with capability comment; `ToolsFromSDK()` now surfaces the tool | BuildFlow `tools/providers/sdk_imports.go` |
| 9 | BuildFlow tests updated per their own guards: glue test rewritten as `TestToolsFromSDK_OxlintAutoConfigure`; `TestOxlintAutoConfigureProviderRegistered` added (the meta-guard `TestAllSDKToolsHaveProviderRegisteredTest` demanded it); DAG data-flow snapshot updated 215 → 213 with rationale per the test's documented protocol | BuildFlow `go_auto_upgrade_helpers_test.go`, `sdk_imports_test.go`, `dataflow_edge_snapshot_test.go` |
| 10 | Unblocked two PRE-EXISTING BuildFlow breakages (not mine): `buildUpgradeProjectInfo` test call sites missing the committed `root string` arg; `go.sum` entries for licenseforge's new `samber/oops` import | BuildFlow `sdk_external_tools_test.go`, `tools/go.sum` (via `go mod tidy`) |
| 11 | Living docs: CHANGELOG `[Unreleased]` (Added + Changed), FEATURES row (BuildFlow integration → FULLY_FUNCTIONAL), AGENTS (Key Files, Key Test Files, Dependencies, BuildFlow-integration + stale-flake-pin gotchas), TODO_LIST (BuildFlow Integration section) | 4 files in this repo |
| 12 | End-to-end proof: nix binary reports correct version metadata; BuildFlow suite discovers the tool through the blank import | `result/bin/oxlint-auto-configure --version`; BuildFlow `go test ./providers/` ok |

## b) PARTIALLY DONE

| # | Item | Works now | Missing | Blocker | Effort |
| --- | --- | --- | --- | --- | --- |
| 1 | README visibility | FEATURES/CHANGELOG/AGENTS updated | README (the sales page) does not mention the BuildFlow integration | None — pure follow-up | S |
| 2 | Formatter validation | Go files: golangci formatters (gci/goimports/gofumpt/golines) all pass | dprint (canonical for MD/JSON) unavailable in `$PATH` AND in `nix develop` shell this session — the 4 edited .md files were never dprint-validated; table pipe alignment likely off | dprint not installed/wired in dev shell | S |
| 3 | BuildFlow CHANGELOG | Code + tests updated there | No entry in BuildFlow's CHANGELOG.md for the glue removal + edge-count change; I did not research their changelog conventions | None | S |
| 4 | `DependsOn: ["oxlint"]` ordering | Glue parity preserved exactly | The ordering is questionable: config generation AFTER oxlint lints means the generated config only benefits the NEXT run. I flagged it mentally, chose parity, documented nothing | Owner decision (DAG topology is BuildFlow's design) | S |
| 5 | toolsdk `Outputs` gap | Documented in the snapshot-test comment + TODO_LIST | The Spec contract has no Outputs field, so `**/.oxlintrc.json` producer edges are lost (part of the 215→213 delta); same limitation affects dependabot. Nothing filed upstream | Upstream repo decision | M |
| 6 | LSP hygiene | Refuted all stale gopls errors via fresh CLI builds (per AGENTS lesson) | Never ran `lsp_restart` after the go.mod change; the false "toolsdk not in go.mod" errors polluted every tool result all session | None | S |
| 7 | BuildFlow root module | tools module (where all wiring lives) fully verified | Root module never built/tested this session (root go.mod lists this repo as `indirect`; almost certainly unaffected — but unverified claim) | None | S |

## c) NOT STARTED

- **Tag + BuildFlow flake bump (the actual release path):** tag next version (e.g. `v0.6.0`), bump BuildFlow's `flake.nix` input (`refs/tags/v0.5.0` → new tag) + their vendorHash, drop reliance on the local replace for nix CI. Queued in TODO_LIST.
- **Config-drift detection:** Detect currently only flags a MISSING config. Reusing `pkg/diff` to flag stale/divergent configs was considered and consciously deferred (risk: stomping user customizations; needs an owner decision).
- **Upstream toolsdk contribution:** Outputs field / trigger-Files gap — not started.
- **Consumer-side Register example in a README/docs** (mirroring linter-autoconfigure-sdk's backlog item #25) — not started.

## d) TOTALLY FUCKED UP

Nothing catastrophic: no failing final state, no data loss, no broken builds left behind. Radical-honesty list:

1. **Didn't read BuildFlow's AGENTS "Adding a tool" steps before wiring.** Their documented step 4 requires a `*ProviderRegistered` test, and a meta-guard enforces it. I only added the test after the guard failed. Reading the host repo's docs first is the rule; I applied it to code but not to their process docs.
2. **"Improved" a test pattern into a failure.** I used the constant `config.ToolOxlintAutoConfigure` where all three existing sibling tests use the literal `config.ToolName("oxlint-auto-configure")` — the meta-guard greps the literal. Second avoidable red run. When mirroring a pattern, mirror it exactly.
3. **Wrote tests from assumption instead of reading the domain first.** `TestDetect_UnknownProjectType` failed because I never read `inferFallbackTypes` — a bare package.json IS a Node project by this domain's definition. The mandatory pre-write reflection would have caught it.
4. **Predictable `go get`-before-import mistake.** Ran `go get toolsdk` first; `go mod tidy` (no importer yet) dropped the require. I know MVS behavior; our sibling repo's status report documents this exact mistake class. One wasted round trip.
5. **Built BuildFlow without `GOWORK=off` first.** The workspace resolution bypassed `tools/go.mod`'s local replace and produced phantom "undefined API" errors against proxy v0.5.0 of go-auto-upgrade. Our own AGENTS documents the GOWORK=off reflex for this exact parent-workspace situation. One wasted build cycle and a false "pre-existing breakage" scare.
6. **Self-inflicted edit conflicts.** `golangci-lint run --fix` reformatted `provider.go` between my read and edit → two "file modified since read" failures.
7. **Formatter gap reported late and partly silently.** dprint was unavailable; I validated Go formatting but let the .md validation gap slip out of the final summary until now.
8. **Commit history quality (systemic, not mine alone):** everything landed in heuristic daemon messages ("chore: auto-commit N file(s)") — the same burying problem linter-autoconfigure-sdk's status report complained about. I don't commit unbidden, but I also didn't propose a mitigation.

## e) WHAT WE SHOULD IMPROVE

1. **Host-repo process docs are part of the contract.** Before wiring into a foreign repo: read its AGENTS' "adding X" checklist FIRST (BuildFlow's step 4 would have saved a red run).
2. **Mirror patterns verbatim; "improve" only deliberately.** Sibling tests encode machine-checked conventions (the literal-string meta-guard). Divergence = failure, not improvement.
3. **Read the domain logic before writing its tests.** Tests encode the model; writing them from assumptions tests the assumption.
4. **Order: imports → go get → tidy.** Never `go get` without a referencing import when tidy will run.
5. **GOWORK=off is the default posture** for every nested-module build in this ecosystem, not a recovery step.
6. **Restart the LSP after go.mod changes** (or configure gopls env) instead of enduring stale errors all session.
7. **Report tooling gaps explicitly** (dprint unavailable) instead of letting them vanish between verification lines.
8. **Kill the glue at both ends in one motion:** blank import + registration-line removal + glue deletion were done sequentially; a single coordinated change (as done) should be the plan from the start, and the DAG snapshot test should be anticipated as the arbiter of contract changes.

## f) Top things to get done next (ranked by impact)

| # | Task | Impact | Effort | Category |
| --- | --- | --- | --- | --- |
| 1 | Tag `v0.6.0`; bump BuildFlow flake input + vendorHash so nix CI builds the provider | High | S | Release |
| 2 | Decide `DependsOn` ordering: generate config BEFORE oxlint lints (flip to `oxlint` depending on us) vs glue parity | High | S | Decision |
| 3 | Decide config-drift detection scope (missing-only vs stale-config findings via `pkg/diff`) | High | S | Decision |
| 4 | Update README with the BuildFlow integration (sales page duty) | Medium | S | Documentation |
| 5 | Add BuildFlow CHANGELOG entry for glue removal + snapshot change | Medium | S | Documentation |
| 6 | Wire dprint into the dev shell (or install) and validate the 4 edited .md files | Medium | S | Tooling |
| 7 | Upstream toolsdk: propose `Outputs []string` on Spec (restores producer edges for us + dependabot) | Medium | M | Feature |
| 8 | BuildFlow root-module build/test verification (close the unverified claim) | Medium | S | Quality |
| 9 | Configure gopls with `GOEXPERIMENT=jsonv2` env / restart-after-go.mod reflex to kill stdversion + stale-module noise | Medium | S | Tooling |
| 10 | Add provider-side dry-run godoc note (BuildFlow's `--dry-run` reaches Repair via ctx) — parity with linter-autoconfigure-sdk backlog #5 | Medium | S | Documentation |
| 11 | Test: Repair forward ctx so future closures can read `DryRunFromContext` (assert flag reaches the pipeline) — currently only my own closure reads it | Low | S | Quality |
| 12 | `-count=2` / shuffle run of `pkg/provider` tests (registry is process-global; guard against order dependence) | Low | S | Quality |
| 13 | Consider `toolsdk.Register` duplicate-name panic policy feedback upstream (today duplicates append silently) | Low | S | Decision |
| 14 | Check whether BuildFlow's `--list` output renders our Description acceptably (inventory UI) | Low | S | Quality |
| 15 | Provider Inputs: consider adding root `pnpm-lock.yaml`/`package-lock.json` detection inputs if BuildFlow file-gating ever gates on Inputs presence | Low | S | Decision |

(15 items; the remaining headroom of the 50 is deliberately left to the decisions above — items 1–3 reshape everything downstream, so I won't pad the list with speculative work.)

## g) Questions I cannot figure out myself

1. **DAG order:** should the oxlint config be generated BEFORE the oxlint lint step in BuildFlow's DAG (i.e. flip `DependsOn` so oxlint depends on us and benefits from the config in the same run), or is after-the-lint intentional?
2. **Release:** tag `v0.6.0` now so BuildFlow's flake/nix CI picks up the provider, or hold until drift detection (if approved) ships in the same tag?
3. **Drift contract:** should Detect also flag stale/divergent configs (reusing `pkg/diff`), accepting that Repair may then overwrite user customizations — or is missing-config-only the permanent contract?

---

_Assisted-by: Crush <crush@charm.land>_
