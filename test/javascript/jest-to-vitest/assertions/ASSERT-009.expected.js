import { describe, it, expect } from 'vitest';

describe('ConfigLoader', () => {
  it('should have a database section', () => {
    const config = loadConfig('production');
    expect(config).toHaveProperty('database');
  });

  it('should have nested logging level', () => {
    const config = loadConfig('production');
    expect(config).toHaveProperty('logging.level');
  });

  it('should have a server port defined', () => {
    const config = loadConfig('development');
    expect(config).toHaveProperty('server.port');
  });

  it('should have deeply nested retry settings', () => {
    const config = loadConfig('production');
    expect(config).toHaveProperty('database.connection.retry.maxAttempts');
  });
});
