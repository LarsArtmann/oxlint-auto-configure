# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

Run the following commands to set up your development environment:

    GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race
    GOWORK=off GOEXPERIMENT=jsonv2 golangci-lint run ./...

Or use Nix, which handles the experiment automatically:

    nix build .       # build and test
    nix flake check . # all checks

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
