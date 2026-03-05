import Database from 'better-sqlite3'
import { Pool } from 'pg'
import { loadConfig } from './config.server'

export interface RoleRow {
  id: number | string
  repo_owner: string
  repo_name: string
  role_path: string
  name: string | null
  description: string | null
  source_url: string | null
  score: number | null
  tags: string | string[] | null
  created_at: string | Date
  updated_at: string | Date
}

interface IngestEventRow {
  response_code: number
  response_body: string
}

const SQLITE_SCHEMA = `
CREATE TABLE IF NOT EXISTS ingest_events (
  idempotency_key TEXT PRIMARY KEY,
  response_code INTEGER NOT NULL,
  response_body TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS role_records (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  repo_owner TEXT NOT NULL,
  repo_name TEXT NOT NULL,
  role_path TEXT NOT NULL,
  name TEXT,
  description TEXT,
  source_url TEXT,
  score REAL,
  tags TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (repo_owner, repo_name, role_path)
);

CREATE INDEX IF NOT EXISTS role_records_repo_idx ON role_records (repo_owner, repo_name);
`

const POSTGRES_SCHEMA = `
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
`

let sqliteDb: Database | null = null
let pgPool: Pool | null = null
let initPromise: Promise<void> | null = null

async function ensureInitialized(): Promise<void> {
  if (!initPromise) {
    initPromise = initialize()
  }
  return initPromise
}

async function initialize(): Promise<void> {
  const config = loadConfig()
  if (config.dbDialect === 'sqlite') {
    const filePath = parseSqlitePath(config.dbDsn)
    sqliteDb = new Database(filePath)
    sqliteDb.exec(SQLITE_SCHEMA)
    return
  }

  pgPool = new Pool({ connectionString: config.dbDsn })
  await pgPool.query(POSTGRES_SCHEMA)
}

function parseSqlitePath(dsn: string): string {
  if (dsn.startsWith('file:')) {
    const withoutPrefix = dsn.slice(5)
    const [path] = withoutPrefix.split('?')
    return path || ':memory:'
  }
  return dsn
}

export async function listRoleRows(): Promise<RoleRow[]> {
  await ensureInitialized()
  if (sqliteDb) {
    return sqliteDb.prepare('SELECT * FROM role_records').all() as RoleRow[]
  }
  if (!pgPool) return []
  const { rows } = await pgPool.query('SELECT * FROM role_records')
  return rows as RoleRow[]
}

export async function listRoleRowsByRepo(owner: string, repo: string): Promise<RoleRow[]> {
  await ensureInitialized()
  if (sqliteDb) {
    return sqliteDb
      .prepare('SELECT * FROM role_records WHERE repo_owner = ? AND repo_name = ?')
      .all(owner, repo) as RoleRow[]
  }
  if (!pgPool) return []
  const { rows } = await pgPool.query('SELECT * FROM role_records WHERE repo_owner = $1 AND repo_name = $2', [owner, repo])
  return rows as RoleRow[]
}

export async function getIngestEvent(idempotencyKey: string): Promise<IngestEventRow | null> {
  await ensureInitialized()
  if (sqliteDb) {
    const row = sqliteDb
      .prepare('SELECT response_code as response_code, response_body as response_body FROM ingest_events WHERE idempotency_key = ?')
      .get(idempotencyKey) as IngestEventRow | undefined
    return row ?? null
  }
  if (!pgPool) return null
  const { rows } = await pgPool.query(
    'SELECT response_code as response_code, response_body as response_body FROM ingest_events WHERE idempotency_key = $1',
    [idempotencyKey],
  )
  return rows[0] ?? null
}

export async function insertIngestEvent(idempotencyKey: string, responseCode: number, responseBody: string): Promise<void> {
  await ensureInitialized()
  if (sqliteDb) {
    sqliteDb
      .prepare('INSERT INTO ingest_events (idempotency_key, response_code, response_body) VALUES (?,?,?)')
      .run(idempotencyKey, responseCode, responseBody)
    return
  }
  if (!pgPool) return
  await pgPool.query(
    'INSERT INTO ingest_events (idempotency_key, response_code, response_body) VALUES ($1,$2,$3)',
    [idempotencyKey, responseCode, responseBody],
  )
}

export interface RoleRecordInput {
  repo_owner: string
  repo_name: string
  role_path: string
  name: string
  description: string
  source_url?: string
  score?: number | null
  tags?: string[]
}

export async function upsertRoleRecords(records: RoleRecordInput[]): Promise<(Error | null)[]> {
  await ensureInitialized()
  const errors: (Error | null)[] = records.map(() => null)
  if (records.length === 0) return errors

  if (sqliteDb) {
    const stmt = sqliteDb.prepare(`
      INSERT INTO role_records (repo_owner, repo_name, role_path, name, description, source_url, score, tags, created_at, updated_at)
      VALUES (?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
      ON CONFLICT (repo_owner, repo_name, role_path)
      DO UPDATE SET name=excluded.name,
                    description=excluded.description,
                    source_url=excluded.source_url,
                    score=excluded.score,
                    tags=excluded.tags,
                    updated_at=CURRENT_TIMESTAMP;
    `)

    const tx = sqliteDb.transaction((items: RoleRecordInput[]) => {
      items.forEach((record, index) => {
        try {
          const tags = record.tags && record.tags.length > 0 ? JSON.stringify(record.tags) : null
          stmt.run(
            record.repo_owner,
            record.repo_name,
            record.role_path,
            record.name,
            record.description,
            record.source_url ?? null,
            record.score ?? null,
            tags,
          )
        } catch (error) {
          errors[index] = error as Error
        }
      })
    })

    tx(records)
    return errors
  }

  if (!pgPool) return errors

  const client = await pgPool.connect()
  try {
    await client.query('BEGIN')
    const stmt = `
      INSERT INTO role_records (repo_owner, repo_name, role_path, name, description, source_url, score, tags, created_at, updated_at)
      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
      ON CONFLICT (repo_owner, repo_name, role_path)
      DO UPDATE SET name=EXCLUDED.name,
                    description=EXCLUDED.description,
                    source_url=EXCLUDED.source_url,
                    score=EXCLUDED.score,
                    tags=EXCLUDED.tags,
                    updated_at=CURRENT_TIMESTAMP;
    `

    for (const [index, record] of records.entries()) {
      try {
        const tags = record.tags && record.tags.length > 0 ? JSON.stringify(record.tags) : null
        await client.query(stmt, [
          record.repo_owner,
          record.repo_name,
          record.role_path,
          record.name,
          record.description,
          record.source_url ?? null,
          record.score ?? null,
          tags,
        ])
      } catch (error) {
        errors[index] = error as Error
      }
    }

    await client.query('COMMIT')
    return errors
  } catch (error) {
    await client.query('ROLLBACK')
    return errors.map((existing) => existing ?? (error as Error))
  } finally {
    client.release()
  }
}
