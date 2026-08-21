#!/usr/bin/env bash
set -euo pipefail

PATH="$(go env GOPATH)/bin:${PATH}"

cd "$(dirname "$0")"

echo "Formatting and building..."
gofmt -s -w internal/provider/*.go
go mod tidy
go build ./...

echo "Done."
