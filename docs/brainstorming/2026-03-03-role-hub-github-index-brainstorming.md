# Role Management Site Brainstorming

**Date:** 2026-03-03  
**Role Used:** general strategist  
**Status:** Approved

## Problem Statement and Goals

Build a non-SaaS role management website (similar to skills.sh interaction style) with:

- GitHub as the source of truth for all role content.
- A hosted database for list/search/index data.
- No self-managed servers.
- Tight integration with current `agent-team` CLI behavior, especially: ingest role candidates immediately after `role-repo find`.

## Constraints and Assumptions

- Project is open source and cost-sensitive.
- Must avoid dedicated server ops.
- Data model must support `find`-time ingestion (not only post-install).
- Selected DB direction: Postgres only (no Redis for MVP).
- Chosen deployment model: Vercel-hosted web app + managed Postgres provider (Neon via Vercel Marketplace).
- Hard requirement: CLI ingest is async and must never block/impact `find` UX.

## Candidate Approaches with Trade-offs

### A. Separate repository for web app (Recommended)

- Description: New `role-hub` repo for Next.js full-stack web app; keep `agent-team` CLI in current repo; connect via ingest API.
- Pros:
  - Clear separation of concerns.
  - Independent release cadence for CLI and web.
  - Lower CI/CD coupling risk between Go and TS stacks.
- Cons:
  - Two repos to maintain.

### B. Monorepo inside current `agent-team` repository

- Description: Add web app under current repo (e.g., `apps/role-web`).
- Pros:
  - Easier local code sharing and single repo visibility.
- Cons:
  - Higher build/test/release coupling.
  - More complex pipeline and dependency management.

### C. GitHub-only static index (no DB)

- Description: Generate list directly from GitHub only.
- Pros:
  - Lowest infra complexity.
- Cons:
  - Does not meet requirement for DB-backed listing/search behavior.

## Recommended Design

Adopt **Approach A**: separate `role-hub` repository with GitHub truth + Postgres index.

### Architecture

- `agent-team` (current repo): continues CLI ownership.
- `role-hub` (new repo): Next.js app on Vercel (frontend + API).
- Source of truth: GitHub role repos/files.
- Index/search storage: Neon Postgres.

### Components

- `agent-team role-repo find`
  - After returning find results, asynchronously POST candidates to `role-hub` ingest API.
  - On failure, only warn/log; never affect command success.
- `role-hub ingest API`
  - HMAC signature + timestamp verification.
  - Batch upsert with idempotency key: `repo + role_path + ref`.
- `role-hub normalize/sync worker`
  - Fetch GitHub metadata and validate role structure.
  - Promote records from discovered pool to verified catalog when valid.
- `role-hub public API/frontend`
  - Query verified catalog for pagination, sorting, filtering, search.

### Data Flow

1. User runs `role-repo find`.
2. CLI outputs results immediately.
3. CLI asynchronously sends discovered candidates to `role-hub`.
4. Ingest writes/upserts `discovered` records.
5. Background normalization validates and enriches from GitHub.
6. Valid records appear in `catalog` for website listing.
7. Webhook/scheduled jobs keep data fresh.

## Error Handling Strategy

- Ingest timeout/error: non-blocking for CLI; retry later.
- Duplicate discoveries: idempotent upsert, counted as deduplicated.
- GitHub rate limits: exponential backoff and retry queue.
- Invalid role structure: mark `invalid`, keep diagnostics.
- Repo removed/private: mark `unreachable`, keep audit trail.
- Auth/signature failures: reject immediately and log security event.
- DB outages: ingest fast-fail with observability alerts; CLI remains unaffected.

## Validation and Testing Strategy

- CLI unit tests:
  - Ensure async ingest trigger after `find`.
  - Ensure `find` result behavior unchanged under ingest failures.
- API unit tests:
  - Signature/timestamp validation.
  - Idempotent batch upsert behavior.
- Integration tests:
  - End-to-end path `find -> ingest -> normalize -> catalog`.
  - Rate-limit and invalid-structure cases.
- E2E checks:
  - Real find ingestion visibility in site list within expected delay.
- Operational metrics:
  - Ingest success rate, dedupe rate, normalize latency, invalid ratio.

## Risks and Mitigations

- Cross-repo integration complexity
  - Mitigation: strict API contract + shared schema docs.
- GitHub API volatility/rate-limit pressure
  - Mitigation: cached metadata + retry and backoff policies.
- Data quality drift from discovered-only records
  - Mitigation: explicit status lifecycle (`discovered`, `verified`, `invalid`, `unreachable`).
- Cost overrun risk on free tiers
  - Mitigation: monitor query volume and storage growth from day one.

## Open Questions

- Whether website needs direct user edit/submit flow, or stays read/manage-only with GitHub-driven changes.
- Exact admin operations for moderation (blacklist, manual verify, forced resync).
- SLA target for discovery-to-visible latency.
