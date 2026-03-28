import { expect } from 'chai'
import { runCommand } from '@oclif/test'
import * as fs from 'fs-extra'
import * as path from 'path'

/**
 * Integration tests for complete CLI workflows
 */
describe('Integration Tests', () => {
  const testDir = path.join(__dirname, 'fixtures', 'integration')

  before(async () => {
    await fs.ensureDir(testDir)
  })

  after(async () => {
    await fs.remove(testDir)
  })

  describe('Complete workflow', () => {
    it('runs full command chain', async () => {
      // Step 1: Initialize
      const initResult = await runCommand(['init', '--dir', testDir])
      expect(initResult).to.have.property('code', 0)

      // Verify initialization
      const configPath = path.join(testDir, '.clirc')
      expect(await fs.pathExists(configPath)).to.be.true

      // Step 2: Configure
      const configResult = await runCommand([
        'config',
        '--set',
        'key=value',
        '--dir',
        testDir,
      ])
      expect(configResult).to.have.property('code', 0)

      // Verify configuration
      const config = await fs.readJson(configPath)
      expect(config).to.have.property('key', 'value')

      // Step 3: Execute main operation
      const execResult = await runCommand(['execute', '--dir', testDir])
      expect(execResult).to.have.property('code', 0)

      // Verify output
      const outputPath = path.join(testDir, 'output.json')
      expect(await fs.pathExists(outputPath)).to.be.true
    })

    it('handles errors gracefully', async () => {
      // Attempt operation without initialization
      try {
        await runCommand(['execute', '--dir', '/nonexistent'])
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error.message).to.include('not initialized')
      }
    })
  })

  describe('Plugin integration', () => {
    it('loads and executes plugin commands', async () => {
      // Install plugin
      const installResult = await runCommand(['plugins:install', '@mycli/plugin-test'])
      expect(installResult).to.have.property('code', 0)

      // Execute plugin command
      const pluginResult = await runCommand(['test:command', '--option', 'value'])
      expect(pluginResult).to.have.property('code', 0)

      // Uninstall plugin
      const uninstallResult = await runCommand(['plugins:uninstall', '@mycli/plugin-test'])
      expect(uninstallResult).to.have.property('code', 0)
    })
  })

  describe('Multi-command workflows', () => {
    it('chains commands with data flow', async () => {
      // Generate data
      const generateResult = await runCommand([
        'generate',
        '--output',
        path.join(testDir, 'data.json'),
      ])
      expect(generateResult).to.have.property('code', 0)

      // Process data
      const processResult = await runCommand([
        'process',
        '--input',
        path.join(testDir, 'data.json'),
        '--output',
        path.join(testDir, 'processed.json'),
      ])
      expect(processResult).to.have.property('code', 0)

      // Validate output
      const validateResult = await runCommand([
        'validate',
        path.join(testDir, 'processed.json'),
      ])
      expect(validateResult).to.have.property('code', 0)
    })
  })

  describe('Environment-specific behavior', () => {
    it('respects environment variables', async () => {
      // Set environment
      process.env.CLI_ENV = 'production'
      process.env.CLI_DEBUG = 'false'

      const result = await runCommand(['status'])
      expect(result).to.have.property('code', 0)

      // Cleanup
      delete process.env.CLI_ENV
      delete process.env.CLI_DEBUG
    })

    it('handles CI environment', async () => {
      // Simulate CI environment
      process.env.CI = 'true'

      // Commands should not prompt in CI
      const result = await runCommand(['deploy', '--auto-confirm'])
      expect(result).to.have.property('code', 0)

      // Cleanup
      delete process.env.CI
    })
  })

  describe('Error recovery', () => {
    it('recovers from partial failures', async () => {
      // Start operation
      const startResult = await runCommand([
        'start-operation',
        '--output',
        path.join(testDir, 'operation.lock'),
      ])
      expect(startResult).to.have.property('code', 0)

      // Simulate failure (lock file exists)
      expect(await fs.pathExists(path.join(testDir, 'operation.lock'))).to.be.true

      // Retry with cleanup
      const retryResult = await runCommand([
        'start-operation',
        '--output',
        path.join(testDir, 'operation.lock'),
        '--force',
      ])
      expect(retryResult).to.have.property('code', 0)
    })
  })

  describe('Performance', () => {
    it('handles large datasets efficiently', async () => {
      const largeFile = path.join(testDir, 'large.json')

      // Generate large dataset
      const largeData = Array.from({ length: 10000 }, (_, i) => ({
        id: i,
        name: `item-${i}`,
        data: 'x'.repeat(100),
      }))
      await fs.writeJson(largeFile, largeData)

      // Process large file
      const startTime = Date.now()
      const result = await runCommand(['process-large', '--input', largeFile])
      const duration = Date.now() - startTime

      expect(result).to.have.property('code', 0)
      expect(duration).to.be.lessThan(30000) // Should complete within 30 seconds
    })
  })

  describe('Concurrent operations', () => {
    it('handles concurrent commands', async () => {
      // Run multiple commands in parallel
      const results = await Promise.all([
        runCommand(['task-1', '--output', path.join(testDir, 'out1.json')]),
        runCommand(['task-2', '--output', path.join(testDir, 'out2.json')]),
        runCommand(['task-3', '--output', path.join(testDir, 'out3.json')]),
      ])

      // All should succeed
      results.forEach(result => {
        expect(result).to.have.property('code', 0)
      })

      // Verify all outputs
      expect(await fs.pathExists(path.join(testDir, 'out1.json'))).to.be.true
      expect(await fs.pathExists(path.join(testDir, 'out2.json'))).to.be.true
      expect(await fs.pathExists(path.join(testDir, 'out3.json'))).to.be.true
    })
  })
})
