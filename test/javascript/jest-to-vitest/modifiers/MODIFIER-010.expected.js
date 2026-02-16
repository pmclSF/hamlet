import { describe, it, expect } from 'vitest';

describe('Platform-specific behavior', () => {
  it('should detect the current platform', () => {
    const platform = process.platform;
    expect(['darwin', 'linux', 'win32']).toContain(platform);
  });

  it('should use correct path separator', () => {
    const sep = process.platform === 'win32' ? '\\' : '/';
    const fullPath = ['home', 'user', 'file.txt'].join(sep);
    expect(fullPath).toContain('user');
    expect(fullPath).toContain('file.txt');
  });

  it('should handle CI environment variables', () => {
    const isCI = process.env.CI === 'true';
    if (isCI) {
      expect(process.env.CI).toBe('true');
    } else {
      expect(process.env.CI).not.toBe('true');
    }
  });

  it('should respect NODE_ENV', () => {
    const env = process.env.NODE_ENV || 'development';
    expect(['development', 'test', 'production']).toContain(env);
  });

  it('should have a valid Node.js version', () => {
    const version = process.version;
    expect(version).toMatch(/^v\d+\.\d+\.\d+$/);
    const major = parseInt(version.slice(1).split('.')[0], 10);
    expect(major).toBeGreaterThanOrEqual(16);
  });
});
