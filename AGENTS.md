Reference guide for contributors and automated agents working on this repository.

## Architecture

- Terraform **provider** (Go) — NOT a module. Manages Altinity.Cloud environments.
- Built with **Terraform Plugin Framework** (not SDKv2).
- Backend is a **GraphQL API**; the SDK is code-generated.
- Go 1.26.2. Module: `github.com/altinity/terraform-provider-altinitycloud`

## Provider Configuration

- `api_url` — API endpoint. Env var `ALTINITYCLOUD_API_URL`. Default `https://anywhere.altinity.cloud`. GraphQL path `/api/v1/graphql`.
- `api_token` — auth token. Env var `ALTINITYCLOUD_API_TOKEN`. Required for real API calls (`make testacc`, manual `local/` testing).
- Auth logic lives in `internal/sdk/auth/`.

## Project Structure

```
internal/provider/
  provider.go                        # provider entrypoint
  common/                            # shared schema attributes, doc strings, helpers
  env/{aws,azure,gcp,hcloud,k8s}/    # per-cloud resources (model, schema, resource, data_source)
  env/common/                        # shared env models and SDK mappers
  env_certificate/
  env_secret/
  env_status/
  modifiers/                         # plan modifiers
  validators/                        # schema validators
internal/sdk/client/
  graphql.schema                     # GraphQL schema
  *.graphql                          # queries/mutations
  # generated SDK lives here too
docs/
  resources/
  data-sources/
  index.md                           # generated; do not edit by hand
```

## Development Commands

| Command | What it does |
|---|---|
| `make build` | `go build` |
| `make test` | `go test -v -cover ./...` |
| `make testacc` | Acceptance tests (`TF_ACC=1`), 120m timeout, hits real API |
| `make lint` | `golangci-lint run` |
| `make fmt` | `go fmt ./...` + `terraform fmt -recursive` |
| `make local` | Build + install to `~/.terraform.d/plugins/local/altinity/...` for local manual testing |
| `make sdk` | Fetch PROD GraphQL schema via curl, regenerate SDK with gqlgenc |
| `make docs` | Regenerate docs via tfplugindocs |
| `make gen` | `sdk` + `docs` |
| `make bump type=major\|minor\|patch` | Git-tag a new version |
| `make install-hooks` | Install pre-commit hook (golangci-lint via `.githooks`) |

Always run `make fmt` and `make lint` before committing. If you ran `make install-hooks`, the pre-commit hook enforces lint automatically.

## Key Design Patterns

- **No HasChange.** Plugin Framework has no SDKv2-style `HasChange`. For partial updates, compare `req.Plan` vs `req.State` manually.
- **Full-state replace.** The standard update pattern sends the entire plan to the API. GraphQL `UpdateStrategy` is `REPLACE` or `MERGE`. This is idiomatic for Plugin Framework — not a shortcut.
- **Model types.** Use `types.String`, `types.Bool`, `types.Int64` (Plugin Framework types), not raw Go primitives in models.
- **toSDK / toModel.** `toSDK()` converts model → SDK create/update inputs. `toModel()` converts SDK response → model. Keep this split clean.
- **Reordering functions.** Used to preserve user config ordering in state and avoid spurious plan diffs.
- **Env-specific vs shared.** Code referencing env-specific GraphQL fragment types (e.g., `AWSEnvSpecFragment_Iceberg`) stays in the per-env package. Code using shared SDK input types is hoisted into `env/common/`.

## SDK Regeneration

- Config: `internal/sdk/client/gqlgenc.yml`. Tool: `github.com/Yamashou/gqlgenc`.
- `make sdk` runs from repo root (Makefile handles the `cd`). It pulls the **PROD** schema by default.
- Dev-only GraphQL fields are absent from the prod schema. To generate against them: surgically edit the local `graphql.schema` file, then run `gqlgenc` locally — do **not** use `make sdk` (which would overwrite with prod schema).
- Never hand-edit generated SDK files. Re-run `make sdk` instead.

## Adding a New Cloud Environment

1. Create `internal/provider/env/<name>/` with four files: `model.go`, `schema.go`, `resource.go`, `data_source.go`.
2. Follow the layout of an existing env (e.g., `aws` or `gcp`) as a template.
3. Env-specific GraphQL fragment types stay in the new package. SDK input conversion that uses shared types goes in `internal/provider/env/common/`.
4. Register the new resource and data source in `internal/provider/provider.go`.
5. Add GraphQL queries/mutations to `internal/sdk/client/` and run `make sdk` to regenerate.
6. Run `make docs` to regenerate provider documentation.

## Documentation & Examples

- `docs/` is **generated** by `tfplugindocs` — never hand-edit. `tfplugindocs` reads from `examples/` (`provider/`, `resources/`, `data-sources/`) and doc strings in `internal/provider/common/docs.go`.
- To change docs: edit the relevant `examples/` HCL and/or `docs.go`, then run `make docs`.
- `make sync` injects the current version into `docs/*.md` (replaces `%%VERSION%%`).

## CI & Release

- `test.yml` (push/PR): `go build`, `make test`, then `make gen` + `git diff` — **generated SDK and docs must be committed**; CI fails on drift. Uses `GRAPHQL_SCHEMA_URL` secret (can point at dev schema).
- `sanity-test.yml`: acceptance/sanity checks.
- `release.yml`: on tag push `v*` → runs sanity tests, then GoReleaser (self-hosted runner). Tag via `make bump type=...`.

## Code Style

- **No comments by default.** Never restate what code does. Comment only a non-obvious WHY — hidden invariant, workaround for a specific bug, external constraint. One short line max; no doc-blocks on types or short functions.
- Shared logic belongs in `common/` packages. Do not duplicate across env packages.
- `make fmt` is mandatory. `make lint` must pass cleanly.
- Acceptance tests (`make testacc`) require live API credentials and are slow — run them intentionally, not on every change.
