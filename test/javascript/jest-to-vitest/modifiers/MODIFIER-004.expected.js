import { describe, it, expect } from 'vitest';

const isCI = process.env.CI === 'true';
const runIf = (condition) => (condition ? it : it.skip);

describe('Environment-dependent tests', () => {
  runIf(isCI)('should access the staging API on CI', () => {
    const endpoint = 'https://staging.api.example.com';
    expect(endpoint).toContain('staging');
  });

  runIf(!isCI)('should use local mock server in development', () => {
    const endpoint = 'http://localhost:3001';
    expect(endpoint).toContain('localhost');
  });

  it('should always validate input format', () => {
    const input = { type: 'request', payload: {} };
    expect(input.type).toBe('request');
    expect(input.payload).toBeDefined();
  });
});
