# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `make check-versions` + `scripts/check-versions.sh`: warns when portal has a `support-3.NN.x` branch not tracked by `SDK_VERSIONS` (a new Kion version was cut and needs `refresh-spec` + `scaffold-version`), or when `SDK_VERSIONS` lists a version whose portal branch is gone. Cheap `git ls-remote` check — no checkout, no regeneration. `refresh-and-commit` runs it as a non-fatal heads-up before regenerating.
- **Doc-comments on generated schema types** (all versions). `fixspec` now promotes portal's descriptive schema `title`s (full sentences like "`BudgetCreate` is used to create a new project budget…") to `description` via a new `cleanDescriptiveTitles` pass; ogen emits `description` as Go doc-comments, so the generated types gain ~2,250 lines of documentation. Purely additive — no type, field, signature, or behavior change. This also readies the shared OpenAPI spec for non-Go SDK generators (e.g. Python), which name classes from `title` and would otherwise produce sentence-length class names.

### Changed

- **BREAKING**: `Scope{Create,CriteriaCreate,CriteriaRecord,CriteriaUpdate}.Criteria` change from `*Scope…Criteria` (an empty struct) to `jx.Raw` (raw JSON), and the four `Scope…Criteria` types are removed (`master`, `v3_15`). This is the regeneration finally materializing the already-merged `fixRawJSONFields` handling (freeform "raw JSON" objects modeled as `jx.Raw`) that the committed clients had never been regenerated against. Consumers reading/writing those fields must switch to `jx.Raw`.
- **BREAKING**: `GetAzureRoleIndex` and `GetAzureRoleIndex2` swap signatures (`master`, `v3_13`, `v3_15`): `GetAzureRoleIndex(ctx)` no longer takes a `params` argument, and `GetAzureRoleIndex2(ctx, params)` now does. Brings the committed client in line with the current portal spec (completing the azure-role signature change 0.6.0 already announced).
- `make install-ogen` now pins ogen to `v1.20.1` (matching `go.mod`) instead of `@latest`, so local and CI regeneration no longer drift the generated output toward whatever ogen release is newest (`@latest` had reached v1.23.0, a ~26-file diff vs the pinned generator).
- The derived OpenAPI specs' `info.version` is normalized to `0.0.0` (new `neutralizeInfoVersion` pass) instead of the Go-specific `kion-sdk-go-<v>`. Affects only the gitignored `spec/*/openapi3.json` consumed by future non-Go generators (`kion-sdk-go-<v>` is not a valid PEP 440 version) — **no effect on any committed Go**.
- The smoke test (`cmd/smoketest`) and integration suite (`integration_test.go`) now target `generated/v3_16` (current release) instead of `v3_15`.

> The two BREAKING items above are source-breaking to the generated Go API and warrant a minor version bump (→ `0.7.0`) at the next release.

## [0.6.0] - 2026-07-06

### Changed

- **BREAKING**: Regenerated every version (`master`, `v3_12`–`v3_16`) with **go-swagger 0.33.1** (bumped from 0.30.5 to match portal and to build/run under Go 1.25 — 0.30.5's type importer panicked with `unsupported version: 2` on Go 1.25 export data). The newer generator re-renders operations differently, so **method signatures change even on API-unchanged branches** — e.g. `GetAzureRoleIndex` no longer takes a `params` argument on v3_12–v3_16. Consumers upgrading across this release should re-check the affected call sites. The regenerated clients now match what each portal branch's source actually declares.

### Added

- New operations captured on **master**: `OU Note` CRUD (`GetOUNote`, `GetOUNotes`, `PostOUNote`, `PatchOUNote`, `DeleteOUNote`), `ScopeCriteria` (`PostScopeCriteria`, `PatchScopeCriteria`, `DeleteScopeCriteria`), `PostAccountLinkage`, `PostGoogleAccountLinkage`, `GetGCPRegionsList`, and `DeleteOU`.
- New operations captured on **v3_15**: `ScopeCriteria` (`PostScopeCriteria`, `PatchScopeCriteria`, `DeleteScopeCriteria`), `PostAccountLinkage`, `PostGoogleAccountLinkage` — operations present in `support-3.15.x`'s spec that go-swagger 0.30.5 was silently dropping and 0.33.1 now emits.

## [0.5.0] - 2026-07-06

### Changed

- **BREAKING**: Module path reverted from `git.kion.io/kion/delivery-support/dev-tools/terraform-provider/kion-sdk-go` back to `github.com/kionsoftware/kion-sdk-go` (reverses the 0.4.0 rename). The SDK is consumed by the Kion Terraform provider under the `github.com/kionsoftware` import path — matching Kion's other published modules — so the internal-only `git.kion.io` path is reverted. Consumers must update imports back to `github.com/kionsoftware/kion-sdk-go[/generated/<v>]`.

### Added

- `make refresh-and-commit` + `scripts/refresh-and-commit.sh`: regenerate every supported version from its portal support branch and commit only the versions whose generated client actually changed. Change detection is **stateless** — it diffs the regenerated `generated/<v>/` against the committed client (the source of truth; `spec/*` is gitignored), so there is no per-branch SHA to track. Drives the existing `refresh-spec` target, so it runs unchanged from any scheduler. Adds `make print-versions` (exposes `SDK_VERSIONS`); pass `ARGS=--no-commit` to regenerate without committing.

## [0.4.2] - 2026-07-06

### Added

- `make diff-ops` target + `scripts/diff-ops.sh`: lists operations added/removed between two SDK versions from the **committed generated client** (`generated/<v>/oas_client_gen.go`) — no OpenAPI spec or oasdiff required, so it runs on a bare clone with zero setup. Groups added operations by resource family to answer "which operations/resources are new" (e.g. for scaffolding downstream consumers like the Terraform provider). Complements `make diff` (oasdiff on specs), which covers field-level and breaking changes; `diff-ops` is deliberately blind to those. Honors the same `DIFF_FROM`/`DIFF_TO` defaults (v3_15 → v3_16).

## [0.4.1] - 2026-07-06

### Added

- `make diff`, `make diff-breaking`, and `make diff-summary` targets to compare the API between two SDK versions using [oasdiff](https://github.com/oasdiff/oasdiff) (pinned via `OASDIFF_VERSION`, installed with `make install-oasdiff` — intentionally not part of `make deps` since it is an analysis tool, not a codegen dependency). `diff` shows the full changelog, `diff-breaking` only breaking changes, `diff-summary` a high-level count. Defaults to `DIFF_FROM=v3_15 DIFF_TO=v3_16`; override either on the command line to compare any pair (applies to all three targets). A shared `_diff-guard` prerequisite verifies oasdiff is installed and both `spec/<v>/openapi3.json` files exist — since specs are gitignored/derived, it points to `make refresh-spec VERSION=<v>` when one is missing.

## [0.4.0] - 2026-07-06

### Changed

- **BREAKING**: Module path renamed from `github.com/kionsoftware/kion-sdk-go` to `git.kion.io/kion/delivery-support/dev-tools/terraform-provider/kion-sdk-go`. The SDK is an internal-only tool hosted on GitLab and is not published to GitHub, so the module path now matches its actual location and is resolvable with `go get` (set `GOPRIVATE=git.kion.io/*`). Update all imports accordingly.

### Added

- `generated/v3_16` sub-package for Kion 3.16.x, generated from `portal/support-3.16.x`. Kion 3.16 is now the current supported release; `v3_12` through `v3_16` are all active (`v3_12` is carried one release beyond the standard current-plus-3 window). Wired `v3_16` into the Makefile's `SDK_VERSIONS` and `refresh-spec` portal-branch lookup.

## [0.3.0] - 2026-04-23

### Fixed

- `fixspec`: handle real-world null values and paginated Azure list responses. Three spec-vs-reality mismatches were causing ogen's strict decoders to reject real Kion responses:
  - `aws_iam_permissions_boundary` on `OUCloudAccessRoleFull` and `ProjectCloudAccessRoleFull` is declared as an `IAMPolicy` `$ref` but the API returns `null` for most records. `shouldBeNullable` now recognizes this field, and `fixNullableFields` handles `$ref` properties correctly (nullability applies only to the specific property via `allOf + nullable`, not to the shared target schema).
  - `AzurePolicyListResponse` / `AzureRoleListResponse` declared `data` as a flat array (the v3 shape), but the `/v4/azure-policy` and `/v4/azure-role` endpoints return `{pagination, total, items}` (v4 paginated). New `fixAzurePaginatedResponses` creates sibling `*Paginated` response schemas and rebinds only the `/v4` operations to them; `/v3` paths keep their original flat response.
  - Boolean fields and the `pagination` wrapper frequently come back `null` for disabled records or non-paginating endpoints that share the paginated envelope. `shouldBeNullable` now covers all `boolean`-typed schemas and the named `pagination` property; pagination is inlined via `allOf + nullable` in the Azure rewrites to also catch nested inline schemas.
- Regenerated all supported SDK versions (`master`, `v3_12`, `v3_13`, `v3_14`, `v3_15`) with the above fixes.

## [0.2.0] - 2026-04-09

### Changed

- **BREAKING**: Restructured from a single `generated/` package to per-version sub-packages (`generated/v3_12`, `generated/v3_13`, `generated/v3_14`, `generated/v3_15`, `generated/master`). Consumers must update imports from `github.com/kionsoftware/kion-sdk-go/generated` to a version-specific sub-package (e.g. `github.com/kionsoftware/kion-sdk-go/generated/v3_15`).
- **BREAKING**: `kion.NewClient` removed from the root package. Each sub-package now exports its own `New` constructor (e.g. `v3_15.New(baseURL, kion.WithAPIKey(...))`). The root `kion` package retains only version-agnostic shared code: `Config`, `Option`, `With*`, error helpers, `ConfigFor`, `NormalizeServerURL`, `BuildHTTPClient`.
- `auth.go` removed from root package. Each sub-package has its own `bearerAuth` type implementing the ogen-generated `SecuritySource` interface.
- VERSION file now tracks SDK version (semver) rather than Kion API version.

### Added

- Per-version sub-packages supporting Kion 3.12 through 3.15, plus a `master` development target tracking portal's unreleased API.
- Multi-version consumer support: a single program can import multiple sub-packages and talk to different Kion instances running different versions simultaneously.
- Makefile targets: `refresh-spec VERSION=<v>`, `refresh-all`, `scaffold-version VERSION=<v>`, `install-swagger`, `help`.
- `install-swagger` target compiles go-swagger v0.30.5 from source to work around upstream prebuilt binary bug (portal issue 8218, go-swagger#2897) that produces swagger files with empty definitions and response properties.
- `scaffold-version` target generates per-version `client.go` and `auth.go` wrappers from the `generated/master/` template via sed substitution.
- Swagger spec refresh invokes `swagger generate spec` the same way portal's CI does (`devops/portal-ci/jobs/build_swagger.sh`): from `portal/src` with `cloudtamer.io/app/webapi` as the positional import path and an `info.version` template.
- `cmd/fixspec`: new `sanitizeBrokenRefs` preprocessing step that replaces dangling `$ref` pointers to undefined responses, parameters, and schemas with inline empty objects before kin-openapi parses the swagger. Handles portal support-3.13.x's missing `CloudProviderType` response reference.
- `cmd/fixspec`: query-string-discriminated paths (e.g. `POST /v3/account?account-type=aws`) are now promoted to synthetic distinct paths under a `/__qs__/` marker instead of being collapsed into a single operation. The old merge logic silently dropped 9 out of 13 query-string operations per version, losing SDK methods like `PostAwsAccount`, `PostAzureSubscription`, and `PostGoogleCloudAccount`.
- Each sub-package's `client.go` installs a `queryStringRewriter` HTTP RoundTripper that detects the synthetic `/__qs__/` path marker and rewrites URLs to the real query-string form before sending requests to Kion.
- `cmd/smoketest` and `integration_test.go` placed behind build tags (`smoketest` and `integration` respectively) to isolate per-version schema drift from the default build.
- `OptStringToFrameworkLegacy` flex helper (in terraform-provider, not the SDK itself — noted here for completeness).
- Comprehensive README documenting the multi-version layout, Makefile workflow, go-swagger source-build rationale, adding new versions, consumer import patterns, and troubleshooting.

### Fixed

- fixspec no longer collapses query-string-discriminated Swagger paths, which was silently dropping SDK methods across all versions.
- fixspec now tolerates portal swagger specs with dangling `$ref` pointers instead of failing at the Swagger 2.0 to OpenAPI 3.0 conversion step.

## [0.1.0] - 2026-03-02

### Added

- ogen-generated typed client with 1,289 methods from 654 Kion API operations
- Spec fixup tool (`cmd/fixspec`) converting Swagger 2.0 to OpenAPI 3.0 with fixes for:
  - Query-string paths moved to proper query parameters
  - Circular `$ref` resolution (FeatureFlagData)
  - Schema name conflict resolution
  - Operation ID normalization (strip `apiV1`/`APIBeta` prefixes)
  - Missing response schema backfill
- Convenience client constructor (`kion.NewClient`) with functional options
- API key and bearer token authentication via `SecuritySource`
- `WithAPIKey`, `WithBearerToken`, `WithSkipVerify`, `WithTimeout`, `WithHTTPClient` options
- Error helpers: `IsNotFound`, `IsAuthError`, `IsConflict`, `StatusCode`
- VERSION file for tracking target Kion API version (3.15.1)
- Integration tests covering 18 list endpoints verified against live Kion API
- Smoke test CLI (`cmd/smoketest`) for quick manual verification
