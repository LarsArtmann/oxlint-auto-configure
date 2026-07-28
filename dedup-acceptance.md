# Dedup Acceptance Log

Clone groups that art-dupl reports but are **intentionally accepted** — verified,
not "good enough." Each entry cites the constraint that makes the duplication
unavoidable or idiomatic.

---

## Clone Group: `t.Parallel()` + `reg := loadTestRegistry(t)` — 26 occurrences

**Locations:** `pkg/rule/registry_test.go`, `pkg/config/configure_test.go`,
`pkg/config/generator_test.go`, `pkg/config/validate_test.go`,
`pkg/profile/profile_test.go`

**Status:** Accepted — linter-enforced, cannot be eliminated.

**Reason:** The `paralleltest` linter (enabled in `.golangci.yml`) requires
`t.Parallel()` to appear as a direct statement in every `Test*` function body.
Moving it into the `loadTestRegistry` helper causes:

```
Function TestX missing the call to method parallel (paralleltest)
```

Verified empirically: moving `t.Parallel()` into `loadTestRegistry` produces 2
`paralleltest` failures per test file (one per test function). The duplication is
therefore structural — 2 lines × 26 sites — enforced by CI, not negligence.

**Do not attempt to refactor.** See `AGENTS.md` → "Test boilerplate is intentional".
