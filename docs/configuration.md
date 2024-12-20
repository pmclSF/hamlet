# Hamlet Configuration Guide

## Configuration Files

### .hamletrc.json
The main configuration file for Hamlet.

#### Test Management Section
```json
{
  "testManagement": {
    "type": "string",      // Type of test management system (testrail, azure, xray)
    "config": {
      "url": "string",     // Base URL for the test management system
      "username": "string", // Username for authentication
      "apiKey": "string",  // API key or token for authentication
      "project": "string", // Project name or ID
      "suite": "string"    // Test suite name or ID
    }
  }
}

Conversion Section
jsonCopy{
  "conversion": {
    "patterns": "string",   // Path to patterns.json file
    "output": "string",     // Output directory for converted tests
    "recursive": "boolean", // Process directories recursively
    "createMissing": "boolean", // Create missing directories
    "cleanOutput": "boolean"    // Clean output directory before conversion
  }
}
Reporting Section
jsonCopy{
  "reporting": {
    "format": "string[]",  // Array of report formats (json, html, markdown)
    "outputDir": "string", // Output directory for reports
    "includeStats": "boolean", // Include statistics in reports
    "includeScreenshots": "boolean" // Include screenshots in reports
  }
}
Security Section
jsonCopy{
  "security": {
    "encryptCredentials": "boolean", // Encrypt sensitive credentials
    "auditLog": "boolean",          // Enable audit logging
    "logLevel": "string"            // Logging level (debug, info, warn, error)
  }
}
patterns.json
Defines custom conversion patterns.
jsonCopy{
  "commands": {
    "custom": {
      "pattern": "string",     // Cypress pattern to match
      "playwright": "string",  // Playwright replacement
      "imports": "string[]"    // Required imports
    }
  }
}
test-management.json
Configure test management system mappings.
jsonCopy{
  "mapping": {
    "describe": "string", // Maps describe blocks to test suites
    "it": "string",       // Maps it blocks to test cases
    "tags": {
      "pattern": "string" // Maps tags to test management system fields
    }
  }
}
Environment Variables

HAMLET_CONFIG: Path to config file
HAMLET_TMS_TYPE: Test management system type
HAMLET_TMS_URL: Test management system URL
HAMLET_TMS_TOKEN: Test management system API token
HAMLET_OUTPUT_DIR: Default output directory
HAMLET_LOG_LEVEL: Logging level

CLI Configuration
The CLI can be configured using command line arguments or environment variables:
Copyhamlet convert --config ./config/custom.hamletrc.json
hamlet convert --tms-type testrail --tms-config ./config/testrail.json
Using Configuration Templates

Copy an example configuration:
Copycp examples/configs/testrail-config.json .hamletrc.json

Update with your values:
hamlet init --config .hamletrc.json

Verify configuration:
Copyhamlet verify-config