import * as fs from 'fs-extra'
import * as path from 'path'
import { Config } from '@oclif/core'

/**
 * Test helper utilities for oclif commands
 */

/**
 * Create a temporary test directory
 */
export async function createTestDir(name: string): Promise<string> {
  const dir = path.join(__dirname, 'fixtures', name)
  await fs.ensureDir(dir)
  return dir
}

/**
 * Clean up test directory
 */
export async function cleanTestDir(dir: string): Promise<void> {
  await fs.remove(dir)
}

/**
 * Create a test fixture file
 */
export async function createFixture(
  dir: string,
  filename: string,
  content: string
): Promise<string> {
  const filePath = path.join(dir, filename)
  await fs.writeFile(filePath, content)
  return filePath
}

/**
 * Read fixture file
 */
export async function readFixture(dir: string, filename: string): Promise<string> {
  const filePath = path.join(dir, filename)
  return fs.readFile(filePath, 'utf-8')
}

/**
 * Create test config
 */
export async function createTestConfig(overrides?: any): Promise<any> {
  const config = {
    version: '1.0.0',
    settings: {
      verbose: false,
      timeout: 30000,
    },
    ...overrides,
  }
  return config
}

/**
 * Mock stdin input
 */
export function mockStdin(input: string): void {
  const originalStdin = process.stdin
  const Readable = require('stream').Readable
  const stdin = new Readable()
  stdin.push(input)
  stdin.push(null)
  // @ts-ignore
  process.stdin = stdin
}

/**
 * Restore stdin
 */
export function restoreStdin(): void {
  // Restore original stdin if needed
}

/**
 * Capture stdout
 */
export class StdoutCapture {
  private originalWrite: any
  private output: string[] = []

  start(): void {
    this.output = []
    this.originalWrite = process.stdout.write
    process.stdout.write = ((chunk: any, encoding?: any, callback?: any) => {
      this.output.push(chunk.toString())
      return true
    }) as any
  }

  stop(): void {
    process.stdout.write = this.originalWrite
  }

  getOutput(): string {
    return this.output.join('')
  }

  getLines(): string[] {
    return this.getOutput().split('\n').filter(Boolean)
  }

  clear(): void {
    this.output = []
  }
}

/**
 * Capture stderr
 */
export class StderrCapture {
  private originalWrite: any
  private output: string[] = []

  start(): void {
    this.output = []
    this.originalWrite = process.stderr.write
    process.stderr.write = ((chunk: any, encoding?: any, callback?: any) => {
      this.output.push(chunk.toString())
      return true
    }) as any
  }

  stop(): void {
    process.stderr.write = this.originalWrite
  }

  getOutput(): string {
    return this.output.join('')
  }

  clear(): void {
    this.output = []
  }
}

/**
 * Wait for a condition to be true
 */
export async function waitFor(
  condition: () => boolean | Promise<boolean>,
  timeout = 5000,
  interval = 100
): Promise<void> {
  const start = Date.now()
  while (Date.now() - start < timeout) {
    if (await condition()) {
      return
    }
    await sleep(interval)
  }
  throw new Error('Timeout waiting for condition')
}

/**
 * Sleep for specified milliseconds
 */
export function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * Create mock HTTP response
 */
export function createMockResponse(status: number, data: any): any {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    json: async () => data,
    text: async () => JSON.stringify(data),
    headers: new Map(),
  }
}

/**
 * Mock fetch globally
 */
export function mockFetch(responses: Map<string, any>): void {
  const originalFetch = global.fetch
  global.fetch = async (url: string | URL, options?: any) => {
    const urlStr = url.toString()
    const response = responses.get(urlStr)
    if (!response) {
      throw new Error(`No mock response for URL: ${urlStr}`)
    }
    return response
  } as any
}

/**
 * Restore fetch
 */
export function restoreFetch(): void {
  // Restore if needed
}

/**
 * Create test environment variables
 */
export function withEnv(vars: Record<string, string>, fn: () => void | Promise<void>): any {
  return async () => {
    const original = { ...process.env }
    Object.assign(process.env, vars)
    try {
      await fn()
    } finally {
      process.env = original
    }
  }
}

/**
 * Assert file exists
 */
export async function assertFileExists(filePath: string): Promise<void> {
  const exists = await fs.pathExists(filePath)
  if (!exists) {
    throw new Error(`Expected file to exist: ${filePath}`)
  }
}

/**
 * Assert file contains
 */
export async function assertFileContains(filePath: string, content: string): Promise<void> {
  await assertFileExists(filePath)
  const fileContent = await fs.readFile(filePath, 'utf-8')
  if (!fileContent.includes(content)) {
    throw new Error(`Expected file ${filePath} to contain: ${content}`)
  }
}

/**
 * Create test oclif config
 */
export async function createOclifConfig(root: string): Promise<Config> {
  return Config.load(root)
}

/**
 * Run command programmatically
 */
export async function runCommand(args: string[], config?: Config): Promise<any> {
  const { run } = await import('@oclif/core')
  return run(args, config ? config.root : undefined)
}
