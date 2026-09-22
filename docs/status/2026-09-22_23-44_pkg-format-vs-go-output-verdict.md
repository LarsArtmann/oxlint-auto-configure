# Status Report — pkg/format vs go-output verdict

**Date:** 2026-09-22 23:44 CEST
**Session scope:** One exploratory architecture question (`pkg/format/` vs `go-output/`?) — answered, no code changes requested or made. This report covers THIS session only, plus what I noticed while answering.
**Format note:** User explicitly requested `.md` (overrides the status-report skill's HTML default).

---

## What this session actually did

1. Read `pkg/format/format.go` in full (199 lines: `FindingView`, `SummaryView`, `PrintSummary`, `PrintFindingsJSON`, `PrintFindingsTable`, top-N helpers, `Map`).
2. Confirmed `go-output` exists as a local sibling repo; read its README (head) and repo layout.
3. Checked `go.mod` (go-output is **not** a dependency) and `internal/cli` usage of `format.*` (only `cmd_analyze.go` calls it).
4. Delivered verdict: **keep `pkg/format`; do not migrate to go-output** — outputs are API contracts, no format-matrix need, YAGNI. Migration escape hatch exists via the view-struct seam.
5. (Post-hoc, for this report) verified my own claims: test pinning, `format.Map` callers, go-output workspace claim.

---

## a) FULLY DONE

- **Comparison delivered** with a decision, rationale, and a concrete revisit trigger (runtime `--format` matrix or NOM-style progress → re-evaluate go-output).
- **No collateral damage**: zero file modifications, working tree untouched except this report.
- **Claims audit (this report)**: all three load-bearing claims from my answer now verified:
  - ✅ Output shapes ARE pinned by tests (`TestPrintFindingsJSON`, `TestPrintFindingsTable`, truncation + sort tests exist in `pkg/format/format_test.go`).
  - ❌→corrected: go-output sub-module friction was **overstated** (see section d).
  - 🎁 **New finding**: `format.Map` is a **ghost export** — zero production callers (only its own test calls it).

## b) PARTIALLY DONE

- **Decision hygiene**: verdict given verbally in chat, but NOT recorded anywhere durable (no AGENTS.md gotcha entry, no ROADMAP trigger condition, no ADR). A future session could re-litigate this from scratch.
- **Claim verification order**: I asserted "pinned by tests" and "workspace required" in the answer *before* verifying either. Both got verified afterwards — one held, one was wrong. Right conclusions, wrong epistemic order (should verify-then-claim, not claim-then-hope).

## c) NOT STARTED

- Any code change (none was requested — exploration mode only).
- Harvesting the verdict + findings into `TODO_LIST.md` / `ROADMAP.md` (status-report skill requires a HARVEST pass if session continues).
- `format.Map` ghost-export removal.
- The four code-smell fixes in `pkg/format` noticed this session (see e).

## d) TOTALLY FUCKED UP!

**Nothing destructive — but two honesty hits on my own answer:**

1. **I overstated the go-output workspace premise.** I wrote sub-modules "require cloning the repo and setting up the workspace." The README (line 222) actually says: external `go get` **resolves the published release** — cloning the workspace is merely the *intended consumption path*, not a hard requirement. My friction argument (#2 of 4) was stronger than reality. Verdict survives on arguments #1 (output contracts), #3 (no format-matrix need), #4 (seam already exists) — but argument #2 as stated was exaggerated. I took a README callout box at face value and didn't read the Installation section it pointed to. That's exactly the failure mode `verify-external-claims` exists for.
2. **Fabricated-before-verified test claim.** "Pinned by tests" was asserted without having opened `format_test.go`. It happened to be true. A claim being accidentally true is still a process failure.

**No code was harmed; the only damage is to the trustworthiness of a recommendation the owner might act on.**

## e) WHAT WE SHOULD IMPROVE!

**Process (mine):**
- Verify-then-claim, always. When citing a repo's constraints, read the section the summary links to, not the summary.
- When recommending keep-vs-adopt for a sibling library, check callers of every exported symbol I'm defending (this is how `format.Map`'s ghost status surfaced — only during report verification, not during the answer).

**Codebase smells noticed while reading `pkg/format/format.go` (not fixed, not previously tracked):**
- `PrintFindingsTable` truncates with `msg[:maxMessageLength-3]` — **byte slicing, not rune-safe**. A multibyte message can be split mid-UTF-8-sequence → invalid output bytes.
- Table cells escape `|` but **not newlines/control characters** — a message containing `\n` silently breaks the Markdown table row.
- Idiom inconsistency: `slices.Sorted(maps.Keys(m))` sits next to `sort.Slice`/`sort.Strings` in the same file (AGENTS mandates the `slices` idiom for Go 1.23+).
- `PrintSummary` signature returns `error` but **always returns nil** — `fmt.Fprintf` errors are discarded. Either propagate or document why not.
- `format.Map` — tested, exported, **never called** in production code (ghost export; split-brain risk with `pkg/diff`'s `formatValue` which does similar map→string formatting).
- `FindingView.DocsURL` and `Snippet` are serialized to JSON but never surfaced in any human-readable format (table/summary ignore them) — free value left on the table.
- Possible third formatting split-brain to audit: `internal/cli/cmd_report.go` has its own format helpers (per AGENTS) alongside `pkg/format` — overlap unexamined.

## f) Things to get done next (prioritized; ≤50, this list: 28)

*Session-derived items first; final group is carried from loaded AGENTS context, not re-verified today.*

| # | Task | Impact / Effort |
|---|------|-----------------|
| 1 | Delete `format.Map` + `TestFormatMap` (ghost export) or wire it | Med / XS |
| 2 | Record the pkg/format-vs-go-output decision in AGENTS.md gotchas incl. revisit trigger | Med / XS |
| 3 | Fix rune-safe truncation in `PrintFindingsTable` (`[]rune` or clip at rune boundary) | Med / XS |
| 4 | Escape newlines/control chars in Markdown table cells | Med / XS |
| 5 | Add multibyte-truncation + newline-in-message regression tests | Med / XS |
| 6 | Migrate `sort.Slice`/`sort.Strings` in format.go to `slices.SortFunc`/`slices.Sorted` | Low / XS |
| 7 | Audit split-brain: `cmd_report.go` format helpers vs `pkg/format` vs `pkg/diff.formatValue` — one formatting seam or documented boundaries | Med / S |
| 8 | HARVEST this report into TODO_LIST.md/ROADMAP.md per docs-health | Med / XS |
| 9 | Decide `PrintSummary` error contract: propagate `fmt.Fprintf` errors or drop the return | Low / XS |
| 10 | Surface `DocsURL` in table output (e.g., `--docs` flag) or document why JSON-only | Low / S |
| 11 | ROADMAP: optional CSV/HTML output for `analyze` — the trigger that would justify go-output adoption | Low / M |
| 12 | ROADMAP: NOM-style progress visualization for pipeline iterations (go-output feature) — evaluate if analyze iterations ever feel opaque | Low / M |
| 13 | If/when go-output adoption spikes: verify root-module-only rendering coverage (README claims root needs only `go get`) | Low / M |
| 14 | Package doc polish: add a godoc example for `pkg/format` | Low / XS |
| 15 | Correct my earlier session statement in any doc that repeated the "cloning required" claim (none known to exist — verify) | Low / XS |
| 16 | Add `internal/cli` grep-test or convention note pinning that `pkg/format` is the only findings-rendering path (guards the seam) | Low / S |
| 17 | Consider a `TestNoGhostExports` style check: exported symbols in `pkg/format` must have a production caller (generalizable) | Low / M |
| 18 | Check whether `maxMessageLength-3` handles messages of exactly `maxMessageLength`-2/-1 bytes (edge tests) | Low / XS |
| 19 | `FindingView.Snippet`: render as fenced code line in table or drop from struct if permanently unused visually | Low / S |
| 20 | Document the output-contract policy (JSON field names + table columns are public CLI API) in AGENTS.md | Med / XS |
| 21 | Track go-output version compatibility (Go 1.26+ floor there, 1.27.1 here — currently fine; recheck on adoption) | Low / XS |
| 22 | If deleting `format.Map` instead of wiring: also prune `formatMap` (single caller gone) | Low / XS |
| 23 | Sweep other `msg[:n]` byte-slicing patterns in repo for the same rune bug | Med / S |
| 24 | Consider `fmt.Fprintf` → single `strings.Builder` in `PrintSummary` for one write syscall | Low / XS |
| 25 | Add test that `PrintFindingsTable` output round-trips as valid Markdown (lint via renderer if ever adopted) | Low / S |
| 26 | *(carried context)* Watch golangci-lint v2.13.2 pin vs `.golangci.yml` generation on next bump (known coupling) | Low / XS |
| 27 | *(carried context)* `goreleaser check` deprecation warnings (`dockers_v2` etc.) — recheck at next GoReleaser release | Low / XS |
| 28 | *(carried context)* Confirm `workflow-health.yml` canary is active post-billing-fix (asserted done 2026-09-22 earlier session; cheap to re-verify) | Low / XS |

## g) Questions I cannot figure out myself (max 3)

1. **Record the decision now?** Should the "keep `pkg/format`, adopt go-output only on a format-matrix need" verdict go into AGENTS.md as a binding decision (item 2), or stay chat-ephemeral until the need actually materializes?
2. **Are analyze's JSON/table outputs consumed by anything external** (scripts, CI annotation parsers, dashboards) whose parsing would break if we fix truncation/escaping or touch column layout? I can see the repo, not your environment.
3. **House standard mandate?** If/when a multi-format need lands in ANY LarsArtmann CLI, do you want go-output mandated as the default rendering layer (one library everywhere), or chosen per-tool on merit?

---

*Auto-commit daemon will pick up this file. Harness forbids manual commits without explicit request.*
