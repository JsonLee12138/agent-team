import { loadConfig } from './config.server'
import { getIngestEvent, incrementInstallCounts, insertIngestEvent, upsertRoleRecords, type RoleRecordInput } from './db.server'

export interface IngestRequest {
  idempotency_key: string
  trace_id: string
  timestamp: string
  query: string
  result_count: number
  results: IngestResult[]
}

export interface IngestResult {
  repo: string
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

export interface IngestResponse {
  status: 'ok'
  idempotency_key: string
  accepted: number
  rejected: number
  errors?: IngestFailure[]
}

export interface IngestFailure {
  index: number
  repo: string
  message: string
}

export interface FieldDetail {
  field: string
  message: string
}

export interface ErrorResponse {
  error: {
    code: string
    message: string
    details?: FieldDetail[]
    missing_fields?: string[]
  }
}

const repoPattern = /^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/
const rolePathPattern = /^[A-Za-z0-9_.-]+(\/[A-Za-z0-9_.-]+)*$/

class RateLimiter {
  private rps: number
  private burst: number
  private items = new Map<string, { tokens: number; lastSeen: number }>()

  constructor(rps: number, burst: number) {
    this.rps = rps
    this.burst = burst
  }

  allow(key: string): boolean {
    const now = Date.now()
    const existing = this.items.get(key)
    const entry = existing ?? { tokens: this.burst, lastSeen: now }
    const elapsed = Math.max(0, (now - entry.lastSeen) / 1000)
    entry.tokens = Math.min(this.burst, entry.tokens + elapsed * this.rps)
    entry.lastSeen = now

    if (entry.tokens < 1) {
      this.items.set(key, entry)
      return false
    }

    entry.tokens -= 1
    this.items.set(key, entry)
    return true
  }

  cleanup(maxAgeMs: number) {
    const cutoff = Date.now() - maxAgeMs
    for (const [key, entry] of this.items.entries()) {
      if (entry.lastSeen < cutoff) {
        this.items.delete(key)
      }
    }
  }
}

let rateLimiter: RateLimiter | null = null
let inflight = 0

export async function handleIngestRequest(request: Request): Promise<Response> {
  const config = loadConfig()
  if (request.method !== 'POST') {
    return jsonError(405, 'METHOD_NOT_ALLOWED', 'only POST is supported')
  }

  const bodyBuffer = await request.arrayBuffer()
  if (config.maxBodyBytes > 0 && bodyBuffer.byteLength > config.maxBodyBytes) {
    return jsonError(413, 'PAYLOAD_TOO_LARGE', 'request body too large')
  }

  let raw: unknown
  try {
    raw = JSON.parse(new TextDecoder().decode(bodyBuffer))
  } catch {
    return jsonError(400, 'INVALID_BODY', 'invalid JSON payload')
  }

  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) {
    return jsonError(400, 'INVALID_BODY', 'invalid JSON payload')
  }

  if ('roles' in raw) {
    return jsonError(400, 'UNSUPPORTED_PAYLOAD_VERSION', 'legacy roles[] payload is not supported')
  }

  const missingFields = findMissingFields(raw as Record<string, unknown>)
  const issues = validateRequest(raw as IngestRequest, config.maxResults)
  if (issues.length > 0 || missingFields.length > 0) {
    return jsonError(422, 'VALIDATION_ERROR', 'payload validation failed', {
      details: issues,
      missing_fields: missingFields.length > 0 ? missingFields : undefined,
    })
  }

  const req = raw as IngestRequest
  const cached = await getIngestEvent(req.idempotency_key)
  if (cached) {
    return new Response(cached.response_body, {
      status: cached.response_code,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  const clientKey = extractClientKey(request)
  if (!rateLimiter) {
    rateLimiter = new RateLimiter(config.rateLimitRps, config.rateLimitBurst)
  }
  rateLimiter.cleanup(30 * 60 * 1000)
  if (!rateLimiter.allow(clientKey)) {
    return jsonError(429, 'RATE_LIMITED', 'rate limit exceeded')
  }

  if (inflight >= config.maxInflight) {
    return jsonError(429, 'CONCURRENCY_LIMIT', 'too many concurrent requests')
  }

  inflight += 1
  try {
    const response = await processIngest(req)
    const body = JSON.stringify(response)
    await insertIngestEvent(req.idempotency_key, 200, body)
    return new Response(body, {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Idempotency-Key': req.idempotency_key,
      },
    })
  } finally {
    inflight = Math.max(0, inflight - 1)
  }
}

async function processIngest(req: IngestRequest): Promise<IngestResponse> {
  const records: RoleRecordInput[] = req.results.map((result) => mapToRecord(result))
  const errors = await upsertRoleRecords(records)

  if (req.query.startsWith('install:')) {
    const installKeys = req.results.map((result) => {
      const [owner = '', repo = ''] = result.repo.split('/')
      return { repo_owner: owner, repo_name: repo, role_path: result.role_path }
    })
    await incrementInstallCounts(installKeys)
  }

  const response: IngestResponse = {
    status: 'ok',
    idempotency_key: req.idempotency_key,
    accepted: 0,
    rejected: 0,
    errors: [],
  }

  req.results.forEach((result, index) => {
    const error = errors[index]
    if (error) {
      response.rejected += 1
      response.errors?.push({ index, repo: result.repo, message: 'upsert failed' })
    } else {
      response.accepted += 1
    }
  })

  if (response.errors && response.errors.length === 0) {
    delete response.errors
  }

  return response
}

function mapToRecord(result: IngestResult): RoleRecordInput {
  const [owner = '', repo = ''] = result.repo.split('/')
  return {
    repo_owner: owner,
    repo_name: repo,
    role_path: result.role_path,
    name: result.name,
    description: result.description,
    source_url: result.source_url,
    score: result.score ?? null,
    tags: result.tags ?? [],
    system_md: result.system_md,
    skills: result.skills,
    in_scope: result.in_scope,
    out_of_scope: result.out_of_scope,
  }
}

function validateRequest(req: IngestRequest, maxResults: number): FieldDetail[] {
  const issues: FieldDetail[] = []
  if (!req.idempotency_key?.trim()) {
    issues.push({ field: 'idempotency_key', message: 'required' })
  }
  if (req.idempotency_key && req.idempotency_key.length > 128) {
    issues.push({ field: 'idempotency_key', message: 'too long' })
  }
  if (!req.trace_id?.trim()) {
    issues.push({ field: 'trace_id', message: 'required' })
  }
  if (!req.timestamp?.trim()) {
    issues.push({ field: 'timestamp', message: 'required' })
  } else if (!isValidRFC3339(req.timestamp)) {
    issues.push({ field: 'timestamp', message: 'must be RFC3339' })
  }
  if (!req.query?.trim()) {
    issues.push({ field: 'query', message: 'required' })
  }
  if (typeof req.result_count !== 'number' || req.result_count < 0) {
    issues.push({ field: 'result_count', message: 'must be >= 0' })
  }
  if (req.result_count !== req.results.length) {
    issues.push({ field: 'result_count', message: `expected ${req.result_count} results` })
  }
  if (req.results.length > maxResults) {
    issues.push({ field: 'results', message: `max ${maxResults} results` })
  }
  req.results.forEach((result, index) => {
    issues.push(...validateResult(result, index))
  })
  return issues
}

function validateResult(result: IngestResult, index: number): FieldDetail[] {
  const issues: FieldDetail[] = []
  const prefix = `results[${index}]`
  if (!result.repo?.trim()) {
    issues.push({ field: `${prefix}.repo`, message: 'required' })
  } else if (!repoPattern.test(result.repo)) {
    issues.push({ field: `${prefix}.repo`, message: 'must be owner/repo' })
  }
  if (!result.role_path?.trim()) {
    issues.push({ field: `${prefix}.role_path`, message: 'required' })
  } else if (!rolePathPattern.test(result.role_path)) {
    issues.push({ field: `${prefix}.role_path`, message: 'invalid role path' })
  }
  if (result.source_url && !isValidGithubUrl(result.source_url)) {
    issues.push({ field: `${prefix}.source_url`, message: 'must be a github.com URL' })
  }
  if (result.tags && result.tags.length > 64) {
    issues.push({ field: `${prefix}.tags`, message: 'too many tags' })
  }
  return issues
}

function isValidRFC3339(value: string): boolean {
  const date = new Date(value)
  return !Number.isNaN(date.getTime()) && value.includes('T')
}

function isValidGithubUrl(value: string): boolean {
  try {
    const parsed = new URL(value)
    if (!['http:', 'https:'].includes(parsed.protocol)) return false
    if (parsed.hostname.toLowerCase() !== 'github.com') return false
    const parts = parsed.pathname.split('/').filter(Boolean)
    return parts.length >= 2
  } catch {
    return false
  }
}

function findMissingFields(raw: Record<string, unknown>): string[] {
  const missing = new Set<string>()
  checkRequired(raw, missing, 'idempotency_key', 'trace_id', 'timestamp', 'query', 'result_count', 'results')

  const rawResults = raw.results
  if (Array.isArray(rawResults)) {
    rawResults.forEach((item, index) => {
      if (!item || typeof item !== 'object' || Array.isArray(item)) return
      const prefix = `results[${index}]`
      checkRequired(item as Record<string, unknown>, missing, `${prefix}.repo`, `${prefix}.role_path`)
    })
  }

  return Array.from(missing).sort()
}

function checkRequired(raw: Record<string, unknown>, missing: Set<string>, ...fields: string[]) {
  fields.forEach((field) => {
    const key = field.includes('.') ? field.slice(field.lastIndexOf('.') + 1) : field
    if (!(key in raw) || isMissing(raw[key])) {
      missing.add(field)
    }
  })
}

function isMissing(value: unknown): boolean {
  if (value === null || value === undefined) return true
  if (typeof value === 'string' && value.trim().length === 0) return true
  return false
}

function extractClientKey(request: Request): string {
  const forwarded = request.headers.get('x-forwarded-for')
  if (forwarded) return forwarded.split(',')[0].trim()
  const realIp = request.headers.get('x-real-ip')
  if (realIp) return realIp
  return 'unknown'
}

function jsonError(status: number, code: string, message: string, extras?: { details?: FieldDetail[]; missing_fields?: string[] }) {
  const payload: ErrorResponse = {
    error: {
      code,
      message,
      details: extras?.details,
      missing_fields: extras?.missing_fields,
    },
  }
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}
