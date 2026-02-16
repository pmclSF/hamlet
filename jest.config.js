/** @type {import('jest').Config} */
export default {
  testEnvironment: 'node',
  testTimeout: 30000,
  transform: {},
  extensionsToTreatAsEsm: [],
  testMatch: [
    '<rootDir>/test/**/*.spec.js',
    '<rootDir>/test/**/*.test.js'
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
