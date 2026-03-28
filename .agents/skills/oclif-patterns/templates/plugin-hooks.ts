import { Hook } from '@oclif/core'

/**
 * oclif Hook Types:
 * - init: Runs before any command
 * - prerun: Runs before a command's run method
 * - postrun: Runs after a command's run method
 * - command_not_found: Runs when command not found
 */

/**
 * Init hook - runs before any command
 */
export const init: Hook<'init'> = async function (opts) {
  // Access configuration
  const { config } = opts

  // Plugin initialization logic
  this.log('Plugin {{PLUGIN_NAME}} initialized')

  // Example: Check for required environment variables
  if (!process.env.PLUGIN_API_KEY) {
    this.warn('PLUGIN_API_KEY not set - some features may not work')
  }

  // Example: Load plugin configuration
  try {
    // Load config from default locations
  } catch (error) {
    this.debug('Failed to load plugin config')
  }
}

/**
 * Prerun hook - runs before each command
 */
export const prerun: Hook<'prerun'> = async function (opts) {
  const { Command, argv } = opts

  // Log command execution (in debug mode)
  this.debug(`Executing command: ${Command.id}`)

  // Example: Validate environment before running commands
  // Example: Log analytics
  // Example: Check for updates
}

/**
 * Postrun hook - runs after each command
 */
export const postrun: Hook<'postrun'> = async function (opts) {
  const { Command } = opts

  this.debug(`Command completed: ${Command.id}`)

  // Example: Cleanup operations
  // Example: Log analytics
  // Example: Cache results
}

/**
 * Command not found hook - runs when command doesn't exist
 */
export const command_not_found: Hook<'command_not_found'> = async function (opts) {
  const { id } = opts

  this.log(`Command "${id}" not found`)

  // Example: Suggest similar commands
  const suggestions = this.config.commands
    .filter(c => c.id.includes(id) || id.includes(c.id))
    .map(c => c.id)
    .slice(0, 5)

  if (suggestions.length > 0) {
    this.log('\nDid you mean one of these?')
    suggestions.forEach(s => this.log(`  - ${s}`))
  }

  // Example: Check if command is in a plugin that's not installed
  // Example: Suggest installing missing plugin
}

/**
 * Custom plugin-specific hooks
 * Export them and register in package.json oclif.hooks
 */
export const customHook = async function (opts: any) {
  // Custom hook logic
}
