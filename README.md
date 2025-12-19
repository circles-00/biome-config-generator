# Biome Config Generator

A CLI tool that automatically migrates ESLint and Prettier configurations to [Biome](https://biomejs.dev/).

## Features

- Recursively scans directories for ESLint and Prettier config files
- Automatically runs Biome's migration commands for each found configuration
- Supports dry-run mode to preview changes before applying them
- Skips common non-source directories (`node_modules`, `.git`, `dist`, `build`, `.devops`)
- Patches generated `biome.json` with useful defaults:
  - `formatWithErrors: true` - format files even if they have errors
  - `unsafeParameterDecoratorsEnabled: true` - enable TypeScript parameter decorators

## Supported Config Files

### ESLint
- `.eslintrc.json`, `.eslintrc.js`, `.eslintrc.cjs`
- `.eslintrc.yaml`, `.eslintrc.yml`, `.eslintrc`
- `eslint.config.js`, `eslint.config.mjs`, `eslint.config.cjs`

### Prettier
- `.prettierrc`, `.prettierrc.json`, `.prettierrc.yml`, `.prettierrc.yaml`
- `.prettierrc.json5`, `.prettierrc.js`, `.prettierrc.cjs`, `.prettierrc.mjs`
- `.prettierrc.toml`, `prettier.config.js`, `prettier.config.cjs`, `prettier.config.mjs`

## Installation

```bash
go install github.com/circles-00/biome-config-generator@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/biome-config-generator.git
cd biome-config-generator
go build -o biome_configurator
```

## Requirements

- Go 1.24+
- Node.js with `npx` available in PATH
- `@biomejs/biome` (automatically fetched via npx)

## Usage

```bash
biome_configurator -input <directory> [-dry-run]
```

### Options

| Flag | Description |
|------|-------------|
| `-input` | Input directory to scan for ESLint/Prettier configs (required) |
| `-dry-run` | Only show what would be done without actually doing it |

### Examples

Migrate all configs in a project:

```bash
biome_configurator -input ./my-project
```

Preview changes without applying them:

```bash
biome_configurator -input ./my-project -dry-run
```

Migrate configs in current directory:

```bash
biome_configurator -input .
```

## Post-Migration

After running the migration, you may want to add `biome.json` to your global gitignore if you don't want to commit the generated configs:

```bash
echo 'biome.json' >> ~/.gitignore_global
git config --global core.excludesfile ~/.gitignore_global
```

## License

MIT
