#!/usr/bin/env bash
set -euo pipefail

SPEC="https://raw.githubusercontent.com/rixlhq/openapi/main/openapi.json"
OUT="provider_code_spec.json"

cd "$(dirname "$0")"

echo "Downloading OpenAPI spec..."
CURL_OPTS=(-sSL)
if [ -n "${GITHUB_TOKEN:-}" ]; then
  CURL_OPTS+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
fi
curl "${CURL_OPTS[@]}" -o openapi.json "$SPEC"

echo "Generating generator config and data source specs..."
python3 scripts/generate_config.py

echo "Generating provider code spec..."
tfplugingen-openapi generate --config generator_config.yml --output "$OUT" openapi.json

echo "Filtering generated data source specs for schemas skipped by the OpenAPI generator..."
python3 scripts/filter_missing_specs.py "$OUT" internal/provider/data_source_specs.go

echo "Generating Terraform plugin framework code..."
tfplugingen-framework generate all --input "$OUT" --output internal/provider --package provider

echo "Namespacing generated nested types to avoid cross-file collisions..."
python3 scripts/prefix_generated_types.py

echo "Renaming feed data source model to avoid conflict with feed resource model..."
sed -i 's/^type FeedModel struct {/type FeedDataSourceModel struct {/' internal/provider/feed_data_source_gen.go

echo "Formatting and building..."
gofmt -s -w internal/provider/*_gen.go internal/provider/data_source_specs.go
go mod tidy
go build ./...

echo "Done."
