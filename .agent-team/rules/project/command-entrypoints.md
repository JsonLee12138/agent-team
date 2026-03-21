# Project Command Entrypoints

## Root Makefile Entrypoints

Run these from the repository root where `Makefile` exists:

- `make build`
- `make test`
- `make lint`
- `make clean`
- `make install`
- `make uninstall`
- `make migrate`
- `make plugin-pack`
- `make plugin-test`
- `make release V=<x.y.z>`

## Go CLI Entrypoints

This repository's CLI entrypoint is the module root (`go.mod` at repository root):

- `go run . --help`
- `go run . init`
- `go run . rules sync`
- `go run . rules validate`
- `go run . role ...`
- `go run . role-repo ...`
- `go run . worker ...`
- `go run . task ...`
- `go run . workflow plan ...`
- `go run . skill ...`
- `go run . reply ...`
- `go run . reply-main ...`
- `go run . catalog ...`
- `go run . migrate`

## Node Entrypoints

Use root-level `npm --prefix` invocations for nested packages:

- `npm --prefix adapters/opencode run typecheck`
- `npm --prefix role-hub run dev`
- `npm --prefix role-hub run build`
- `npm --prefix role-hub run build:css`
- `npm --prefix role-hub run start`
- `npm --prefix role-hub run lint`
- `npm --prefix role-hub run test`
- `npm --prefix role-hub run e2e`

## Signal Filtering

Detected targets such as `make all`, `make chars`, `make default`, `make fmt`, `make install_deps`, and `make richtest` are present in dependency cache files under `.tmp/gomodcache/...` and are not repo-root command entrypoints for this checkout.
