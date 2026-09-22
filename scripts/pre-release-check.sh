#!/usr/bin/env bash
# Pre-release gate: everything that must pass locally before tagging a release.
#
# Encodes the manual validation from the v0.5.0 release session:
#   1. module consistency  (go mod tidy leaves go.mod/go.sum unchanged)
#   2. build + vet + tests (the same gate CI runs)
#   3. golangci-lint       (when available)
#   4. goreleaser check    (config deprecations/errors)
#   5. goreleaser snapshot (a full local release dry-run)
#
# Usage: scripts/pre-release-check.sh
set -euo pipefail

cd "$(dirname "$0")/.."

export GOWORK=off
export GOEXPERIMENT=jsonv2
export GOTOOLCHAIN=auto

step() { printf '\n==> %s\n' "$1"; }

step "go mod tidy (module consistency)"
tmp="$(mktemp -d)"
cp go.mod go.sum "$tmp/"
go mod tidy
if ! cmp -s go.mod "$tmp/go.mod" || ! cmp -s go.sum "$tmp/go.sum"; then
	diff -u "$tmp/go.mod" go.mod || true
	diff -u "$tmp/go.sum" go.sum || true
	printf 'ERROR: go mod tidy changed go.mod/go.sum — commit the result first\n' >&2
	exit 1
fi

step "go build"
go build -v ./...

step "go vet"
go vet ./...

step "go test"
go test -race ./...

if command -v golangci-lint >/dev/null 2>&1; then
	step "golangci-lint"
	golangci-lint run --timeout=5m
else
	printf 'golangci-lint not found, skipping\n'
fi

if ! command -v goreleaser >/dev/null 2>&1; then
	printf '\nERROR: goreleaser not found — install it to validate the release pipeline\n' >&2
	exit 1
fi

step "goreleaser check"
goreleaser check

step "goreleaser release --snapshot"
# docker, sbom, sign, and validate are skipped: they need docker/cosign/syft
# and only run in CI; a local snapshot exists to catch template and archive
# errors (e.g. brew/nix pipe masking) before a tag does.
GOEXPERIMENT=jsonv2 goreleaser release --snapshot --clean --skip=docker,sbom,sign,validate

printf '\nAll pre-release checks passed.\n'
