#!/usr/bin/env bash
set -euo pipefail

PATH="$(go env GOPATH)/bin:${PATH}"

cd "$(dirname "$0")"

echo "Generating provider code spec from OpenAPI..."
tfplugingen-openapi generate openapi.yaml --config generator_config.yml --output provider_code_spec.json

# Remove the sensitive 'secret' attribute from the API keys data source so it is not
# exposed through a Terraform data source.
jq 'walk(if type == "array" then map(select(.name? != "secret")) else . end)' provider_code_spec.json > provider_code_spec.json.tmp
mv provider_code_spec.json.tmp provider_code_spec.json

echo "Generating data source code..."
tfplugingen-framework generate data-sources --input provider_code_spec.json --output internal/provider --package provider

echo "Renaming generated data source types..."
go run ./tools/dsrename

echo "Formatting and building..."
gofmt -s -w internal/provider/*.go
go mod tidy
go build ./...

echo "Done."
