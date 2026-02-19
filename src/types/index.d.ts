/**
 * Hamlet - Multi-framework test converter type definitions
 */

export type Framework =
  | 'cypress'
  | 'playwright'
  | 'selenium'
  | 'jest'
  | 'vitest'
  | 'mocha'
  | 'jasmine'
  | 'webdriverio'
  | 'puppeteer'
  | 'testcafe'
  | 'junit4'
  | 'junit5'
  | 'testng'
  | 'pytest'
  | 'unittest'
  | 'nose2';

export interface ConversionOptions {
  /** Preserve original directory structure */
  preserveStructure?: boolean;
  /** Number of files to process in batch */
  batchSize?: number;
  /** Test type hint (e2e, component, api, etc.) */
  testType?: string;
  /** Custom configuration */
  config?: Record<string, unknown>;
}

export interface ConversionResult {
  /** Converted content */
  content: string;
  /** Source file path */
  sourcePath?: string;
  /** Output file path */
  outputPath?: string;
  /** Warnings generated during conversion */
  warnings?: string[];
  /** Statistics about the conversion */
  stats?: ConversionStats;
}

export interface ConversionStats {
  /** Number of conversions performed */
  conversions: number;
  /** Number of patterns matched */
  patternsMatched: number;
  /** Number of warnings generated */
  warnings: number;
  /** Processing time in milliseconds */
  processingTime?: number;
}

export interface ConversionReport {
  /** Confidence score 0-100 */
  confidence: number;
  /** Confidence level */
  level: 'high' | 'medium' | 'low';
  /** Number of patterns successfully converted */
  converted: number;
  /** Number of unconvertible patterns */
  unconvertible: number;
  /** Number of patterns converted with warnings */
  warnings: number;
  /** Total number of patterns */
  total: number;
  /** Detailed breakdown of issues */
  details: Array<{
    type: 'unconvertible' | 'warning';
    nodeType: string;
    line: number | null;
    source: string;
  }>;
}

export interface DetectionResult {
  /** Detected framework */
  framework: Framework | null;
  /** Confidence score (0-1) */
  confidence: number;
  /** Detection method used */
  method: 'filename' | 'content' | 'combined' | 'path';
  /** Detailed content analysis */
  contentAnalysis?: {
    scores: Record<string, number>;
    details: Record<string, { commands: number; imports: number; keywords: number }>;
  };
  /** Path-based analysis */
  pathAnalysis?: {
    framework: Framework | null;
    confidence: number;
    reason: string;
  };
}

export interface PatternDefinition {
  /** Regex pattern to match */
  pattern: string;
  /** Replacement string or function */
  replacement: string | ((match: string, ...groups: string[]) => string);
  /** Pattern category */
  category?: string;
  /** Priority (higher = applied first) */
  priority?: number;
}

export interface ConverterConfig {
  /** Source framework */
  sourceFramework: Framework;
  /** Target framework */
  targetFramework: Framework;
  /** Custom patterns to add */
  customPatterns?: PatternDefinition[];
  /** Patterns to skip */
  skipPatterns?: string[];
}

/**
 * Base converter interface
 */
export interface IConverter {
  /** Source framework */
  readonly sourceFramework: string;
  /** Target framework */
  readonly targetFramework: string;
  /** Conversion statistics */
  readonly stats: ConversionStats;

  /**
   * Convert test content from source to target framework
   * @param content - Source test content
   * @param options - Conversion options
   * @returns Converted content
   */
  convert(content: string, options?: ConversionOptions): Promise<string>;

  /**
   * Convert configuration file
   * @param configPath - Path to source config
   * @param options - Conversion options
   * @returns Converted config content
   */
  convertConfig(configPath: string, options?: ConversionOptions): Promise<string>;

  /**
   * Detect test types from content
   * @param content - Test content to analyze
   * @returns Array of detected test types
   */
  detectTestTypes(content: string): string[];

  /**
   * Get required imports for target framework
   * @param testTypes - Detected test types
   * @returns Array of import statements
   */
  getImports(testTypes: string[]): string[];
}

/**
 * Converter factory interface
 */
export interface IConverterFactory {
  /**
   * Create a converter for the specified frameworks
   * @param from - Source framework
   * @param to - Target framework
   * @param options - Converter options
   * @returns Converter instance
   */
  createConverter(from: Framework, to: Framework, options?: ConversionOptions): Promise<IConverter>;

  /**
   * Check if a conversion direction is supported
   * @param from - Source framework
   * @param to - Target framework
   * @returns True if supported
   */
  isSupported(from: Framework, to: Framework): boolean;

  /**
   * Get all supported conversion directions
   * @returns Array of "from-to" strings
   */
  getSupportedConversions(): string[];

  /**
   * Get all supported frameworks
   * @returns Array of framework names
   */
  getFrameworks(): Framework[];
}

/**
 * Framework detector interface
 */
export interface IFrameworkDetector {
  /**
   * Detect framework from file content and/or path
   * @param content - File content
   * @param filePath - Optional file path
   * @returns Detection result
   */
  detect(content: string, filePath?: string): DetectionResult;

  /**
   * Detect framework from content only
   * @param content - File content
   * @returns Detection result
   */
  detectFromContent(content: string): DetectionResult;

  /**
   * Detect framework from file path only
   * @param filePath - File path
   * @returns Detection result
   */
  detectFromPath(filePath: string): DetectionResult;
}

/**
 * Pattern engine interface
 */
export interface IPatternEngine {
  /**
   * Register patterns for a category
   * @param category - Pattern category
   * @param patterns - Pattern definitions
   */
  registerPatterns(category: string, patterns: Record<string, string>): void;

  /**
   * Apply all registered patterns to content
   * @param content - Content to transform
   * @param categories - Optional categories to apply
   * @returns Transformed content
   */
  applyPatterns(content: string, categories?: string[]): string;

  /**
   * Get all registered categories
   * @returns Array of category names
   */
  getCategories(): string[];

  /**
   * Clear all registered patterns
   */
  clear(): void;
}

// Export converter classes
export class BaseConverter implements IConverter {
  readonly sourceFramework: string;
  readonly targetFramework: string;
  readonly stats: ConversionStats;

  constructor(options?: ConversionOptions);
  convert(content: string, options?: ConversionOptions): Promise<string>;
  convertConfig(configPath: string, options?: ConversionOptions): Promise<string>;
  detectTestTypes(content: string): string[];
  getImports(testTypes: string[]): string[];
}

export class PipelineConverter extends BaseConverter {
  constructor(
    sourceFrameworkName: string,
    targetFrameworkName: string,
    frameworkDefinitions: object[],
    options?: ConversionOptions
  );
  convert(content: string, options?: ConversionOptions): Promise<string>;
  convertConfig(configPath: string, options?: ConversionOptions): Promise<string>;
  getLastReport(): ConversionReport | null;
}

export class CypressToPlaywright extends BaseConverter {}
export class CypressToSelenium extends BaseConverter {}
export class PlaywrightToCypress extends BaseConverter {}
export class PlaywrightToSelenium extends BaseConverter {}
export class SeleniumToCypress extends BaseConverter {}
export class SeleniumToPlaywright extends BaseConverter {}

export class ConverterFactory implements IConverterFactory {
  static createConverter(from: Framework, to: Framework, options?: ConversionOptions): Promise<IConverter>;
  static isSupported(from: Framework, to: Framework): boolean;
  static getSupportedConversions(): string[];
  static getFrameworks(): Framework[];
  static getConversionMatrix(): Record<Framework, Record<Framework, boolean>>;
}

export class FrameworkDetector implements IFrameworkDetector {
  static detect(content: string, filePath?: string): DetectionResult;
  static detectFromContent(content: string): DetectionResult;
  static detectFromPath(filePath: string): DetectionResult;
  static getDetectableFrameworks(): string[];
  static isTestFile(content: string): boolean;
}

export class PatternEngine implements IPatternEngine {
  registerPatterns(category: string, patterns: Record<string, string>): void;
  applyPatterns(content: string, categories?: string[]): string;
  getCategories(): string[];
  clear(): void;
}

// Convenience functions
export function convertFile(
  inputPath: string,
  outputPath: string,
  options?: ConversionOptions & { from?: Framework; to?: Framework }
): Promise<ConversionResult>;

export function convertRepository(
  repoUrl: string,
  outputPath: string,
  options?: ConversionOptions & { from?: Framework; to?: Framework }
): Promise<ConversionResult[]>;

// Constants
export const FRAMEWORKS: {
  CYPRESS: 'cypress';
  PLAYWRIGHT: 'playwright';
  SELENIUM: 'selenium';
  JEST: 'jest';
  VITEST: 'vitest';
  MOCHA: 'mocha';
  JASMINE: 'jasmine';
  WEBDRIVERIO: 'webdriverio';
  PUPPETEER: 'puppeteer';
  TESTCAFE: 'testcafe';
  JUNIT4: 'junit4';
  JUNIT5: 'junit5';
  TESTNG: 'testng';
  PYTEST: 'pytest';
  UNITTEST: 'unittest';
  NOSE2: 'nose2';
};
