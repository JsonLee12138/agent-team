import { Command, Flags, ux } from '@oclif/core'

export default class {{COMMAND_NAME}} extends Command {
  static description = 'Async command with parallel operations and error handling'

  static examples = [
    '<%= config.bin %> <%= command.id %> --urls https://api1.com,https://api2.com',
    '<%= config.bin %> <%= command.id %> --urls https://api.com --retry 3 --timeout 5000',
  ]

  static flags = {
    urls: Flags.string({
      char: 'u',
      description: 'Comma-separated list of URLs to fetch',
      required: true,
    }),
    parallel: Flags.integer({
      char: 'p',
      description: 'Maximum parallel operations',
      default: 5,
    }),
    timeout: Flags.integer({
      char: 't',
      description: 'Request timeout in milliseconds',
      default: 10000,
    }),
    retry: Flags.integer({
      char: 'r',
      description: 'Number of retry attempts',
      default: 0,
    }),
    verbose: Flags.boolean({
      char: 'v',
      description: 'Verbose output',
      default: false,
    }),
  }

  async run(): Promise<void> {
    const { flags } = await this.parse({{COMMAND_NAME}})

    const urls = flags.urls.split(',').map(u => u.trim())

    if (flags.verbose) {
      this.log(`Processing ${urls.length} URLs with max ${flags.parallel} parallel operations`)
    }

    // Process with progress bar
    ux.action.start(`Fetching ${urls.length} URLs`)

    try {
      const results = await this.fetchWithConcurrency(urls, flags)
      ux.action.stop('done')

      // Display results table
      this.displayResults(results, flags.verbose)
    } catch (error) {
      ux.action.stop('failed')
      this.error(`Operation failed: ${error instanceof Error ? error.message : 'Unknown error'}`, {
        exit: 1,
      })
    }
  }

  private async fetchWithConcurrency(
    urls: string[],
    flags: any
  ): Promise<Array<{ url: string; status: string; data?: any; error?: string }>> {
    const results: Array<{ url: string; status: string; data?: any; error?: string }> = []
    const chunks = this.chunkArray(urls, flags.parallel)

    for (const chunk of chunks) {
      const promises = chunk.map(url => this.fetchWithRetry(url, flags))
      const chunkResults = await Promise.allSettled(promises)

      chunkResults.forEach((result, index) => {
        if (result.status === 'fulfilled') {
          results.push({
            url: chunk[index],
            status: 'success',
            data: result.value,
          })
        } else {
          results.push({
            url: chunk[index],
            status: 'failed',
            error: result.reason.message,
          })
        }
      })
    }

    return results
  }

  private async fetchWithRetry(url: string, flags: any): Promise<any> {
    let lastError: Error | null = null

    for (let attempt = 0; attempt <= flags.retry; attempt++) {
      try {
        if (flags.verbose && attempt > 0) {
          this.log(`Retry ${attempt}/${flags.retry} for ${url}`)
        }

        const response = await this.fetchWithTimeout(url, flags.timeout)
        return response
      } catch (error) {
        lastError = error instanceof Error ? error : new Error('Unknown error')
        if (attempt < flags.retry) {
          // Exponential backoff
          await this.sleep(Math.pow(2, attempt) * 1000)
        }
      }
    }

    throw lastError
  }

  private async fetchWithTimeout(url: string, timeout: number): Promise<any> {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), timeout)

    try {
      const response = await fetch(url, { signal: controller.signal })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()
      return data
    } finally {
      clearTimeout(timeoutId)
    }
  }

  private displayResults(
    results: Array<{ url: string; status: string; data?: any; error?: string }>,
    verbose: boolean
  ): void {
    const successCount = results.filter(r => r.status === 'success').length
    const failCount = results.filter(r => r.status === 'failed').length

    this.log(`\nResults: ${successCount} successful, ${failCount} failed\n`)

    if (verbose) {
      ux.table(results, {
        url: {
          header: 'URL',
        },
        status: {
          header: 'Status',
        },
        error: {
          header: 'Error',
          get: row => row.error || '-',
        },
      })
    } else {
      // Show only failures
      const failures = results.filter(r => r.status === 'failed')
      if (failures.length > 0) {
        this.log('Failed URLs:')
        failures.forEach(f => this.log(`  × ${f.url}: ${f.error}`))
      }
    }
  }

  private chunkArray<T>(array: T[], size: number): T[][] {
    const chunks: T[][] = []
    for (let i = 0; i < array.length; i += size) {
      chunks.push(array.slice(i, i + size))
    }
    return chunks
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }
}
