CREATE TABLE IF NOT EXISTS ingest_events (
  idempotency_key TEXT PRIMARY KEY,
  response_code INTEGER NOT NULL,
  response_body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS role_records (
  id BIGSERIAL PRIMARY KEY,
  repo_owner TEXT NOT NULL,
  repo_name TEXT NOT NULL,
  role_path TEXT NOT NULL,
  name TEXT,
  description TEXT,
  source_url TEXT,
  score DOUBLE PRECISION,
  tags JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (repo_owner, repo_name, role_path)
);

CREATE INDEX IF NOT EXISTS role_records_repo_idx ON role_records (repo_owner, repo_name);
