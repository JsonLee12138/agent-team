import { expect } from 'chai'
import * as path from 'path'

/**
 * Global test setup for oclif commands
 *
 * This file is loaded before all tests
 */

// Extend chai with custom assertions if needed
expect.extend = function (assertions: any) {
  Object.assign(expect, assertions)
}

// Set test environment
process.env.NODE_ENV = 'test'
process.env.CLI_TEST = 'true'

// Disable colors in test output
process.env.FORCE_COLOR = '0'

// Set test timeout
const DEFAULT_TIMEOUT = 10000
if (typeof (global as any).setTimeout !== 'undefined') {
  ;(global as any).setTimeout(DEFAULT_TIMEOUT)
}

// Setup global test fixtures directory
export const FIXTURES_DIR = path.join(__dirname, 'fixtures')

// Mock console methods if needed
export function mockConsole() {
  const originalLog = console.log
  const originalError = console.error
  const originalWarn = console.warn

  const logs: string[] = []
  const errors: string[] = []
  const warns: string[] = []

  console.log = (...args: any[]) => {
    logs.push(args.join(' '))
  }

  console.error = (...args: any[]) => {
    errors.push(args.join(' '))
  }

  console.warn = (...args: any[]) => {
    warns.push(args.join(' '))
  }

  return {
    logs,
    errors,
    warns,
    restore: () => {
      console.log = originalLog
      console.error = originalError
      console.warn = originalWarn
    },
  }
}

// Global before hook
before(async () => {
  // Setup test database, services, etc.
})

// Global after hook
after(async () => {
  // Cleanup test resources
})

// Global beforeEach hook
beforeEach(() => {
  // Reset state before each test
})

// Global afterEach hook
afterEach(() => {
  // Cleanup after each test
})

/**
 * Custom matchers for oclif tests
 */
export const customMatchers = {
  /**
   * Check if output contains text
   */
  toContainOutput(received: string, expected: string): boolean {
    return received.includes(expected)
  },

  /**
   * Check if command succeeded
   */
  toSucceed(received: { code: number }): boolean {
    return received.code === 0
  },

  /**
   * Check if command failed with specific code
   */
  toFailWith(received: { code: number }, expectedCode: number): boolean {
    return received.code === expectedCode
  },
}
