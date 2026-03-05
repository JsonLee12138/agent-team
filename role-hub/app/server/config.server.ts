export type DbDialect = 'sqlite' | 'postgres'

export interface RoleHubConfig {
  dbDialect: DbDialect
  dbDsn: string
  dbTimeoutMs: number
  rateLimitRps: number
  rateLimitBurst: number
  maxBodyBytes: number
  maxResults: number
  maxInflight: number
}

const DEFAULTS = {
  dbDialect: 'sqlite' as DbDialect,
  dbDsn: 'file:rolehub.db?cache=shared',
  dbTimeoutMs: 3000,
  rateLimitRps: 5,
  rateLimitBurst: 10,
  maxBodyBytes: 1 << 20,
  maxResults: 500,
  maxInflight: 100,
}

export function loadConfig(): RoleHubConfig {
  const dbDialect = (process.env.ROLE_HUB_DB_DIALECT || DEFAULTS.dbDialect) as DbDialect
  const dbDsn = process.env.ROLE_HUB_DB_DSN || DEFAULTS.dbDsn

  return {
    dbDialect: dbDialect === 'postgres' ? 'postgres' : 'sqlite',
    dbDsn,
    dbTimeoutMs: parseDurationMs(process.env.ROLE_HUB_DB_TIMEOUT) ?? DEFAULTS.dbTimeoutMs,
    rateLimitRps: parseNumber(process.env.ROLE_HUB_RATE_LIMIT_RPS) ?? DEFAULTS.rateLimitRps,
    rateLimitBurst: parseNumber(process.env.ROLE_HUB_RATE_LIMIT_BURST) ?? DEFAULTS.rateLimitBurst,
    maxBodyBytes: parseNumber(process.env.ROLE_HUB_MAX_BODY_BYTES) ?? DEFAULTS.maxBodyBytes,
    maxResults: parseNumber(process.env.ROLE_HUB_MAX_RESULTS) ?? DEFAULTS.maxResults,
    maxInflight: parseNumber(process.env.ROLE_HUB_MAX_INFLIGHT) ?? DEFAULTS.maxInflight,
  }
}

function parseNumber(value?: string): number | null {
  if (!value) return null
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : null
}

function parseDurationMs(value?: string): number | null {
  if (!value) return null
  const match = /^(\d+)(ms|s|m|h)?$/i.exec(value.trim())
  if (!match) return null
  const amount = Number(match[1])
  const unit = match[2]?.toLowerCase() ?? 'ms'
  if (!Number.isFinite(amount)) return null
  switch (unit) {
    case 'ms':
      return amount
    case 's':
      return amount * 1000
    case 'm':
      return amount * 60_000
    case 'h':
      return amount * 3_600_000
    default:
      return null
  }
}
