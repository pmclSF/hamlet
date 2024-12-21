# Getting Started with Hamlet

## Installation

```bash
npm install -g hamlet-test-converter
```

## Basic Usage

1. Initialize in your project:
```bash
hamlet init
```

2. Convert your first test:
```bash
hamlet convert cypress/e2e/my-test.cy.js
```

3. Verify the conversion:
```bash
hamlet validate playwright/tests/my-test.spec.js
```

## Converting Different Test Types

### E2E Tests
```bash
hamlet convert cypress/e2e --test-type e2e
```

### Component Tests
```bash
hamlet convert cypress/component --test-type component
```

### API Tests
```bash
hamlet convert cypress/api --test-type api
```

## Common Scenarios

### Converting an Entire Repository
```bash
hamlet convert https://github.com/user/repo.git
```

### Converting with Custom Configuration
```bash
hamlet convert cypress/e2e --config my-config.js
```

### Generating Reports
```bash
hamlet convert cypress/e2e --report html
```

## Next Steps

1. Read about [test types](./test-types.md)
2. Learn about [custom commands](./custom-commands.md)
3. Explore [advanced features](../advanced/repository-conversion.md)
4. Check out [example conversions](../examples/)