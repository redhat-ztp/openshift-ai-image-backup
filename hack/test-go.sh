#!/usr/bin/env bash
# single test: go test -v ./pkg/storage/
# without cache: go test -count=1 -v ./pkg/storage/
set -e -x
echo "Linting go code..."
make golangci-lint
make verify
echo "Running go tests..."
KUBEBUILDER_ASSETS="$(pwd)/bin" go test -v -covermode=count -coverprofile=coverage.out ./...
