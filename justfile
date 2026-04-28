# oxlint-auto-configure

build:
    go build -o bin/oxlint-auto-configure ./cmd/oxlint-auto-configure/

install: build
    cp bin/oxlint-auto-configure $(GOPATH)/bin/oxlint-auto-configure

test:
    go test -race ./pkg/... -count=1

bench:
    go test -bench=. -benchmem ./pkg/...

cover:
    go test -race -coverprofile=coverage.out ./pkg/...
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

check: fmt-check lint test
    @echo "All checks passed."

clean:
    rm -rf bin/ coverage.out coverage.html

run: build
    ./bin/oxlint-auto-configure {{argv}}

# Update rules from oxlint
update-rules:
    oxlint -f json --rules > pkg/rule/rules_data.json
