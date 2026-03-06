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
  install_count: number
  tags: string | string[] | null
  system_md: string | null
  skills_json: string | string[] | null
  in_scope_json: string | string[] | null
  out_of_scope_json: string | string[] | null
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
  install_count INTEGER NOT NULL DEFAULT 0,
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
  install_count INTEGER NOT NULL DEFAULT 0,
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
    migrateSqlite(sqliteDb)
    return
  }

  pgPool = new Pool({ connectionString: config.dbDsn })
  await pgPool.query(POSTGRES_SCHEMA)
  await migratePostgres(pgPool)
}

function migrateSqlite(db: Database): void {
  const cols = db.pragma('table_info(role_records)') as { name: string }[]
  const colNames = new Set(cols.map((c) => c.name))
  if (!colNames.has('install_count')) {
    db.exec('ALTER TABLE role_records ADD COLUMN install_count INTEGER NOT NULL DEFAULT 0')
  }
  if (!colNames.has('system_md')) {
    db.exec('ALTER TABLE role_records ADD COLUMN system_md TEXT')
  }
  if (!colNames.has('skills_json')) {
    db.exec('ALTER TABLE role_records ADD COLUMN skills_json TEXT')
  }
  if (!colNames.has('in_scope_json')) {
    db.exec('ALTER TABLE role_records ADD COLUMN in_scope_json TEXT')
  }
  if (!colNames.has('out_of_scope_json')) {
    db.exec('ALTER TABLE role_records ADD COLUMN out_of_scope_json TEXT')
  }
}

async function migratePostgres(pool: Pool): Promise<void> {
  await pool.query(`
    ALTER TABLE role_records ADD COLUMN IF NOT EXISTS install_count INTEGER NOT NULL DEFAULT 0;
    ALTER TABLE role_records ADD COLUMN IF NOT EXISTS system_md TEXT;
    ALTER TABLE role_records ADD COLUMN IF NOT EXISTS skills_json JSONB;
    ALTER TABLE role_records ADD COLUMN IF NOT EXISTS in_scope_json JSONB;
    ALTER TABLE role_records ADD COLUMN IF NOT EXISTS out_of_scope_json JSONB;
  `)
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
  system_md?: string
  skills?: string[]
  in_scope?: string[]
  out_of_scope?: string[]
}

export async function upsertRoleRecords(records: RoleRecordInput[]): Promise<(Error | null)[]> {
  await ensureInitialized()
  const errors: (Error | null)[] = records.map(() => null)
  if (records.length === 0) return errors

  if (sqliteDb) {
    const stmt = sqliteDb.prepare(`
      INSERT INTO role_records (repo_owner, repo_name, role_path, name, description, source_url, score, tags, system_md, skills_json, in_scope_json, out_of_scope_json, created_at, updated_at)
      VALUES (?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
      ON CONFLICT (repo_owner, repo_name, role_path)
      DO UPDATE SET name=excluded.name,
                    description=excluded.description,
                    source_url=excluded.source_url,
                    score=excluded.score,
                    tags=excluded.tags,
                    system_md=COALESCE(excluded.system_md, role_records.system_md),
                    skills_json=COALESCE(excluded.skills_json, role_records.skills_json),
                    in_scope_json=COALESCE(excluded.in_scope_json, role_records.in_scope_json),
                    out_of_scope_json=COALESCE(excluded.out_of_scope_json, role_records.out_of_scope_json),
                    updated_at=CURRENT_TIMESTAMP;
    `)

    const tx = sqliteDb.transaction((items: RoleRecordInput[]) => {
      items.forEach((record, index) => {
        try {
          const tags = record.tags && record.tags.length > 0 ? JSON.stringify(record.tags) : null
          const skillsJson = record.skills && record.skills.length > 0 ? JSON.stringify(record.skills) : null
          const inScopeJson = record.in_scope && record.in_scope.length > 0 ? JSON.stringify(record.in_scope) : null
          const outOfScopeJson = record.out_of_scope && record.out_of_scope.length > 0 ? JSON.stringify(record.out_of_scope) : null
          stmt.run(
            record.repo_owner,
            record.repo_name,
            record.role_path,
            record.name,
            record.description,
            record.source_url ?? null,
            record.score ?? null,
            tags,
            record.system_md ?? null,
            skillsJson,
            inScopeJson,
            outOfScopeJson,
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
      INSERT INTO role_records (repo_owner, repo_name, role_path, name, description, source_url, score, tags, system_md, skills_json, in_scope_json, out_of_scope_json, created_at, updated_at)
      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
      ON CONFLICT (repo_owner, repo_name, role_path)
      DO UPDATE SET name=EXCLUDED.name,
                    description=EXCLUDED.description,
                    source_url=EXCLUDED.source_url,
                    score=EXCLUDED.score,
                    tags=EXCLUDED.tags,
                    system_md=COALESCE(EXCLUDED.system_md, role_records.system_md),
                    skills_json=COALESCE(EXCLUDED.skills_json, role_records.skills_json),
                    in_scope_json=COALESCE(EXCLUDED.in_scope_json, role_records.in_scope_json),
                    out_of_scope_json=COALESCE(EXCLUDED.out_of_scope_json, role_records.out_of_scope_json),
                    updated_at=CURRENT_TIMESTAMP;
    `

    for (const [index, record] of records.entries()) {
      try {
        const tags = record.tags && record.tags.length > 0 ? JSON.stringify(record.tags) : null
        const skillsJson = record.skills && record.skills.length > 0 ? JSON.stringify(record.skills) : null
        const inScopeJson = record.in_scope && record.in_scope.length > 0 ? JSON.stringify(record.in_scope) : null
        const outOfScopeJson = record.out_of_scope && record.out_of_scope.length > 0 ? JSON.stringify(record.out_of_scope) : null
        await client.query(stmt, [
          record.repo_owner,
          record.repo_name,
          record.role_path,
          record.name,
          record.description,
          record.source_url ?? null,
          record.score ?? null,
          tags,
          record.system_md ?? null,
          skillsJson,
          inScopeJson,
          outOfScopeJson,
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

export interface InstallCountKey {
  repo_owner: string
  repo_name: string
  role_path: string
}

export async function incrementInstallCounts(keys: InstallCountKey[]): Promise<void> {
  if (keys.length === 0) return
  await ensureInitialized()

  if (sqliteDb) {
    const stmt = sqliteDb.prepare(`
      UPDATE role_records
      SET install_count = install_count + 1, updated_at = CURRENT_TIMESTAMP
      WHERE repo_owner = ? AND repo_name = ? AND role_path = ?
    `)
    const tx = sqliteDb.transaction((items: InstallCountKey[]) => {
      for (const key of items) {
        stmt.run(key.repo_owner, key.repo_name, key.role_path)
      }
    })
    tx(keys)
    return
  }

  if (!pgPool) return
  const client = await pgPool.connect()
  try {
    await client.query('BEGIN')
    for (const key of keys) {
      await client.query(
        `UPDATE role_records
         SET install_count = install_count + 1, updated_at = CURRENT_TIMESTAMP
         WHERE repo_owner = $1 AND repo_name = $2 AND role_path = $3`,
        [key.repo_owner, key.repo_name, key.role_path],
      )
    }
    await client.query('COMMIT')
  } catch {
    await client.query('ROLLBACK')
  } finally {
    client.release()
  }
}
