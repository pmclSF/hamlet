# CLI Documentation

## Commands

### `convert`

Convert Cypress tests to Playwright format.

```bash
hamlet convert <source> [options]
```

#### Arguments

- `source`: Path to Cypress test file or directory

#### Options

- `-o, --output <path>`: Output path for converted tests
- `-c, --config <path>`: Custom configuration file path
- `-t, --test-type <type>`: Specify test type (e2e, component, api, etc.)
- `--validate`: Validate converted tests
- `--report <format>`: Generate conversion report (html, json, markdown)
- `--preserve-structure`: Maintain original directory structure
- `--batch-size <number>`: Number of tests to process in parallel
- `--ignore <pattern>`: Files to ignore (glob pattern)
- `--hooks`: Convert test hooks and configurations
- `--plugins`: Convert Cypress plugins

### Examples

```bash
# Convert a single test file
hamlet convert cypress/e2e/login.cy.js

# Convert all tests in a directory
hamlet convert cypress/e2e --output playwright/tests

# Convert and validate
hamlet convert cypress/e2e --validate --report html

# Convert repository from GitHub
hamlet convert https://github.com/user/repo.git
```

### Exit Codes

| Code | Meaning                  | Example                                                     |
| ---- | ------------------------ | ----------------------------------------------------------- |
| 0    | Success                  | Conversion completed                                        |
| 1    | Runtime error            | Converter failed, unexpected exception                      |
| 2    | Usage / validation error | Invalid framework, file not found, missing flags            |
| 3    | Partial success          | Batch conversion where some files converted and some failed |

### Global Options

| Flag         | Description                                              |
| ------------ | -------------------------------------------------------- |
| `--verbose`  | Show per-file detail and diagnostics                     |
| `--debug`    | Show stack traces on error (also respects `DEBUG=1` env) |
| `--no-color` | Disable colored output (also respects `NO_COLOR=1` env)  |

### Error Output

All errors are prefixed with `Error:` and written to stderr. Common errors include a
`Next steps:` hint. Use `--debug` to see full stack traces.
