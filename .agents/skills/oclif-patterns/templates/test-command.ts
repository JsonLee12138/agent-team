import { expect, test } from '@oclif/test'
import * as fs from 'fs-extra'
import * as path from 'path'

describe('{{COMMAND_NAME}}', () => {
  // Setup and teardown
  const testDir = path.join(__dirname, 'fixtures', 'test-output')

  beforeEach(async () => {
    await fs.ensureDir(testDir)
  })

  afterEach(async () => {
    await fs.remove(testDir)
  })

  // Test basic execution
  test
    .stdout()
    .command(['{{COMMAND_NAME}}', '--help'])
    .it('shows help', ctx => {
      expect(ctx.stdout).to.contain('{{DESCRIPTION}}')
    })

  // Test with required flags
  test
    .stdout()
    .command(['{{COMMAND_NAME}}', '--name', 'World'])
    .it('runs with --name flag', ctx => {
      expect(ctx.stdout).to.contain('Hello, World!')
    })

  // Test flag validation
  test
    .command(['{{COMMAND_NAME}}'])
    .catch(error => {
      expect(error.message).to.contain('Missing required flag')
    })
    .it('fails without required flags')

  // Test with multiple flags
  test
    .stdout()
    .command(['{{COMMAND_NAME}}', '--name', 'Test', '--verbose'])
    .it('runs with verbose flag', ctx => {
      expect(ctx.stdout).to.contain('Verbose mode enabled')
      expect(ctx.stdout).to.contain('Hello, Test!')
    })

  // Test force flag behavior
  test
    .stdout()
    .command(['{{COMMAND_NAME}}', '--name', 'Test', '--force'])
    .it('runs with force flag', ctx => {
      expect(ctx.stdout).to.contain('Force mode')
    })

  // Test error handling
  test
    .command(['{{COMMAND_NAME}}', '--name', ''])
    .catch(error => {
      expect(error.message).to.contain('Invalid')
    })
    .it('handles invalid input')

  // Test exit codes
  test
    .command(['{{COMMAND_NAME}}', '--name', 'Test'])
    .exit(0)
    .it('exits with code 0 on success')

  test
    .command(['{{COMMAND_NAME}}'])
    .exit(2)
    .it('exits with code 2 on missing flags')

  // Test with environment variables
  test
    .env({ CLI_NAME: 'EnvTest' })
    .stdout()
    .command(['{{COMMAND_NAME}}', '--name', 'fromEnv'])
    .it('reads from environment variables', ctx => {
      // Test env-based behavior
    })

  // Test with stdin input
  test
    .stdin('input data\n')
    .stdout()
    .command(['{{COMMAND_NAME}}', '--name', 'Test'])
    .it('handles stdin input', ctx => {
      // Test stdin handling
    })

  // Test file operations
  test
    .do(() => {
      const filePath = path.join(testDir, 'test.txt')
      return fs.writeFile(filePath, 'test content')
    })
    .stdout()
    .command(['{{COMMAND_NAME}}', '--name', 'Test', '--file', path.join(testDir, 'test.txt')])
    .it('processes file input', ctx => {
      expect(ctx.stdout).to.contain('Success')
    })

  // Test async operations
  test
    .stdout()
    .timeout(5000) // Increase timeout for async operations
    .command(['{{COMMAND_NAME}}', '--name', 'Test', '--async'])
    .it('handles async operations', ctx => {
      expect(ctx.stdout).to.contain('completed')
    })

  // Test with mocked dependencies
  test
    .nock('https://api.example.com', api =>
      api.get('/data').reply(200, { result: 'success' })
    )
    .stdout()
    .command(['{{COMMAND_NAME}}', '--name', 'Test', '--api'])
    .it('handles API calls', ctx => {
      expect(ctx.stdout).to.contain('success')
    })

  // Test JSON output
  test
    .stdout()
    .command(['{{COMMAND_NAME}}', '--name', 'Test', '--json'])
    .it('outputs JSON format', ctx => {
      const output = JSON.parse(ctx.stdout)
      expect(output).to.have.property('name', 'Test')
    })
})

// Grouped tests by functionality
describe('{{COMMAND_NAME}} - Flag Parsing', () => {
  test
    .stdout()
    .command(['{{COMMAND_NAME}}', '-n', 'Short'])
    .it('accepts short flags', ctx => {
      expect(ctx.stdout).to.contain('Short')
    })

  test
    .stdout()
    .command(['{{COMMAND_NAME}}', '--name=Inline'])
    .it('accepts inline flag values', ctx => {
      expect(ctx.stdout).to.contain('Inline')
    })
})

describe('{{COMMAND_NAME}} - Error Cases', () => {
  test
    .stderr()
    .command(['{{COMMAND_NAME}}', '--invalid-flag'])
    .catch(error => {
      expect(error.message).to.contain('Unexpected argument')
    })
    .it('handles invalid flags')

  test
    .stderr()
    .command(['{{COMMAND_NAME}}', '--name', 'Test', 'extra-arg'])
    .catch(error => {
      expect(error.message).to.contain('Unexpected argument')
    })
    .it('rejects unexpected arguments')
})
