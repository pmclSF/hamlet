/**
 * Consumer type-test — verifies that the public API types are valid and usable.
 * This file is compiled with tsc --noEmit; it is never executed at runtime.
 */
import type {
  Framework,
  ConversionOptions,
  ConversionResult,
  ConversionStats,
  ConversionReport,
  DetectionResult,
  PatternDefinition,
  ConverterConfig,
  IConverter,
  IConverterFactory,
  IFrameworkDetector,
  IPatternEngine,
} from 'hamlet-converter';

import {
  convertFile,
  convertRepository,
  processTestFiles,
  validateTests,
  generateReport,
  convertCypressToPlaywright,
  convertConfig,
  VERSION,
  SUPPORTED_TEST_TYPES,
  DEFAULT_OPTIONS,
  ConverterFactory,
  FrameworkDetector,
  PatternEngine,
  BaseConverter,
  PipelineConverter,
  BatchProcessor,
  ConversionReporter,
  RepositoryConverter,
  FRAMEWORKS,
} from 'hamlet-converter';

// ── Framework type ──
const fw: Framework = 'jest';
const _fws: Framework[] = ['cypress', 'playwright', 'selenium', 'jest', 'vitest',
  'mocha', 'jasmine', 'webdriverio', 'puppeteer', 'testcafe',
  'junit4', 'junit5', 'testng', 'pytest', 'unittest', 'nose2'];

// ── ConversionOptions ──
const opts: ConversionOptions = {
  preserveStructure: true,
  batchSize: 5,
  testType: 'e2e',
  config: { custom: true },
};

// ── convertFile ──
async function testConvertFile() {
  const result: ConversionResult = await convertFile('input.js', 'output.js', {
    from: 'jest',
    to: 'vitest',
  });
  const _content: string = result.content;
}

// ── convertRepository ──
async function testConvertRepo() {
  const results: ConversionResult[] = await convertRepository('.', 'out/', {
    from: 'cypress',
    to: 'playwright',
  });
  const _first: ConversionResult = results[0];
}

// ── processTestFiles ──
async function testProcessFiles() {
  const results: ConversionResult[] = await processTestFiles(['a.js'], {
    from: 'jest',
    to: 'vitest',
  });
  const _r = results;
}

// ── validateTests ──
async function testValidate() {
  const _result: object = await validateTests('tests/');
}

// ── generateReport ──
async function testReport() {
  await generateReport('report.json', 'json', { summary: {} });
}

// ── convertCypressToPlaywright ──
async function testConvertCyPw() {
  const _output: string = await convertCypressToPlaywright("cy.visit('/')");
}

// ── convertConfig ──
async function testConvertConfig() {
  const _output: string = await convertConfig('jest.config.js', {
    from: 'jest',
    to: 'vitest',
  });
}

// ── Constants ──
const _version: string = VERSION;
const _types: string[] = SUPPORTED_TEST_TYPES;
const _defaults = DEFAULT_OPTIONS;
const _bs: number = _defaults.batchSize;
const _ts: boolean = _defaults.typescript;

// ── FRAMEWORKS constant ──
const _cy: 'cypress' = FRAMEWORKS.CYPRESS;
const _pw: 'playwright' = FRAMEWORKS.PLAYWRIGHT;

// ── ConverterFactory ──
async function testFactory() {
  const converter: IConverter = await ConverterFactory.createConverter('jest', 'vitest');
  const _output: string = await converter.convert('test code');
  const _supported: boolean = ConverterFactory.isSupported('jest', 'vitest');
  const _directions: string[] = ConverterFactory.getSupportedConversions();
  const _frameworks: Framework[] = ConverterFactory.getFrameworks();
}

// ── FrameworkDetector ──
function testDetector() {
  const result: DetectionResult = FrameworkDetector.detect('code', 'file.js');
  const _fw: Framework | null = result.framework;
  const _conf: number = result.confidence;
  const _frameworks: string[] = FrameworkDetector.getDetectableFrameworks();
  const _isTest: boolean = FrameworkDetector.isTestFile('code');
}

// ── PatternEngine ──
function testPatternEngine() {
  const engine = new PatternEngine();
  engine.registerPatterns('assertions', { 'expect\\(': 'assert(' });
  const _output: string = engine.applyPatterns('expect(true)');
  const _cats: string[] = engine.getCategories();
  engine.clear();
}

// ── Public classes ──
function testClasses() {
  const _base = new BaseConverter();
  const _batch = new BatchProcessor();
  const _reporter = new ConversionReporter({ format: 'json' });
  const _repo = new RepositoryConverter();
}

// ── Interface conformance checks ──
function testInterfaces(
  _converter: IConverter,
  _factory: IConverterFactory,
  _detector: IFrameworkDetector,
  _engine: IPatternEngine,
) {
  // These parameters prove the interfaces are valid
}

// Suppress unused variable warnings — this file is type-checked only, never run
void testConvertFile;
void testConvertRepo;
void testProcessFiles;
void testValidate;
void testReport;
void testConvertCyPw;
void testConvertConfig;
void testFactory;
void testDetector;
void testPatternEngine;
void testClasses;
void testInterfaces;
void fw;
void opts;
