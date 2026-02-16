import { ConfigConverter } from '../../src/core/ConfigConverter.js';

describe('ConfigConverter', () => {
  let converter;

  beforeEach(() => {
    converter = new ConfigConverter();
  });

  describe('Jest → Vitest', () => {
    it('should convert testEnvironment', () => {
      const input = `module.exports = { testEnvironment: 'jsdom' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('vitest/config');
      expect(result).toContain("environment: 'jsdom'");
    });

    it('should convert testEnvironment node', () => {
      const input = `module.exports = { testEnvironment: 'node' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain("environment: 'node'");
    });

    it('should convert setupFiles', () => {
      const input = `module.exports = { setupFiles: './setup.js' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('setupFiles');
    });

    it('should convert testTimeout', () => {
      const input = `module.exports = { testTimeout: 30000 };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('testTimeout');
      expect(result).toContain('30000');
    });

    it('should convert clearMocks', () => {
      const input = `module.exports = { clearMocks: true };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('clearMocks');
      expect(result).toContain('true');
    });

    it('should add HAMLET-TODO for unrecognized keys', () => {
      const input = `module.exports = { testEnvironment: 'node', moduleNameMapper: './mappers' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('moduleNameMapper');
    });

    it('should handle export default syntax', () => {
      const input = `export default { testEnvironment: 'node' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain("environment: 'node'");
    });

    it('should handle empty config', () => {
      const input = `module.exports = {};`;
      const result = converter.convert(input, 'jest', 'vitest');

      // Should still produce valid vitest config structure
      expect(result).toContain('defineConfig');
    });
  });

  describe('Cypress → Playwright', () => {
    it('should convert baseUrl', () => {
      const input = `module.exports = { baseUrl: 'http://localhost:3000' };`;
      const result = converter.convert(input, 'cypress', 'playwright');

      expect(result).toContain('@playwright/test');
      expect(result).toContain('baseURL');
    });

    it('should convert viewportWidth and viewportHeight', () => {
      const input = `module.exports = { viewportWidth: 1280, viewportHeight: 720 };`;
      const result = converter.convert(input, 'cypress', 'playwright');

      expect(result).toContain('viewport');
      expect(result).toContain('1280');
    });

    it('should convert retries', () => {
      const input = `module.exports = { retries: 2 };`;
      const result = converter.convert(input, 'cypress', 'playwright');

      expect(result).toContain('retries');
      expect(result).toContain('2');
    });

    it('should add HAMLET-TODO for unrecognized Cypress keys', () => {
      const input = `module.exports = { baseUrl: 'http://localhost', chromeWebSecurity: false };`;
      const result = converter.convert(input, 'cypress', 'playwright');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('chromeWebSecurity');
    });
  });

  describe('edge cases', () => {
    it('should handle config with JS logic (conditional) by adding HAMLET-TODO', () => {
      const input = `const env = process.env.NODE_ENV;\nmodule.exports = env === 'ci' ? { retries: 3 } : { retries: 0 };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('HAMLET-TODO');
    });

    it('should handle nested config (projects array) with HAMLET-TODO', () => {
      const input = `module.exports = { projects: [{ displayName: 'unit' }, { displayName: 'e2e' }] };`;
      const result = converter.convert(input, 'jest', 'vitest');

      // projects is unsupported → should have HAMLET-TODO
      expect(result).toContain('HAMLET-TODO');
    });

    it('should handle unsupported conversion direction', () => {
      const input = `module.exports = { baseUrl: '/' };`;
      const result = converter.convert(input, 'selenium', 'playwright');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('Manual action required');
    });
  });
});
