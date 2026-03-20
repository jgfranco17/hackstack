# Project scripts

# Default recipe
_default:
    @just --list --unsorted

# Run the service
start:
    @go run ./cmd/api

# Execute tests
test:
    #!/usr/bin/env bash
    echo "Running tests..."
    go clean -testcache
    go test -cover ./...
