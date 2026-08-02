#!/usr/bin/env bash
set -euo pipefail

SPEC="https://raw.githubusercontent.com/rixlhq/openapi/main/openapi.yaml"
OUT="provider_code_spec.json"

cd "$(dirname "$0")"

echo "Downloading OpenAPI spec..."
curl -sSL -o openapi.yaml "$SPEC"

echo "Generating provider code spec..."
tfplugingen-openapi generate --config generator_config.yml --output "$OUT" openapi.yaml

echo "Generating Terraform plugin framework code..."
tfplugingen-framework generate all --input "$OUT" --output internal/provider --package provider

# Rename the data source model to avoid a conflict with the feed resource model.
sed -i 's/^type FeedModel struct {/type FeedDataSourceModel struct {/' internal/provider/feed_data_source_gen.go

echo "Formatting and building..."
gofmt -s -w internal/provider/*_gen.go
go mod tidy
go build ./...

echo "Done."
