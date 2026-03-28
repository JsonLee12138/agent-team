import { Command, Flags, Args, ux } from '@oclif/core'
import * as fs from 'fs-extra'
import * as path from 'path'

export default class {{COMMAND_NAME}} extends Command {
  static description = '{{DESCRIPTION}}'

  static examples = [
    '<%= config.bin %> <%= command.id %> myfile.txt --output result.json',
    '<%= config.bin %> <%= command.id %> data.csv --format json --validate',
  ]

  static flags = {
    output: Flags.string({
      char: 'o',
      description: 'Output file path',
      required: true,
    }),
    format: Flags.string({
      char: 'f',
      description: 'Output format',
      options: ['json', 'yaml', 'csv'],
      default: 'json',
    }),
    validate: Flags.boolean({
      description: 'Validate input before processing',
      default: false,
    }),
    force: Flags.boolean({
      description: 'Overwrite existing output file',
      default: false,
    }),
    verbose: Flags.boolean({
      char: 'v',
      description: 'Verbose output',
      default: false,
    }),
  }

  static args = {
    file: Args.string({
      description: 'Input file to process',
      required: true,
    }),
  }

  async run(): Promise<void> {
    const { args, flags } = await this.parse({{COMMAND_NAME}})

    // Validation
    await this.validateInput(args.file, flags)

    // Processing with spinner
    ux.action.start('Processing file')

    try {
      const result = await this.processFile(args.file, flags)
      ux.action.stop('done')

      // Output
      await this.writeOutput(result, flags.output, flags.format, flags.force)

      this.log(`✓ Successfully processed ${args.file}`)
      this.log(`✓ Output written to ${flags.output}`)
    } catch (error) {
      ux.action.stop('failed')
      this.error(`Processing failed: ${error instanceof Error ? error.message : 'Unknown error'}`, {
        exit: 1,
      })
    }
  }

  private async validateInput(file: string, flags: any): Promise<void> {
    // Check file exists
    if (!await fs.pathExists(file)) {
      this.error(`File not found: ${file}`, { exit: 1 })
    }

    // Check output directory exists
    const outputDir = path.dirname(flags.output)
    if (!await fs.pathExists(outputDir)) {
      this.error(`Output directory not found: ${outputDir}`, { exit: 1 })
    }

    // Check output file doesn't exist (unless force)
    if (!flags.force && await fs.pathExists(flags.output)) {
      const overwrite = await ux.confirm(`Output file ${flags.output} exists. Overwrite? (y/n)`)
      if (!overwrite) {
        this.error('Operation cancelled', { exit: 0 })
      }
    }

    if (flags.validate) {
      if (flags.verbose) {
        this.log('Running validation...')
      }
      // Perform custom validation here
    }
  }

  private async processFile(file: string, flags: any): Promise<any> {
    // Read input file
    const content = await fs.readFile(file, 'utf-8')

    if (flags.verbose) {
      this.log(`Read ${content.length} bytes from ${file}`)
    }

    // Process content (placeholder - implement your logic)
    const result = {
      source: file,
      format: flags.format,
      timestamp: new Date().toISOString(),
      data: content,
    }

    return result
  }

  private async writeOutput(
    result: any,
    output: string,
    format: string,
    force: boolean
  ): Promise<void> {
    let content: string

    switch (format) {
      case 'json':
        content = JSON.stringify(result, null, 2)
        break
      case 'yaml':
        // Implement YAML formatting
        content = JSON.stringify(result, null, 2) // Placeholder
        break
      case 'csv':
        // Implement CSV formatting
        content = JSON.stringify(result, null, 2) // Placeholder
        break
      default:
        content = JSON.stringify(result, null, 2)
    }

    await fs.writeFile(output, content, 'utf-8')
  }
}
