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
    '<rootDir>/test/fixtures/'
  ],
  moduleNameMapper: {
    '^(\\.{1,2}/.*)\\.js$': '$1',
    '^signal-exit$': '<rootDir>/node_modules/signal-exit/dist/cjs/index.js',
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
