# Basic CLI Example

Complete example of a simple oclif CLI with multiple commands.

## Project Structure

```
mycli/
├── package.json
├── tsconfig.json
├── src/
│   ├── commands/
│   │   ├── hello.ts
│   │   ├── goodbye.ts
│   │   └── config.ts
│   └── index.ts
├── test/
│   └── commands/
│       └── hello.test.ts
└── bin/
    └── run.js
```

## Step-by-Step Setup

### 1. Initialize Project

```bash
mkdir mycli && cd mycli
npm init -y
npm install @oclif/core
npm install --save-dev @oclif/test @types/node typescript ts-node oclif
```

### 2. Create package.json Configuration

Add oclif configuration:

```json
{
  "name": "mycli",
  "version": "1.0.0",
  "oclif": {
    "bin": "mycli",
    "commands": "./lib/commands",
    "plugins": [
      "@oclif/plugin-help"
    ]
  }
}
```

### 3. Create Hello Command

File: `src/commands/hello.ts`

```typescript
import { Command, Flags, Args } from '@oclif/core'

export default class Hello extends Command {
  static description = 'Say hello to someone'

  static examples = [
    '<%= config.bin %> <%= command.id %> Alice',
    '<%= config.bin %> <%= command.id %> Bob --greeting="Hi"',
  ]

  static flags = {
    greeting: Flags.string({
      char: 'g',
      description: 'Greeting to use',
      default: 'Hello',
    }),
    excited: Flags.boolean({
      char: 'e',
      description: 'Add exclamation',
      default: false,
    }),
  }

  static args = {
    name: Args.string({
      description: 'Name to greet',
      required: true,
    }),
  }

  async run(): Promise<void> {
    const { args, flags } = await this.parse(Hello)

    const punctuation = flags.excited ? '!' : '.'
    this.log(`${flags.greeting}, ${args.name}${punctuation}`)
  }
}
```

### 4. Build and Test

```bash
npm run build
./bin/run.js hello World
# Output: Hello, World.

./bin/run.js hello World --greeting="Hi" --excited
# Output: Hi, World!
```

### 5. Add Help Documentation

```bash
./bin/run.js hello --help

# Output:
# Say hello to someone
#
# USAGE
#   $ mycli hello NAME [-g <value>] [-e]
#
# ARGUMENTS
#   NAME  Name to greet
#
# FLAGS
#   -e, --excited          Add exclamation
#   -g, --greeting=<value> [default: Hello] Greeting to use
```

## Testing

File: `test/commands/hello.test.ts`

```typescript
import { expect, test } from '@oclif/test'

describe('hello', () => {
  test
    .stdout()
    .command(['hello', 'World'])
    .it('says hello', ctx => {
      expect(ctx.stdout).to.contain('Hello, World.')
    })

  test
    .stdout()
    .command(['hello', 'Alice', '--excited'])
    .it('says hello with excitement', ctx => {
      expect(ctx.stdout).to.contain('Hello, Alice!')
    })

  test
    .stdout()
    .command(['hello', 'Bob', '--greeting=Hi'])
    .it('uses custom greeting', ctx => {
      expect(ctx.stdout).to.contain('Hi, Bob.')
    })
})
```

## Run Tests

```bash
npm test
```

## Distribution

### Package for npm

```bash
npm pack
npm publish
```

### Install Globally

```bash
npm install -g .
mycli hello World
```

## Key Concepts Demonstrated

1. **Command Structure**: Basic command with flags and args
2. **Flag Types**: String and boolean flags with defaults
3. **Arguments**: Required string argument
4. **Help Documentation**: Auto-generated from metadata
5. **Testing**: Using @oclif/test for command testing
6. **Build Process**: TypeScript compilation to lib/
7. **CLI Binary**: bin/run.js entry point
