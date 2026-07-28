# Dedup Acceptance Log

Clone groups that art-dupl reports but are **intentionally accepted** — verified,
not "good enough." Each entry cites the constraint that makes the duplication
unavoidable or idiomatic.

---

## Clone Group: `t.Parallel()` + `reg := testregistry.Load(t)` — 16 occurrences

**Locations:** `pkg/config/configure_test.go`, `pkg/config/generator_test.go`,
`pkg/config/validate_test.go`, `pkg/profile/profile_test.go`

**Status:** Accepted — linter-enforced, cannot be eliminated.

**Reason:** The `paralleltest` linter (enabled in `.golangci.yml`) requires
`t.Parallel()` to appear as a direct statement in every `Test*` function body.
Moving it into the `testregistry.Load` helper causes:

```
Function TestX missing the call to method parallel (paralleltest)
```

Verified empirically: moving `t.Parallel()` into the helper produces
`paralleltest` failures on every test function. The duplication is therefore
structural — 2 lines × 16 sites — enforced by CI, not negligence.

The `testregistry.Load(t)` call itself was **extracted** from 3 byte-for-byte
identical per-package `loadTestRegistry` helpers into a single
`internal/testregistry.Load`. The `pkg/rule/registry_test.go` copy remains local
because that test file accesses the unexported `mapFix` function (package
`rule`) and cannot import `internal/testregistry` without an import cycle
(`internal/testregistry` → `pkg/rule` → `internal/testregistry`).

**Do not attempt to refactor the `t.Parallel()` line.**
See `AGENTS.md` → "Test boilerplate is intentional".
