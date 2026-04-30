# oxlint-auto-configure

build:
    go build -o bin/oxlint-auto-configure ./cmd/oxlint-auto-configure/

install: build
    cp bin/oxlint-auto-configure $(GOPATH)/bin/oxlint-auto-configure

test:
    go test -race ./pkg/... ./internal/... -count=1

bench:
    go test -bench=. -benchmem ./pkg/... ./internal/...

cover:
    go test -race -coverprofile=coverage.out ./pkg/... ./internal/...
    go tool cover -func=coverage.out

cover-html: cover
    go tool cover -html=coverage.out -o coverage.html

lint:
    golangci-lint run ./...

fmt:
    gofmt -w -s .

fmt-check:
    gofmt -l -s .

tidy:
    go mod tidy

check: fmt-check vet lint test
    @echo "All checks passed."

vet:
    go vet ./...

clean:
    rm -rf bin/ coverage.out coverage.html

run: build
    ./bin/oxlint-auto-configure {{argv}}

# Update rules from oxlint
update-rules:
    oxlint -f json --rules > pkg/rule/rules_data.json
    oxlint --version | sed 's/Version: //' > pkg/rule/rules_version.txt

# Vendor Go dependencies (required for nix build)
vendor:
    GOWORK=off go mod vendor

# Nix build
nix-build:
    nix build .

# Enter nix dev shell
nix-shell:
    nix develop .
