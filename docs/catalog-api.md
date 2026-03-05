# Catalog HTTP API

The catalog API exposes role list/search/detail/repository/stats endpoints for frontend consumption.

## Start the server

```bash
agent-team catalog serve --addr :8787
```

Override polling intervals (optional):

```bash
agent-team catalog serve --addr :8787 --normalize-interval 2m --visibility-refresh-interval 10s
```

Base URL (default):

```
http://localhost:8787/api
```

## Polling configuration

Catalog polling defaults are stored in `.agents/catalog-pipeline.yaml` (auto-created on first `catalog serve` run).

```yaml
version: 1
normalize:
  interval: 2m
visibility:
  refresh_interval: 10s
```

- `normalize.interval`: background normalization cadence. Set `0` to disable.
- `visibility.refresh_interval`: API cache refresh cadence. Set `0` to disable caching (always read from disk).

### E2E latency model (discover → verified → visible)

Assuming events arrive uniformly between polls:

```
P50 ≈ normalize.interval/2 + visibility.refresh_interval/2
```

Example comparison (legacy ops defaults vs new defaults):

| Setting | normalize.interval | visibility.refresh_interval | Estimated P50 |
| --- | --- | --- | --- |
| Before (legacy) | 10m | 60s | ~5m30s |
| After (default) | 2m | 10s | ~1m05s |

## Response envelope

All endpoints return JSON with a consistent envelope:

```json
{
  "data": { "...": "..." }
}
```

Errors return:

```json
{
  "error": {
    "code": "invalid_status",
    "message": "invalid status \"bad\"; use: discovered, verified, invalid, unreachable, all"
  }
}
```

### Role object

```json
{
  "name": "frontend",
  "source": "acme/roles",
  "source_type": "github",
  "source_url": "https://github.com/acme/roles",
  "role_path": "skills/frontend",
  "folder_hash": "hash-frontend",
  "status": "verified",
  "status_reason": "",
  "discovered_at": "2026-03-04T03:04:05Z",
  "verified_at": "2026-03-05T01:04:05Z",
  "updated_at": "2026-03-05T03:04:05Z",
  "install_count": 3
}
```

## Endpoints

### List roles

```
GET /roles
GET /roles?status=all
```

- Default `status` is `verified`.
- `status=all` returns all statuses. You may also request `discovered`, `invalid`, `unreachable`.

Response:

```json
{
  "data": {
    "items": [/* Role objects */],
    "total": 12,
    "status": "verified"
  }
}
```

### Search roles

```
GET /roles/search?q=frontend
GET /roles/search?q=frontend&status=all
```

Response:

```json
{
  "data": {
    "items": [/* Role objects */],
    "total": 2,
    "status": "all"
  }
}
```

### Role detail

```
GET /roles/{roleName}
GET /roles/{roleName}?source=owner/repo
GET /roles/{roleName}?status=all
```

Response:

```json
{
  "data": {
    "item": {/* Role object */}
  }
}
```

### Repo detail

```
GET /repos/{owner}/{repo}
GET /repos/{owner}/{repo}?status=all
```

Response:

```json
{
  "data": {
    "repo": "acme/roles",
    "source_url": "https://github.com/acme/roles",
    "items": [/* Role objects */],
    "total": 2,
    "status": "all"
  }
}
```

### Stats

```
GET /stats
```

Response:

```json
{
  "data": {
    "total": 12,
    "repositories": 4,
    "by_status": {
      "verified": 8,
      "discovered": 2,
      "invalid": 1,
      "unreachable": 1
    }
  }
}
```
