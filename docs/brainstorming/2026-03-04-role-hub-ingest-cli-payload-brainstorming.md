# Role-Hub Ingest API (CLI Payload) Brainstorming

Role: backend-ingest

## Problem Statement
Implement a hard-cut Role Hub ingest API that only accepts the new CLI payload contract, rejects legacy payloads with a clear error, and updates docs/tests accordingly.

## Goals
- Accept only the new payload shape for `POST /api/v1/ingest`.
- Enforce strict validation (including repo format and GitHub URL checks).
- Idempotent processing via `idempotency_key`.
- Provide clear rejection for legacy `roles[]` payloads.
- Update docs and add tests for pass/reject/validation failures.

## Constraints and Assumptions
- Public ingest endpoint; no API key.
- Hard cut: legacy `roles[]` payload must return 400 with `UNSUPPORTED_PAYLOAD_VERSION`.
- `results[].repo` format is strictly `owner/repo` and must match `^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`.
- `results[].source_url` optional; if present must be a GitHub URL.
- Storage can split `owner/repo_name` but external contract remains `owner/repo`.
- Existing `role_records` schema remains the primary storage target.

## Candidate Approaches
### Approach A (Recommended): Hard Cut + Strict Validation + Map to Existing Store
- Replace request models with new payload.
- Reject `roles[]` or unknown shapes with 400/`UNSUPPORTED_PAYLOAD_VERSION`.
- Validate new fields, then map `results[]` to existing `role_records` upsert.
- Do not expand schema; keep `source_url/query/trace_id` out of storage for now.

Trade-offs:
- Pros: minimal schema changes, clear behavior, fastest to deliver.
- Cons: does not persist trace/query metadata.

### Approach B: Hard Cut + Extend DB to Persist Trace/Query/Source URL
- Same as A, but add migrations to store metadata.

Trade-offs:
- Pros: richer audit trail.
- Cons: larger migration/test surface for current scope.

### Approach C: Raw Payload Archive Table
- Store full JSON payload and separately write normalized role_records.

Trade-offs:
- Pros: best for future data analysis.
- Cons: higher complexity; not required for this scope.

## Recommended Design
Use Approach A to keep changes minimal and aligned with the hard-cut requirement.

### Architecture
- API: `POST /api/v1/ingest` accepts only new payload.
- Validation: strict field validation; reject legacy shapes early with `UNSUPPORTED_PAYLOAD_VERSION`.
- Processing: map `results[]` to existing upsert logic for `role_records`.
- Idempotency: reuse existing `ingest_events` lookup by `idempotency_key`.

### Components
- `model`: new `IngestRequest` and `IngestResult` types.
- `api/handler`: parse, detect unsupported payload, validate, idempotency check, upsert.
- `api/validate`: enforce repo pattern, role_path constraints, GitHub URL.
- `db/store`: reuse existing `UpsertRole` mapping.
- `docs`: update API contract and examples.
- `tests`: unit tests for validation; handler tests for accept/reject cases.

### Data Flow
1. Request arrives at `POST /api/v1/ingest`.
2. If legacy shape or missing new fields, return 400 with `UNSUPPORTED_PAYLOAD_VERSION`.
3. Validate new payload fields.
4. If idempotency key has a recorded event, return cached success response.
5. Map each `results[]` into `role_records` upsert.
6. Record ingest event and return 2xx response.

### Error Handling
- Invalid JSON: 400 `INVALID_BODY`.
- Legacy/unknown payload: 400 `UNSUPPORTED_PAYLOAD_VERSION` with readable message.
- Validation errors: 422 `VALIDATION_ERROR` with field violations.
- Upsert failures: count as rejected; still return 200 with counts.

### Testing Strategy
- Handler test: accepts new payload (200).
- Handler test: rejects legacy `roles[]` payload (400 + `UNSUPPORTED_PAYLOAD_VERSION`).
- Validation tests: repo pattern, role_path, source_url (GitHub URL only), required fields.

## Risks and Mitigations
- Risk: breaking older clients.
  - Mitigation: explicit error code and message; docs updated with new contract.
- Risk: malformed repo strings.
  - Mitigation: strict regex validation.

## Open Questions
- None.
