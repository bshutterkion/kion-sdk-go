# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **Scope: the `go/` lane of the kion-sdk monorepo.** See the repo-root
> `CLAUDE.md` for the monorepo overview.
>
> **You may be reading this in the published mirror.** `scripts/publish-lane.sh`
> copies this lane's tracked files verbatim into
> `kion/delivery-support/dev-tools/kion-sdk/kion-sdk-go`, this file among them,
> so an identical copy sits at the root of that repo. The mirror is **output**,
> not a second source tree and not a legacy repo: it is what `go get
> github.com/kionsoftware/kion-sdk-go` resolves, it is flat (no `go/ python/ ts/`
> subdirectories — those exist only in the monorepo), and CI tags it `vX.Y.Z` at
> publish. Edit this lane; never edit the mirror, whose contents are overwritten
> on the next publish. Same arrangement for `python/` and `ts/`.
>
> The mirror also carries the monorepo-root `CHANGELOG.md` and `VERSION`. It
> never carries `spec/` — mirrors ship the generated client, not the specs it
> was generated from, and publish-lane.sh fails if one appears.
>
> **Regeneration is orchestrated from the repo root**: `make deps`,
> `make refresh-spec VERSION=<v>`, `make refresh-all`, `make all`, and the
> `diff*` targets run at the **root**, not here. This lane's Makefile owns
> `make -C go` targets: `build`, `test`, `ci`, `generate-<v>` (reads
> `../spec/<v>/openapi3.json`), and `scaffold-version`.

## What this is

A typed Go SDK for the [Kion](https://kion.io) API, generated with [ogen](https://ogen.dev) from Kion's OpenAPI spec. It ships **one generated sub-package per supported Kion release** (`generated/v3_14` … `generated/v3_17`, plus `generated/master` for unreleased dev). Consumers import the root `kion` package for shared options/errors AND the sub-package matching their Kion version.

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
make deps                        # install ogen + build go-swagger FROM SOURCE (one-time; version pinned by SWAGGER_VERSION at the root)
make refresh-spec VERSION=v3_15  # pull swagger from portal branch → fixspec → ogen for ONE version
make refresh-all                 # loop refresh-spec over every SDK_VERSIONS entry
make all                         # regenerate every version from already-committed swagger, then build+test
```

`make refresh-*` rewrites files but never touches this repo's git state. A fresh clone builds without portal access because `generated/*/oas_*.go` is committed — as is `spec/<v>/openapi3.json`, at the monorepo root rather than in this lane. Only the raw `spec/<v>/swagger.json` is gitignored.

Comparing versions (oasdiff; pinned `OASDIFF_VERSION` in the Makefile):
```bash
make install-oasdiff             # one-time: go install oasdiff@<pinned> (NOT part of `make deps`)
make diff                        # oasdiff changelog between DIFF_FROM..DIFF_TO (default v3_15 -> v3_16)
make diff-breaking               # breaking changes only
make diff-summary                # high-level change counts
make diff-ops                    # operations added/removed per version
make diff DIFF_FROM=v3_14 DIFF_TO=v3_15   # any pair; applies to all four targets
```

The diff targets read `spec/<v>/openapi3.json`, not the committed generated clients. Those specs are committed, so the targets work in a fresh monorepo checkout; the `_diff-guard` prerequisite fails with a `make refresh-spec VERSION=<v>` hint if one is genuinely missing (a version not yet refreshed).

oasdiff's `go.mod` requires a newer toolchain than the lanes compile against, so `install-oasdiff` builds it under `OASDIFF_BUILD_GOTOOLCHAIN` (`<version>+auto`) rather than the ambient one. Locally that is invisible — `GOTOOLCHAIN` defaults to `auto` — but the CI image sets `GOTOOLCHAIN=local`, which turns the toolchain fetch into a hard failure. Keep the `+auto` suffix: a bare version pin reintroduces exactly that failure the next time `OASDIFF_VERSION` moves to a release needing newer Go.

## The generation pipeline

Portal (Swagger 2.0) → `fixspec` (OpenAPI 3.0 + fixups) → `ogen` (typed Go client). Three moving parts:

1. **go-swagger is compiled from source, never downloaded.** The upstream prebuilt binary has a bug (portal #8218 / go-swagger#2897) that emits empty `definitions`/response schemas, yielding ~250 empty response structs per version. The root `make install-swagger` clones and builds it. Version and build toolchain are pinned at the **root** Makefile as `SWAGGER_VERSION` and `SWAGGER_BUILD_GOTOOLCHAIN` — read them there rather than trusting a number quoted here. If you see many `type XxxResponse struct{}` in generated output, your swagger binary is the broken prebuilt.

2. **`../preprocess/main.go`** (its own Go module at the repo root, *not* `cmd/fixspec/`) — hand-written Go tool that translates Swagger 2.0 → OpenAPI 3.0 and applies ~12 ogen-specific fixups: sanitizing dangling `$ref`s (varies by portal branch), breaking circular refs, marking API-nullable fields the spec declares non-null, retyping decimal fields the spec declares as strings, normalizing operation IDs, adding the API-key security scheme, and promoting query-string-discriminated paths (see #3). Edit this when a new portal branch introduces a spec quirk ogen can't handle.

3. **Synthetic query-string paths.** Swagger 2.0 allows two operations on the same method+path discriminated by a query value (`POST /v3/account?account-type=aws` vs `azure`); OpenAPI 3.0 does not. fixspec rewrites these to a synthetic `/__qs__/key/value` path so ogen sees distinct operations. At runtime the hand-written `queryStringRewriter` transport (in each `generated/*/client.go`) rewrites them back to real query strings. The `queryStringPathMarker` constant is duplicated in `../preprocess/main.go` and every `client.go` — keep them in sync.

## Hand-written vs generated

**Root package** (version-agnostic, no dependency on any generated package, so it never drifts):
- `kion.go` — `ConfigFor`, `NormalizeServerURL` (forces the `/api` suffix), `BuildHTTPClient`; these are called by every sub-package's `New`.
- `options.go` / `errors.go` — `With*` options and `IsNotFound`/`IsAuthError`/`IsConflict`/`StatusCode` (which unwrap ogen's `validate.UnexpectedStatusCodeError`).

**Per generated sub-package**, only two files are hand-written and survive regeneration (`make generate-<v>` rm's `oas_*.go` explicitly rather than using ogen's `-clean`):
- `client.go` — `New(baseURL, opts...)` + the query-string rewriter transport.
- `auth.go` — `bearerAuth`, the ogen `SecuritySource` impl. Both API key and bearer token send `Authorization: Bearer <value>`; auth token wins if both are set.

`generated/master` is the template. `make scaffold-version VERSION=v3_16` sed-copies master's `client.go`/`auth.go` into a new version's package.

## Adding / dropping a Kion version

Adding (portal cut `support-3.18.x`): append `v3_18` to `SDK_VERSIONS` in the **root** Makefile — the only declaration; the portal branch is derived by `scripts/portal-branch.sh`. Then `make check-registries`, which fails and names every hand-maintained list still missing it (the three lane registries plus `ts/package.json` npm exports). Then `make refresh-spec VERSION=v3_18 && make scaffold-version VERSION=v3_18 && make build test`. Finally `make diff DIFF_FROM=v3_17 DIFF_TO=v3_18` to review what the new version added/changed (for the CHANGELOG entry and to scope downstream provider work). Dropping the oldest: remove it from `SDK_VERSIONS`, run `make check-registries` to find every list naming it, delete `spec/<v>/` + `generated/<v>/` + the python and ts trees — this is **breaking** and needs a CHANGELOG entry and version bump. Kion supports 4 versions (current + 3 back), and the SDK ships exactly that — nothing is carried beyond the window (`v3_12` was, until 0.10.0, for no recorded reason).

## Conventions

- **Never bypass hooks.** lefthook runs `gofmt` on pre-commit and the full `make ci` suite on pre-push. The Makefile runs `gofmt -s -w` after ogen so committed and CI-regenerated code are byte-identical (a drift check fails otherwise) — always let formatting run.
- ogen `INFO ... "Type is not defined, using any"` logs during generation are expected (circular/dynamic schemas fall back to `jx.Raw`), not errors.
- After editing a build-tagged file (`cmd/smoketest`, `integration_test.go`) or bumping the version it targets, they only compile under their tag — default `go build`/`go test` skip them by design.
