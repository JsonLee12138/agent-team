# Role Hub Ingest API

## Endpoint

`POST /api/v1/ingest`

Hard-cut to the new payload contract. Legacy `roles[]` payloads are rejected with `400 UNSUPPORTED_PAYLOAD_VERSION`.

## Payload

```json
{
  "idempotency_key": "uuid-or-unique-key",
  "trace_id": "trace-id",
  "timestamp": "2026-03-05T00:00:00Z",
  "query": "search query",
  "result_count": 1,
  "results": [
    {
      "repo": "owner/repo",
      "role_path": "skills/backend",
      "name": "backend",
      "description": "Backend role",
      "source_url": "https://github.com/owner/repo",
      "score": 0.98,
      "tags": ["backend", "go"]
    }
  ]
}
```

### Field Rules

- `idempotency_key`: required, max 128 chars
- `trace_id`: required
- `timestamp`: required RFC3339
- `query`: required
- `result_count`: must equal `len(results)`
- `results[].repo`: required, strict `owner/repo` format (`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
- `results[].role_path`: required, path segments only (`^[A-Za-z0-9_.-]+(/[A-Za-z0-9_.-]+)*$`)
- `results[].source_url`: optional, must be `github.com` URL if present

## Response

```json
{
  "status": "ok",
  "idempotency_key": "uuid-or-unique-key",
  "accepted": 1,
  "rejected": 0,
  "errors": []
}
```

If some rows fail upsert, `rejected` is incremented and `errors` includes per-row messages.

## Errors

- `400 INVALID_BODY`: invalid or empty JSON
- `400 UNSUPPORTED_PAYLOAD_VERSION`: legacy `roles[]` payload
- `422 VALIDATION_ERROR`: field validation failed
- `429 RATE_LIMITED`: too many requests
- `429 CONCURRENCY_LIMIT`: too many concurrent requests

Example:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "payload validation failed",
    "missing_fields": ["idempotency_key", "results[0].repo"],
    "details": [
      {"field": "results[0].repo", "message": "must be owner/repo"}
    ]
  }
}
```

## Local Run

```bash
export ROLE_HUB_DB_DIALECT=sqlite
export ROLE_HUB_DB_DSN='file:rolehub.db?cache=shared'
export ROLE_HUB_HTTP_ADDR=':8080'
export ROLE_HUB_DB_MAX_OPEN_CONNS=20
export ROLE_HUB_DB_MAX_IDLE_CONNS=10
export ROLE_HUB_DB_CONN_MAX_LIFETIME=30m
export ROLE_HUB_DB_CONN_MAX_IDLE_TIME=5m
export ROLE_HUB_DB_TIMEOUT=3s
export ROLE_HUB_MAX_INFLIGHT=100

cd role-hub

go run ./cmd/role-hub migrate

go run ./cmd/role-hub serve
```

## Load Test

```bash
ROLE_HUB_LOADTEST_TARGET=http://localhost:8080/api/v1/ingest \\
ROLE_HUB_LOADTEST_REQUESTS=500 \\
ROLE_HUB_LOADTEST_CONCURRENCY=50 \\
ROLE_HUB_LOADTEST_TIMEOUT=5s \\
go run ./cmd/loadtest
```

## cURL Example

```bash
curl -sS -X POST http://localhost:8080/api/v1/ingest \
  -H 'Content-Type: application/json' \
  -d '"'"'{
    "idempotency_key":"demo-1",
    "trace_id":"trace-1",
    "timestamp":"2026-03-05T00:00:00Z",
    "query":"search",
    "result_count":1,
    "results":[
      {
        "repo":"acme/roles",
        "role_path":"skills/backend",
        "name":"backend",
        "description":"Backend role",
        "source_url":"https://github.com/acme/roles",
        "score":0.92,
        "tags":["backend","go"]
      }
    ]
  }'"'"
```
