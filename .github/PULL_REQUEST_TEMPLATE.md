<!-- Describe the problem this PR solves and why it matters. -->

## Summary

<!-- 1-3 bullets: what changes and the user-visible impact. -->

## Testing

<!-- How was this verified? Command + result, e.g. `GOWORK=off GOEXPERIMENT=jsonv2 go test -race ./...` -->

## Checklist

- [ ] `go test -race ./...` passes (with `GOWORK=off GOEXPERIMENT=jsonv2`)
- [ ] `go vet ./...` passes
- [ ] Config output still round-trips (`configure` → parse → `validate`)
