# Role Hub Remix Full-Stack Migration

## Summary
- Renamed legacy backend `role-hub` to `hub-server` (Go ingest service preserved).
- Created new `role-hub` Remix full-stack app to own UI + `/api/v1` routes.
- Migrated UI pages and interactions (Home/Directory, Role Detail, Repo Detail, Empty State) into Remix routes.
- Added server-side ingest + query endpoints with hub-server-compatible validation and idempotency.
- Added explicit error states and root error boundary to prevent white screens.

## Local Development
1. `cd role-hub`
2. `npm install`
3. Optional environment variables (defaults provided):
- `ROLE_HUB_DB_DIALECT` = `sqlite` or `postgres` (default `sqlite`)
- `ROLE_HUB_DB_DSN` = `file:rolehub.db?cache=shared` (sqlite default) or Postgres connection string
- `ROLE_HUB_DB_TIMEOUT` = `3s`
- `ROLE_HUB_RATE_LIMIT_RPS` = `5`
- `ROLE_HUB_RATE_LIMIT_BURST` = `10`
- `ROLE_HUB_MAX_BODY_BYTES` = `1048576`
- `ROLE_HUB_MAX_RESULTS` = `500`
- `ROLE_HUB_MAX_INFLIGHT` = `100`
4. `npm run dev`

## API Surface (Remix)
- `GET /api/v1/roles` (query: `search`, `category`, `framework`, `sort`, `page`, `limit`)
- `GET /api/v1/roles/:name`
- `GET /api/v1/repos`
- `GET /api/v1/repos/:owner/:repo`
- `POST /api/v1/ingest`

`POST /api/v1/ingest` uses the same payload contract and validation rules defined in `hub-server/docs/ingest-api.md`.

## Deployment
1. `cd role-hub`
2. `npm run build`
3. `npm run start`

Recommended: set `ROLE_HUB_DB_DIALECT` and `ROLE_HUB_DB_DSN` in the deployment environment. The app will auto-create tables on startup if they do not exist.

## Verification Results (2026-03-05)
- `cd role-hub && npm run lint` passed
- `cd role-hub && npm run build` passed (React Router v7 future flag warnings)
- `cd role-hub && npm run test` passed (no tests found, `--passWithNoTests`)
- `cd role-hub && npm run e2e` passed (no tests found, `--pass-with-no-tests`)
