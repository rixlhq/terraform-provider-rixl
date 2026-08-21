# Code-generation design for terraform-provider-rixl

This document describes how Terraform Plugin Framework code is generated and
kept in sync with the Rixl OpenAPI specification.

## Overview

The provider uses two HashiCorp code-generation tools:

1. `tfplugingen-openapi` — converts `openapi.yaml` into a provider code
   specification (`provider_code_spec.json`).
2. `tfplugingen-framework` — converts `provider_code_spec.json` into Go schema
   and model code.

The generated code lives in `internal/provider/*_gen.go` and is imported by
hand-written data source/resource wrappers and by `internal/provider/registry.go`.

## Why `package provider`?

By default `tfplugingen-framework` creates a separate Go package for every data
source and resource (e.g. `datasource_api_keys`). This provider instead uses the
`--package provider` flag so that all generated types live in the
`internal/provider` package. A single package makes it easy for hand-written
files such as `feed_data_source.go`, `feed_resource.go`, and
`project_resource.go` to reuse generated schemas and models directly.

## Renaming data source models

Because some resources and data sources share the same logical name (e.g.
`project` and `projects`), the generated model names can clash within the same
package. The `tools/dsrename` utility rewrites every `*_data_source_gen.go`
file so that model types receive a `DataSource` suffix:

```
ApiKeysModel          -> ApiKeysDataSourceModel
ProjectDataSourceType -> ProjectDataSourceType   (already suffixed, idempotent)
```

Run it after regenerating data sources:

```bash
go run ./tools/dsrename -dir internal/provider
```

## Files that are generated, kept, or removed

- Generated and kept: all `*_data_source_gen.go` files that have no hand-written
  equivalent.
- Generated but removed:
  - `project_resource_gen.go` and `dashboard_resource_gen.go`: `project` is a
    hand-written resource (`internal/provider/project_resource.go`) and
    `dashboard` contains duplicate `FiltersType` definitions that fail to build.
  - `image_*_gen.go` / `video_*_gen.go` / `images_*_gen.go` / `videos_*_gen.go`:
    excluded from `provider_code_spec.json` because they are implemented by
    hand-written files (`image_data_sources.go`, `video_data_source.go`,
    `videos_data_source.go`, `media_schemas.go`, `media_helpers.go`).
- Hand-written resource wrappers:
  - `feed_resource.go` uses the generated `FeedResourceSchema` and `FeedModel`.
  - `project_resource.go` defines its own `ProjectResourceSchema` and
    `ProjectModel`.
- Hand-written data source wrappers:
  - `feed_data_source.go` and `image_data_sources.go` / `video_data_source.go` /
    `videos_data_source.go` call the generated schema/model functions.

## Registering new data sources

`internal/provider/registry.go` contains a `DataSourceDescriptor` and
`New<Name>DataSource` constructor for every generated data source. The
`tools/generate_registry.py` script can rebuild `registry.go` and update the
`DataSources` method in `internal/provider/provider.go` from
`provider_code_spec.json`:

```bash
python3 tools/generate_registry.py
```

After regenerating, run:

```bash
go build ./...
go test ./...
```

## Manual build flow

```bash
tfplugingen-openapi generate \
  --config generator_config.yml \
  --output provider_code_spec.json \
  openapi.yaml

tfplugingen-framework generate all \
  --input provider_code_spec.json \
  --output internal/provider \
  --package provider

go run ./tools/dsrename -dir internal/provider

rm -f internal/provider/project_resource_gen.go
rm -f internal/provider/dashboard_resource_gen.go

python3 tools/generate_registry.py

go mod tidy
go build ./...
go test ./...
```

## Design choices

- `provider_code_spec.json` and `openapi.yaml` are intentionally ignored by
  Git; they are regenerated on demand.
- `*_gen.go` files are generated artifacts. They may be committed so the
  provider builds without the generator tools installed, but they should not be
  edited by hand.
- The generic `managedDataSource` in `internal/provider/generic.go` implements
  the `datasource.DataSource` interface for every descriptor in `registry.go`.
