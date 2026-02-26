/** @type {import('jest').Config} */
// NOTE: This project requires NODE_OPTIONS='--experimental-vm-modules' for
// Jest ESM support.  That flag leaves internal IPC handles in Jest workers
// which can emit a "worker process has failed to exit gracefully" warning on
// shutdown.  This is a known Jest 29 limitation, not a test-level leak.
// See docs/adr/004-jest-esm-strategy.md for details.
export default {
  testEnvironment: 'node',
  testTimeout: 30000,
  transform: {},
  extensionsToTreatAsEsm: [],
  testMatch: [
    '<rootDir>/test/**/*.spec.js',
    '<rootDir>/test/**/*.test.js'
  ],
  modulePathIgnorePatterns: [
    '<rootDir>/benchmarks/',
    '<rootDir>/testing/',
  ],
  testPathIgnorePatterns: [
    '/node_modules/',
    '<rootDir>/test/output/',
    '<rootDir>/test/fixtures/',
    // Skip test/index.test.js in CI — Jest's experimental ESM VM modules on
    // Node 18-20 trigger a cjs-module-lexer bug with signal-exit that causes
    // "Export '__signal_exit_emitter__' is not defined in module"
    ...(process.env.CI ? ['<rootDir>/test/index.test.js'] : []),
  ],
  moduleNameMapper: {
    '^(\\.{1,2}/.*)\\.js$': '$1',
  },
  collectCoverageFrom: [
    'src/**/*.js',
    '!src/**/*.d.ts',
    '!src/types/**/*'
  ],
  coverageThreshold: {
    global: {
      branches: 50,
      functions: 50,
      lines: 50,
      statements: 50
    }
  }
};
