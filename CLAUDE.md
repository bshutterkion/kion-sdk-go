# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **Scope: the `go/` lane of the kion-sdk monorepo.** See the repo-root
> `CLAUDE.md` for the monorepo overview. Two things moved out of this directory
> and the notes below still describe the old standalone layout:
> - **`fixspec` now lives in `../preprocess/`** (its own Go module), not
>   `cmd/fixspec/`. The queryStringPathMarker constant is shared between
>   `../preprocess/main.go` and every `generated/*/client.go`.
> - **Regeneration is orchestrated from the repo root**: `make deps`,
>   `make refresh-spec VERSION=<v>`, `make refresh-all`, `make all`, and the
>   `diff*` targets run at the **root**, not here. This lane's Makefile owns
>   `make -C go` targets: `build`, `test`, `ci`, `generate-<v>` (reads
>   `../spec/<v>/openapi3.json`), and `scaffold-version`.

## What this is

A typed Go SDK for the [Kion](https://kion.io) API, generated with [ogen](https://ogen.dev) from Kion's OpenAPI spec. It ships **one generated sub-package per supported Kion release** (`generated/v3_12` … `generated/v3_16`, plus `generated/master` for unreleased dev). Consumers import the root `kion` package for shared options/errors AND the sub-package matching their Kion version.

The single most important architectural fact: **almost all Go code here is generated, not hand-written.** Do not hand-edit `generated/*/oas_*.go` — those are ogen output and are overwritten by `make generate-<v>`. Only a handful of files are authored by humans (see below).

## Commands

```bash
make build        # go build ./... (root + every generated sub-package; excludes build-tagged files)
make test         # unit tests only (go test ./...) — exercises root package helpers, no generated deps
make ci           # full local mirror of GitLab CI: fmt-check, vet, lint, test (run before pushing)
make lint         # golangci-lint run
make fmt          # gofmt -s -w .  (also runs automatically on commit via lefthook)
```

Run a single test: `go test -run TestName -v ./...` (root package) — there is no per-package split worth memorizing; unit tests live in `kion_test.go`.

Build-tagged suites (excluded from default build/test because they reference version-specific generated fields):
```bash
make build-smoketest      # go build -tags smoketest ./cmd/smoketest ; then KION_URL=… KION_API_KEY=… ./smoketest
make test-integration     # go test -tags integration ./...  (needs KION_URL + KION_API_KEY)
```

Regeneration (needs a clean `portal` checkout at `PORTAL_DIR`, default `../../../../portal`, plus `make deps`):
```bash
make deps                        # install ogen + build go-swagger v0.30.5 FROM SOURCE (one-time)
make refresh-spec VERSION=v3_15  # pull swagger from portal branch → fixspec → ogen for ONE version
make refresh-all                 # loop refresh-spec over every SDK_VERSIONS entry
make all                         # regenerate every version from already-committed swagger, then build+test
```

`make refresh-*` rewrites files but never touches this repo's git state. A fresh clone builds without portal access because `generated/*/oas_*.go` is committed; `spec/*/*.json` is gitignored (derived).

Comparing versions (oasdiff; pinned `OASDIFF_VERSION` in the Makefile):
```bash
make install-oasdiff             # one-time: go install oasdiff@<pinned> (NOT part of `make deps`)
make diff                        # oasdiff changelog between DIFF_FROM..DIFF_TO (default v3_15 -> v3_16)
make diff-breaking               # breaking changes only
make diff-summary                # high-level change counts
make diff DIFF_FROM=v3_14 DIFF_TO=v3_15   # any pair; applies to all three targets
```

Because `spec/*/openapi3.json` is gitignored, the diff targets need those specs present — run `make refresh-spec VERSION=<v>` first (the `_diff-guard` prerequisite fails with that hint if a spec is missing). oasdiff reads the derived `spec/<v>/openapi3.json`, not the committed generated clients.

## The generation pipeline

Portal (Swagger 2.0) → `fixspec` (OpenAPI 3.0 + fixups) → `ogen` (typed Go client). Three moving parts:

1. **go-swagger is compiled from source at v0.30.5, never downloaded.** The upstream prebuilt binary has a bug (portal #8218 / go-swagger#2897) that emits empty `definitions`/response schemas, yielding ~250 empty response structs per version. `make install-swagger` clones and builds it (pinning `GOTOOLCHAIN=go1.21.13` for the build only). If you see many `type XxxResponse struct{}` in generated output, your swagger binary is the broken prebuilt.

2. **`cmd/fixspec/main.go`** — hand-written Go tool that translates Swagger 2.0 → OpenAPI 3.0 and applies ~12 ogen-specific fixups: sanitizing dangling `$ref`s (varies by portal branch), breaking circular refs, marking API-nullable fields the spec declares non-null, normalizing operation IDs, adding the API-key security scheme, and promoting query-string-discriminated paths (see #3). Edit this when a new portal branch introduces a spec quirk ogen can't handle.

3. **Synthetic query-string paths.** Swagger 2.0 allows two operations on the same method+path discriminated by a query value (`POST /v3/account?account-type=aws` vs `azure`); OpenAPI 3.0 does not. fixspec rewrites these to a synthetic `/__qs__/key/value` path so ogen sees distinct operations. At runtime the hand-written `queryStringRewriter` transport (in each `generated/*/client.go`) rewrites them back to real query strings. The `queryStringPathMarker` constant is duplicated in `cmd/fixspec/main.go` and every `client.go` — keep them in sync.

## Hand-written vs generated

**Root package** (version-agnostic, no dependency on any generated package, so it never drifts):
- `kion.go` — `ConfigFor`, `NormalizeServerURL` (forces the `/api` suffix), `BuildHTTPClient`; these are called by every sub-package's `New`.
- `options.go` / `errors.go` — `With*` options and `IsNotFound`/`IsAuthError`/`IsConflict`/`StatusCode` (which unwrap ogen's `validate.UnexpectedStatusCodeError`).

**Per generated sub-package**, only two files are hand-written and survive regeneration (`make generate-<v>` rm's `oas_*.go` explicitly rather than using ogen's `-clean`):
- `client.go` — `New(baseURL, opts...)` + the query-string rewriter transport.
- `auth.go` — `bearerAuth`, the ogen `SecuritySource` impl. Both API key and bearer token send `Authorization: Bearer <value>`; auth token wins if both are set.

`generated/master` is the template. `make scaffold-version VERSION=v3_16` sed-copies master's `client.go`/`auth.go` into a new version's package.

## Adding / dropping a Kion version

Adding (portal cut `support-3.16.x`): append `v3_16` to `SDK_VERSIONS` in the Makefile, add a `v3_16) portal=support-3.16.x` case to `refresh-spec`, then `make refresh-spec VERSION=v3_16 && make scaffold-version VERSION=v3_16 && make build test`. Then `make diff DIFF_FROM=v3_15 DIFF_TO=v3_16` to review what the new version added/changed (for the CHANGELOG entry and to scope downstream provider work). Dropping the oldest: remove it from `SDK_VERSIONS`, delete `spec/<v>/` + `generated/<v>/`, remove its `refresh-spec` case. Kion supports 4 versions (current + 3 back).

## Conventions

- **Never bypass hooks.** lefthook runs `gofmt` on pre-commit and the full `make ci` suite on pre-push. The Makefile runs `gofmt -s -w` after ogen so committed and CI-regenerated code are byte-identical (a drift check fails otherwise) — always let formatting run.
- ogen `INFO ... "Type is not defined, using any"` logs during generation are expected (circular/dynamic schemas fall back to `jx.Raw`), not errors.
- After editing a build-tagged file (`cmd/smoketest`, `integration_test.go`) or bumping the version it targets, they only compile under their tag — default `go build`/`go test` skip them by design.
