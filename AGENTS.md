<!-- bmad:context -->
<!-- Verified 2026-09-19 against bbf39dde5d773e490b1cce4c0d8d74169c265731. Managed by bmad-project-context; edits inside this block are replaced on refresh. Keep anything you want preserved outside the markers. -->

## Betor Search

Betor Search is a Go 1.26 HTTP service that exposes a BeTor-backed search API and health endpoint from `cmd/server`. It keeps the in-memory catalog sync and search logic in `internal/search`, exposes the health contract in `internal/health`, and keeps planning artifacts and operational docs under `_bmad-output/` and `docs/`.

## Policy

- Keep the HTTP bootstrap in `cmd/server`, business logic in `internal/*/application`, and handlers in `internal/*/transport`.
- Treat configuration as environment-driven; default port is `8080`, and secrets or credentials must not be committed to the repo.
- Preserve the `/health` contract and the public search route semantics unless code and tests are updated together.

## Where things are

- Runtime entry point: `cmd/server/main.go`
- Search service and sync logic: `internal/search/application/service.go`
- Health service and response contract: `internal/health/application/service.go`, `internal/health/contract/response.go`
- Prowlarr/Cardigann definition: `betor-search.yml`
- Planning and repo docs: `_bmad-output/`, `docs/prowlarr/`, `README.md`

## Running and verifying

- Use `go test ./...` for the project suite.
- Use `go build -trimpath -ldflags "-s -w" -o betor-search ./cmd/server` for the release-style binary.
- Run locally with `go run ./cmd/server`; default port is `8080`, override with `PORT=9090`.
- CI verifies formatting with `gofmt -w ./... && git diff --exit-code`, then runs `go test ./...`, then builds the binary.
- Do not assume a bare `go run .` or a different wrapper; the repo entry point is `./cmd/server`.

## Conventions that differ from defaults

- Keep the HTTP port configurable via `PORT` with a default of `8080`; do not hard-code the port in handlers or service code.
- Use env-driven overrides for sync behavior: `BETOR_SEARCH_API_BASE_URL`, `BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE`, `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES`, `BETOR_SEARCH_DOWNLOAD_ITEMS_URL`, and `BETOR_SEARCH_COMPONENT_NAME`.
- Treat `BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE` as a pre-encoded Basic auth payload; empty means omit the header.
- Keep service catalog sync in memory and refresh on the configured interval; do not persist catalog state into the repo.

## Known pitfalls

- The server reads environment variables at startup and uses them to set the catalog sync interval and auth behavior; changing them during runtime is not the intended path.
- A missing or empty `BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE` is valid and must not trigger an auth header.
- The default health endpoint is `/health`; changing the route or the response shape without updating tests and callers is a compatibility break.

<!-- /bmad:context -->
