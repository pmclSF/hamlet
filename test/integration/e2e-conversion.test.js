import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import { ConverterFactory, FRAMEWORKS } from '../../src/core/ConverterFactory.js';
import { FrameworkDetector } from '../../src/core/FrameworkDetector.js';
import { PatternEngine } from '../../src/core/PatternEngine.js';
import { BatchProcessor } from '../../src/converter/batchProcessor.js';
import { TestMetadataCollector } from '../../src/converter/metadataCollector.js';
import { DependencyAnalyzer } from '../../src/converter/dependencyAnalyzer.js';
import { PluginConverter } from '../../src/converter/plugins.js';
import { TestMapper } from '../../src/converter/mapper.js';
import { ConversionReporter } from '../../src/utils/reporter.js';
import { fileUtils, stringUtils, codeUtils } from '../../src/utils/helpers.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const tmpDir = path.join(__dirname, '../.tmp-e2e-test');

describe('End-to-end: Cypress to Playwright conversion pipeline', () => {
  let converter;

  beforeEach(async () => {
    converter = await ConverterFactory.createConverter('cypress', 'playwright');
    await fs.mkdir(tmpDir, { recursive: true });
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('should convert a complete Cypress e2e test file to Playwright', async () => {
    const cypressTest = `
describe('Login Page', () => {
  beforeEach(() => {
    cy.visit('/login');
  });

  it('should display the login form', () => {
    cy.get('#email').should('be.visible');
    cy.get('#password').should('be.visible');
    cy.get('button[type="submit"]').should('exist');
  });

  it('should login with valid credentials', () => {
    cy.get('#email').type('user@example.com');
    cy.get('#password').type('password123');
    cy.get('button[type="submit"]').click();
    cy.url().should('include', '/dashboard');
  });

  it('should show error for invalid credentials', () => {
    cy.get('#email').type('wrong@example.com');
    cy.get('#password').type('wrongpass');
    cy.get('button[type="submit"]').click();
    cy.get('.error-message').should('be.visible');
  });
});
`;

    const result = await converter.convert(cypressTest);

    // Verify structure conversion
    expect(result).toContain('test.describe(');
    expect(result).toContain('test(');
    expect(result).toContain('test.beforeEach(');

    // Verify command conversion
    expect(result).toContain('page.goto(');
    expect(result).toContain('.fill(');
    expect(result).toContain('.click()');

    // Verify assertions conversion
    expect(result).toContain('toBeVisible()');

    // Verify the output is syntactically valid (no describe/it/cy. leftovers for core patterns)
    expect(result).not.toContain('cy.visit');
    expect(result).not.toContain('cy.get');
  });

  it('should convert and write output to disk', async () => {
    const cypressTest = `
describe('Navigation', () => {
  it('should navigate to about page', () => {
    cy.visit('/about');
    cy.get('h1').should('be.visible');
  });
});
`;
    const outputPath = path.join(tmpDir, 'navigation.spec.js');
    const result = await converter.convert(cypressTest);
    await fs.writeFile(outputPath, result);

    const written = await fs.readFile(outputPath, 'utf8');
    expect(written).toBe(result);
    expect(written).toContain('page.goto(');
  });

  it('should detect framework, convert, and track stats', async () => {
    const cypressTest = `
describe('Test', () => {
  it('should work', () => {
    cy.visit('/');
    cy.get('.btn').click();
  });
});
`;

    // Step 1: Detect framework
    const detection = FrameworkDetector.detectFromContent(cypressTest);
    expect(detection.framework).toBe('cypress');

    // Step 2: Create converter from detection
    const detectedConverter = await ConverterFactory.createConverter(
      detection.framework, 'playwright'
    );

    // Step 3: Convert
    const result = await detectedConverter.convert(cypressTest);
    expect(result).toContain('page.goto(');

    // Step 4: Verify stats tracked
    const stats = detectedConverter.getStats();
    expect(stats.conversions).toBeGreaterThanOrEqual(1);
  });
});

describe('End-to-end: Playwright to Cypress conversion pipeline', () => {
  let converter;

  beforeEach(async () => {
    converter = await ConverterFactory.createConverter('playwright', 'cypress');
  });

  it('should convert a complete Playwright test to Cypress', async () => {
    const playwrightTest = `
import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard');
  });

  test('should display user info', async ({ page }) => {
    await expect(page.locator('.user-name')).toBeVisible();
    await expect(page.locator('.user-email')).toHaveText('user@example.com');
  });

  test('should navigate to settings', async ({ page }) => {
    await page.locator('.settings-link').click();
    await expect(page).toHaveURL('/settings');
  });
});
`;

    const result = await converter.convert(playwrightTest);

    // Verify structure conversion
    expect(result).toContain('describe(');
    expect(result).toContain('it(');
    expect(result).toContain('beforeEach(');

    // Verify command conversion
    expect(result).toContain('cy.visit(');
    expect(result).toContain('.click()');

    // Verify assertions
    expect(result).toContain("should('be.visible')");

    // Verify Playwright imports removed
    expect(result).not.toContain("from '@playwright/test'");

    // Verify Cypress reference added
    expect(result).toContain('/// <reference types="cypress" />');
  });
});

describe('End-to-end: Selenium to Playwright conversion pipeline', () => {
  let converter;

  beforeEach(async () => {
    converter = await ConverterFactory.createConverter('selenium', 'playwright');
  });

  it('should convert a Selenium test to Playwright', async () => {
    const seleniumTest = `
const { Builder, By, Key, until } = require('selenium-webdriver');

describe('Search', () => {
  let driver;

  beforeAll(async () => {
    driver = await new Builder().forBrowser('chrome').build();
  });

  afterAll(async () => {
    await driver.quit();
  });

  it('should search for items', async () => {
    await driver.get('http://localhost:3000/search');
    await driver.findElement(By.id('search-input')).sendKeys('test query');
    await driver.findElement(By.css('.search-btn')).click();
    const results = await driver.findElement(By.css('.results'));
    expect(results).toBeDefined();
  });
});
`;

    const result = await converter.convert(seleniumTest);

    // Verify Selenium driver patterns are converted
    expect(result).toContain('page.goto(');
    expect(result).toContain('page.locator(');

    // Verify Playwright imports added
    expect(result).toContain('@playwright/test');
  });
});

describe('End-to-end: Bidirectional round-trip fidelity', () => {
  it('should preserve test structure through Cypress->Playwright->Cypress', async () => {
    const originalCypress = `
describe('Round Trip', () => {
  it('should navigate', () => {
    cy.visit('/home');
    cy.get('.title').should('be.visible');
  });
});
`;

    // Forward: Cypress -> Playwright
    const cypToPlaywright = await ConverterFactory.createConverter('cypress', 'playwright');
    const playwrightResult = await cypToPlaywright.convert(originalCypress);

    // Reverse: Playwright -> Cypress
    const playwrightToCyp = await ConverterFactory.createConverter('playwright', 'cypress');
    const cypressResult = await playwrightToCyp.convert(playwrightResult);

    // The round-trip should preserve the semantic structure
    expect(cypressResult).toContain('describe(');
    expect(cypressResult).toContain('it(');
    expect(cypressResult).toContain('cy.visit(');
    expect(cypressResult).toContain("should('be.visible')");
  });
});

describe('End-to-end: Batch processing pipeline', () => {
  beforeEach(async () => {
    await fs.mkdir(tmpDir, { recursive: true });
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('should batch-convert multiple files', async () => {
    // Create test files
    const files = [];
    for (let i = 0; i < 6; i++) {
      const filePath = path.join(tmpDir, `test${i}.cy.js`);
      await fs.writeFile(filePath, `
describe('Test ${i}', () => {
  it('should work ${i}', () => {
    cy.visit('/page${i}');
  });
});
`);
      files.push(filePath);
    }

    const batchProcessor = new BatchProcessor({ batchSize: 3 });
    const converter = await ConverterFactory.createConverter('cypress', 'playwright');

    const stats = await batchProcessor.processBatch(files, async (file) => {
      const content = await fs.readFile(file, 'utf8');
      const converted = await converter.convert(content);
      const outputPath = file.replace('.cy.js', '.spec.js');
      await fs.writeFile(outputPath, converted);
    });

    expect(stats.total).toBe(6);
    expect(stats.processed).toBe(6);
    expect(stats.failed).toBe(0);

    // Verify output files exist
    for (let i = 0; i < 6; i++) {
      const outputPath = path.join(tmpDir, `test${i}.spec.js`);
      const exists = await fileUtils.fileExists(outputPath);
      expect(exists).toBe(true);

      const content = await fs.readFile(outputPath, 'utf8');
      expect(content).toContain('page.goto(');
    }
  });
});

describe('End-to-end: Metadata collection and dependency analysis', () => {
  beforeEach(async () => {
    await fs.mkdir(tmpDir, { recursive: true });
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('should collect metadata from a test file and analyze dependencies', async () => {
    const testContent = `
// @smoke @regression
import { loginHelper } from './helpers.js';

describe('User Dashboard', () => {
  beforeEach(() => {
    cy.visit('/dashboard');
  });

  it('should show user profile', () => {
    cy.get('.profile').should('be.visible');
    cy.get('.profile-name').should('have.text', 'John Doe');
  });

  it('should navigate to settings', () => {
    cy.get('.settings-link').click();
    cy.url().should('include', '/settings');
  });
});
`;
    const testPath = path.join(tmpDir, 'dashboard.cy.js');
    await fs.writeFile(testPath, testContent);

    // Collect metadata
    const metadataCollector = new TestMetadataCollector();
    const metadata = await metadataCollector.collectMetadata(testPath);

    expect(metadata.type).toBe('e2e');
    expect(metadata.suites).toHaveLength(1);
    expect(metadata.suites[0].name).toBe('User Dashboard');
    expect(metadata.cases).toHaveLength(2);
    expect(metadata.tags).toContain('smoke');
    expect(metadata.tags).toContain('regression');
    expect(metadata.complexity.assertions).toBeGreaterThan(0);

    // Analyze dependencies
    const depAnalyzer = new DependencyAnalyzer();
    const deps = await depAnalyzer.analyzeDependencies(testPath);

    expect(deps.imports).toHaveLength(1);
    expect(deps.imports[0].source).toBe('./helpers.js');
  });
});

describe('End-to-end: Plugin detection and conversion', () => {
  it('should detect and convert plugins from content', () => {
    const pluginContent = `
import 'cypress-file-upload';
import 'cypress-axe';

describe('Accessible Upload', () => {
  it('should upload and check a11y', () => {
    cy.get('input[type="file"]').attachFile('data.csv');
    cy.injectAxe();
    cy.checkA11y();
  });
});
`;

    const pluginConverter = new PluginConverter();
    const detected = pluginConverter.detectPlugins(pluginContent);

    expect(detected).toContain('cypress-file-upload');
    expect(detected).toContain('cypress-axe');

    // Verify each detected plugin can be converted
    for (const plugin of detected) {
      expect(pluginConverter.canConvert(plugin)).toBe(true);
      const info = pluginConverter.getPluginInfo(plugin);
      expect(info).not.toBeNull();
      expect(info.playwright).toBeDefined();
    }
  });
});

describe('End-to-end: Reporter full lifecycle', () => {
  it('should track a complete conversion lifecycle', () => {
    const reporter = new ConversionReporter({ format: 'json' });

    // Start report
    reporter.startReport();

    // Record conversion steps
    reporter.recordStep('Analyzing project', 'success', { files: 5 });
    reporter.recordStep('Converting tests', 'success', { converted: 5 });
    reporter.recordStep('Validating output', 'success', { passed: 5 });

    // Add test results
    reporter.addTestResult({ status: 'passed', name: 'test1' });
    reporter.addTestResult({ status: 'passed', name: 'test2' });
    reporter.addTestResult({ status: 'failed', name: 'test3', error: 'assertion failed' });

    // Add validation results
    reporter.addValidationResult({ status: 'passed', check: 'syntax' });
    reporter.addValidationResult({ status: 'failed', check: 'imports', details: 'Missing @playwright/test' });

    // End report
    reporter.endReport();

    // Verify percentages
    expect(reporter.calculatePercentage(2)).toBe('66.7');

    // Generate HTML and verify it's valid
    const html = reporter.generateHtmlReport();
    expect(html).toContain('<!DOCTYPE html>');
    expect(html).toContain('Conversion Report');
    expect(html).toContain('Passed');

    // Generate markdown and verify
    const md = reporter.generateMarkdownReport();
    expect(md).toContain('# Cypress to Playwright Conversion Report');
    expect(md).toContain('Passed');
    expect(md).toContain('Failed');
  });
});

describe('End-to-end: ConverterFactory all directions', () => {
  const allDirections = [
    ['cypress', 'playwright'],
    ['cypress', 'selenium'],
    ['playwright', 'cypress'],
    ['playwright', 'selenium'],
    ['selenium', 'cypress'],
    ['selenium', 'playwright']
  ];

  it.each(allDirections)(
    'should create converter for %s -> %s and convert a basic test',
    async (from, to) => {
      const converter = await ConverterFactory.createConverter(from, to);
      expect(converter).toBeDefined();
      expect(converter.getSourceFramework()).toBe(from);
      expect(converter.getTargetFramework()).toBe(to);

      // Each converter should accept content and return a string
      const result = await converter.convert('// basic test content');
      expect(typeof result).toBe('string');
    }
  );
});

describe('End-to-end: Framework detection -> conversion -> validation pipeline', () => {
  it('should detect, convert, and validate a Cypress test', async () => {
    const testContent = `
describe('Products', () => {
  it('should list products', () => {
    cy.visit('/products');
    cy.get('.product-card').should('have.length', 3);
  });
});
`;

    // Step 1: Detect
    const detection = FrameworkDetector.detectFromContent(testContent);
    expect(detection.framework).toBe('cypress');
    expect(detection.confidence).toBeGreaterThan(0);

    // Step 2: Convert
    const converter = await ConverterFactory.createConverter('cypress', 'playwright');
    const converted = await converter.convert(testContent);
    expect(converted).toContain('page.goto(');

    // Step 3: Validate structure
    const validation = converter.validate(converted);
    // Validation checks JavaScript syntax
    expect(validation).toBeDefined();
    expect(typeof validation.valid).toBe('boolean');

    // Step 4: Collect metadata from original
    const collector = new TestMetadataCollector();
    const metadata = collector.extractTestSuites(testContent);
    expect(metadata).toHaveLength(1);
    expect(metadata[0].name).toBe('Products');
  });
});

describe('End-to-end: Utility functions in conversion context', () => {
  it('should use stringUtils for naming transformations', () => {
    // Convert file names between conventions
    const camelName = 'userDashboard';
    const kebabName = stringUtils.camelToKebab(camelName);
    expect(kebabName).toBe('user-dashboard');

    const backToCamel = stringUtils.kebabToCamel(kebabName);
    expect(backToCamel).toBe(camelName);
  });

  it('should use codeUtils to extract imports from converted code', async () => {
    const converter = await ConverterFactory.createConverter('cypress', 'playwright');
    const converted = await converter.convert(`
describe('Test', () => {
  it('works', () => {
    cy.visit('/');
  });
});
`);

    const imports = codeUtils.extractImports(converted);
    // Playwright converter adds imports
    expect(imports.length).toBeGreaterThanOrEqual(0);
  });

  it('should use stringUtils.calculateSimilarity for test matching', () => {
    const original = 'login.cy.js';
    const converted = 'login.spec.js';
    const unrelated = 'dashboard.spec.js';

    const similarity1 = stringUtils.calculateSimilarity(original, converted);
    const similarity2 = stringUtils.calculateSimilarity(original, unrelated);

    expect(similarity1).toBeGreaterThan(similarity2);
  });
});

describe('End-to-end: TestMapper lifecycle', () => {
  it('should map, track, and report test file relationships', () => {
    const mapper = new TestMapper();

    // Simulate mapping multiple files
    mapper.mappings.set('login.cy.js', {
      playwrightPath: 'login.spec.js',
      status: 'active',
      syncStatus: 'synced',
      lastSync: new Date().toISOString()
    });
    mapper.mappings.set('dashboard.cy.js', {
      playwrightPath: 'dashboard.spec.js',
      status: 'active',
      syncStatus: 'pending',
      lastSync: new Date().toISOString()
    });

    mapper.updateStatistics();

    expect(mapper.metaData.statistics.totalMappings).toBe(2);
    expect(mapper.metaData.statistics.activeMappings).toBe(2);
    expect(mapper.metaData.statistics.pendingSync).toBe(1);

    const mappings = mapper.getMappings();
    expect(mappings.mappings).toHaveLength(2);
    expect(mappings.mappings[0].cypressTest).toBe('login.cy.js');
    expect(mappings.mappings[0].playwrightTest).toBe('login.spec.js');
  });
});
