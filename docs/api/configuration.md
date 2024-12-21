# Configuration

Hamlet can be configured through a configuration file (`hamlet.config.js`) or command-line options.

## Configuration File

Create a `hamlet.config.js` in your project root:

```javascript
module.exports = {
  // Output configuration
  output: {
    directory: './playwright-tests',
    preserveStructure: true,
    createMissingDirs: true
  },

  // Conversion options
  conversion: {
    typescript: true,
    addComments: true,
    preserveDescriptions: true
  },

  // Test options
  test: {
    validateAfterConversion: true,
    generateMigrationGuide: true,
    compareScreenshots: false
  },

  // Test management integration
  testManagement: {
    type: 'azure', // or 'testrail', 'xray'
    enabled: false,
    config: {
      // Test management specific configuration
    }
  },

  // Reporting options
  reporting: {
    format: ['html', 'json'],
    outputDir: './reports',
    includeScreenshots: true
  }
}
```

## Environment Variables

You can also use environment variables for configuration:

- `HAMLET_OUTPUT_DIR`: Output directory for converted tests
- `HAMLET_TYPESCRIPT`: Enable TypeScript conversion
- `HAMLET_VALIDATE`: Enable validation after conversion
- `HAMLET_REPORT_FORMAT`: Report format (html, json, markdown)

## Test Management Configuration

### Azure DevOps
```javascript
testManagement: {
  type: 'azure',
  config: {
    organization: 'your-org',
    project: 'your-project',
    pat: process.env.AZURE_PAT
  }
}
```

### TestRail
```javascript
testManagement: {
  type: 'testrail',
  config: {
    host: 'your-instance.testrail.com',
    username: 'username',
    apiKey: process.env.TESTRAIL_API_KEY
  }
}
```

## Advanced Configuration

### Custom Patterns
```javascript
patterns: {
  commands: {
    'custom-command': 'playwright-command'
  },
  assertions: {
    'custom-assertion': 'playwright-assertion'
  }
}
```

### Plugin Configuration
```javascript
plugins: {
  enabled: true,
  convertCustomCommands: true,
  includeComments: true
}
```