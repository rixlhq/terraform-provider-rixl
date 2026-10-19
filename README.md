# Rixl Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/Terraform%20Registry-rixlhq%2Frixl-844FBA?logo=terraform&logoColor=white)](https://registry.terraform.io/providers/rixlhq/rixl/latest)
[![Release](https://img.shields.io/github/v/release/rixlhq/terraform-provider-rixl?sort=semver)](https://github.com/rixlhq/terraform-provider-rixl/releases)
[![CI](https://github.com/rixlhq/terraform-provider-rixl/actions/workflows/ci.yml/badge.svg)](https://github.com/rixlhq/terraform-provider-rixl/actions/workflows/ci.yml)

The Rixl provider allows Terraform to manage [Rixl](https://rixl.com) platform resources.

## Requirements

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- [Go](https://go.dev/) >= 1.26
- [mise](https://mise.jdx.dev/) (recommended for running tasks and installing tools)

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

This project uses [mise](https://mise.jdx.dev/) to manage tools and tasks. [lefthook](https://github.com/evilmartians/lefthook) is configured for pre-commit and pre-push checks.

Install tools and set up git hooks:

```sh
mise install
```

Run default local checks (format, build, test, lint):

```sh
mise run
```

Run individual tasks:

```sh
mise run build
mise run test
mise run lint
mise run fmt
mise run docs
```

## Publishing to the Terraform Registry

The repository is set up to publish releases that the Terraform Registry can consume.

- The provider address is `registry.terraform.io/rixlhq/rixl`.
- `terraform-registry-manifest.json` declares protocol version `6.0`.
- `.goreleaser.yml` builds cross-platform archives, a SHA256 checksum file, a detached GPG signature, and renames the manifest to the required asset name.
- `.github/workflows/ci.yml` runs lint, test, and build in parallel steps on every push and pull request.
- `.github/workflows/release.yml` runs release-please on successful CI on `main` to create draft releases, then runs `goreleaser/goreleaser-action` to upload signed release artifacts, and finally publishes the release atomically for the Terraform Registry.
