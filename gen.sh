#!/usr/bin/env bash
set -euo pipefail

PATH="$(go env GOPATH)/bin:${PATH}"

cd "$(dirname "$0")"

# The OpenAPI spec is gitignored. When it is not present locally (e.g. in CI),
# download it from the rixlhq/openapi repository.
if [ ! -f openapi.yaml ]; then
  echo "Downloading OpenAPI spec..."
  SPEC="https://raw.githubusercontent.com/rixlhq/openapi/main/openapi.yaml"
  CURL_OPTS=(-sSL)
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    CURL_OPTS+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
  fi
  curl "${CURL_OPTS[@]}" -o openapi.yaml "$SPEC"
fi

echo "Generating provider code spec from OpenAPI..."
tfplugingen-openapi generate \
  --config generator_config.yml \
  --output provider_code_spec.json \
  openapi.yaml

# Add list-based resources (api_key, client_credential), fix resource schemas,
# ensure parent/id parameters are required with RequiresReplace, and remove the
# sensitive 'secret' attribute.
python3 tools/postprocess_spec.py

# Remove stale generated files; tfplugingen-framework overwrites known ones but does
# not delete files that are no longer produced.
rm -f internal/provider/*_gen.go

echo "Generating data source, resource, and provider code..."
tfplugingen-framework generate all \
  --input provider_code_spec.json \
  --output internal/provider \
  --package provider

echo "Renaming generated data source types..."
go run ./tools/dsrename -dir internal/provider

echo "Updating registry and provider registration..."
python3 tools/generate_registry.py

echo "Formatting and building..."
gofmt -s -w internal/provider/*.go
go mod tidy
go build ./...

echo "Done."
