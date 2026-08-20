# Rixl Terraform Provider

The Rixl provider allows Terraform to manage [Rixl](https://rixl.com) platform resources.

## Requirements

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- [Go](https://go.dev/) >= 1.26.5 (for local development)

## Usage

```hcl
terraform {
  required_providers {
    rixl = {
      source  = "rixlhq/rixl"
      version = ">= 0.0.0"
    }
  }
}

provider "rixl" {
  api_key = var.rixl_api_key
}

resource "rixl_feed" "main" {
  project_id  = var.rixl_project_id
  name        = "main"
  description = "Primary content feed"
}
```

## Authentication

The provider supports two authentication methods:

- `api_key` — sent as the `X-API-Key` header.
- `bearer_token` — sent as `Authorization: Bearer <token>`.

These can also be supplied via environment variables:

- `RIXL_API_KEY`
- `RIXL_BEARER_TOKEN`
- `RIXL_BASE_URL` (defaults to `https://api.rixl.com`)

## Development

```bash
go test ./...
go build -o terraform-provider-rixl .
```

Regenerate the provider from the OpenAPI specification with:

```bash
./gen.sh
```

## Releasing

Releases are managed with [release-please](https://github.com/googleapis/release-please) and [GoReleaser](https://goreleaser.com). A `GPG_PRIVATE_KEY` and `GPG_PASSPHRASE` secret are required for signing release artifacts before publishing to the Terraform Registry.

## OpenAPI-driven updates

When `rixlhq/openapi` changes, it dispatches a `regenerate-sdk` event to this repository, which runs `./gen.sh` and commits the regenerated provider code.
