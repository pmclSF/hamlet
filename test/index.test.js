import {
  convertCypressToPlaywright,
  convertConfig,
  VERSION,
  SUPPORTED_TEST_TYPES,
  DEFAULT_OPTIONS,
} from '../src/index.js';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';

describe('index.js exports', () => {
  describe('VERSION', () => {
    it('should be a semver string', () => {
      expect(VERSION).toMatch(/^\d+\.\d+\.\d+$/);
    });
  });

  describe('SUPPORTED_TEST_TYPES', () => {
    it('should be an array of test types', () => {
      expect(Array.isArray(SUPPORTED_TEST_TYPES)).toBe(true);
      expect(SUPPORTED_TEST_TYPES).toContain('e2e');
      expect(SUPPORTED_TEST_TYPES).toContain('component');
      expect(SUPPORTED_TEST_TYPES).toContain('api');
    });
  });

  describe('DEFAULT_OPTIONS', () => {
    it('should have expected default values', () => {
      expect(DEFAULT_OPTIONS).toHaveProperty('typescript', false);
      expect(DEFAULT_OPTIONS).toHaveProperty('validate', true);
      expect(DEFAULT_OPTIONS).toHaveProperty('timeout', 30000);
    });
  });
});

describe('convertCypressToPlaywright', () => {
  it('should convert cy.visit calls to page.goto', async () => {
    const input = `describe('Home', () => {
  it('visits home', () => {
    cy.visit('/home');
  });
});`;
    const result = await convertCypressToPlaywright(input);
    expect(result).toContain('goto(\'/home\')');
  });

  it('should convert cy.get to page.locator', async () => {
    const input = `it('finds button', () => {
  cy.get('#btn').click();
});`;
    const result = await convertCypressToPlaywright(input);
    expect(result).toContain('locator(');
  });

  it('should add playwright imports', async () => {
    const input = `it('test', () => { cy.visit('/'); });`;
    const result = await convertCypressToPlaywright(input);
    expect(result).toContain("import { test, expect } from '@playwright/test'");
  });

  it('should extract metadata from content without crashing', async () => {
    const input = `describe('Suite', () => {
  it('case one', () => { cy.visit('/page'); });
});`;
    const result = await convertCypressToPlaywright(input);
    expect(typeof result).toBe('string');
    expect(result.length).toBeGreaterThan(0);
  });

  it('should detect API test type and add request import', async () => {
    const input = `it('api test', () => {
  cy.request('/api/users');
});`;
    const result = await convertCypressToPlaywright(input);
    expect(result).toContain('request');
  });

  it('should handle empty content', async () => {
    const result = await convertCypressToPlaywright('');
    expect(typeof result).toBe('string');
  });
});

describe('convertConfig', () => {
  let tmpDir;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-test-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('should convert a Cypress JSON config to Playwright format', async () => {
    const cypressConfig = {
      baseUrl: 'http://localhost:3000',
      viewportWidth: 1280,
      viewportHeight: 720,
      video: true,
      defaultCommandTimeout: 5000,
    };
    const configPath = path.join(tmpDir, 'cypress.json');
    await fs.writeFile(configPath, JSON.stringify(cypressConfig));

    const result = await convertConfig(configPath);
    expect(result).toContain('defineConfig');
    expect(result).toContain('http://localhost:3000');
    expect(result).toContain('1280');
  });

  it('should include browser projects in config', async () => {
    const configPath = path.join(tmpDir, 'cypress.json');
    await fs.writeFile(configPath, JSON.stringify({ baseUrl: 'http://localhost' }));

    const result = await convertConfig(configPath);
    expect(result).toContain('chromium');
    expect(result).toContain('firefox');
    expect(result).toContain('webkit');
  });

  it('should handle extendedConfig option', async () => {
    const cypressConfig = { baseUrl: 'http://localhost:3000' };
    const configPath = path.join(tmpDir, 'cypress.json');
    await fs.writeFile(configPath, JSON.stringify(cypressConfig));

    const result = await convertConfig(configPath, { extendedConfig: true });
    expect(result).toContain('defineConfig');

    // Should have generated extended config file
    const extendedPath = path.join(tmpDir, 'playwright.extended.config.js');
    const extendedContent = await fs.readFile(extendedPath, 'utf8');
    expect(extendedContent).toContain('defineConfig');
  });

  it('should handle convertPlugins option', async () => {
    const configPath = path.join(tmpDir, 'cypress.json');
    await fs.writeFile(configPath, JSON.stringify({ baseUrl: 'http://localhost' }));

    const result = await convertConfig(configPath, { convertPlugins: true });
    expect(result).toContain('defineConfig');
  });
});
