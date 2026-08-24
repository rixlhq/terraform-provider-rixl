# Agent notes for terraform-provider-rixl

## Build / test / lint

```bash
go build ./...
go test -count=1 ./...
golangci-lint run
```

## Code generation

`gen.sh` regenerates `provider_code_spec.json` and `internal/provider/*_gen.go` from
`openapi.yaml` and `generator_config.yml` using `tfplugingen-openapi` and
`tfplugingen-framework`.

```bash
./gen.sh
```

Notes:
- `provider_code_spec.json` is a generated artifact and is not tracked in git.
- `tools/postprocess_spec.py` adjusts the generated spec (e.g. removes `secret`
  from data sources and fixes resource attribute modifiers).
- `tools/generate_registry.py` regenerates `internal/provider/registry.go` and
  the resource/data-source registration in `provider.go`.
