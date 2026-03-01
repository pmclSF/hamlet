import type {
  ConversionOptions,
  DetectionResult,
  Framework,
  IConverter,
  PatternDefinition,
} from './index';

export const FRAMEWORKS: {
  CYPRESS: 'cypress';
  PLAYWRIGHT: 'playwright';
  SELENIUM: 'selenium';
  JEST: 'jest';
  VITEST: 'vitest';
  MOCHA: 'mocha';
  JASMINE: 'jasmine';
  JUNIT4: 'junit4';
  JUNIT5: 'junit5';
  TESTNG: 'testng';
  PYTEST: 'pytest';
  UNITTEST: 'unittest';
  NOSE2: 'nose2';
  WEBDRIVERIO: 'webdriverio';
  PUPPETEER: 'puppeteer';
  TESTCAFE: 'testcafe';
};

export class BaseConverter {
  sourceFramework: string | null;
  targetFramework: string | null;
  options: Record<string, unknown>;
  stats: {
    conversions: number;
    warnings: Array<Record<string, unknown>>;
    errors: Array<Record<string, unknown>>;
  };

  constructor(options?: ConversionOptions | Record<string, unknown>);
  convert(content: string, options?: ConversionOptions): Promise<string>;
  convertConfig(configPath: string, options?: ConversionOptions): Promise<string>;
  getImports(testTypes?: string[]): string[];
  detectTestTypes(content: string): string[];
  validate(content: string): { valid: boolean; errors: string[] };
  getStats(): Record<string, unknown>;
  reset(): void;
  addWarning(message: string, context?: Record<string, unknown>): void;
  addError(message: string, context?: Record<string, unknown>): void;
  getSourceFramework(): string | null;
  getTargetFramework(): string | null;
  getConversionDirection(): string;
}

export class PipelineConverter extends BaseConverter {
  constructor(
    sourceFrameworkName: Framework | string,
    targetFrameworkName: Framework | string,
    frameworkDefinitions: Array<Record<string, unknown>>,
    options?: ConversionOptions | Record<string, unknown>
  );
  getLastReport(): Record<string, unknown> | null;
}

export class ConverterFactory {
  static createConverter(
    from: Framework | string,
    to: Framework | string,
    options?: ConversionOptions | Record<string, unknown>
  ): Promise<IConverter>;
  static createConverterSync(
    from: Framework | string,
    to: Framework | string,
    options?: ConversionOptions | Record<string, unknown>,
    converterClasses?: Record<string, new (...args: any[]) => BaseConverter>
  ): IConverter;
  static isSupported(from: Framework | string, to: Framework | string): boolean;
  static isPipelineBacked(
    from: Framework | string,
    to: Framework | string
  ): boolean;
  static getSupportedConversions(): string[];
  static getFrameworks(): Framework[];
}

export class FrameworkDetector {
  static detect(content: string, filePath?: string): DetectionResult;
  static detectFromContent(content: string): DetectionResult;
  static detectFromPath(filePath: string): DetectionResult;
  static getDetectableFrameworks(): string[];
  static isTestFile(content: string): boolean;
}

export class PatternEngine {
  constructor();
  registerPattern(
    category: string,
    sourcePattern: string | RegExp,
    targetReplacement: string | ((match: string, ...groups: string[]) => string),
    options?: { flags?: string; priority?: number; description?: string }
  ): void;
  registerPatterns(
    category: string,
    patterns: Record<string, PatternDefinition['replacement']>,
    options?: { flags?: string; priority?: number; description?: string }
  ): void;
  applyPatterns(content: string, categories?: string[] | null): string;
  getCategories(): string[];
  clear(): void;
}

export class FrameworkRegistry {
  constructor();
}

export class ConversionPipeline {
  constructor(...args: any[]);
}

export class ConfidenceScorer {
  constructor(...args: any[]);
}

export class TodoFormatter {
  constructor(...args: any[]);
}

export class Scanner {
  constructor(...args: any[]);
}

export class FileClassifier {
  constructor(...args: any[]);
}

export class DependencyGraphBuilder {
  constructor(...args: any[]);
}

export class TopologicalSorter {
  constructor(...args: any[]);
}

export class InputNormalizer {
  constructor(...args: any[]);
}

export class ErrorRecovery {
  constructor(...args: any[]);
}

export class OutputValidator {
  constructor(...args: any[]);
}

export class MigrationStateManager {
  constructor(...args: any[]);
}

export class MigrationChecklistGenerator {
  constructor(...args: any[]);
}

export class ImportRewriter {
  constructor(...args: any[]);
}

export class MigrationEngine {
  constructor(...args: any[]);
}

export class SafetyManager {
  constructor(...args: any[]);
}

export class ConfigConverter {
  constructor(...args: any[]);
  convert(
    configContent: string,
    fromFramework: Framework | string,
    toFramework: Framework | string
  ): string;
}

export class MigrationEstimator {
  constructor(...args: any[]);
}

export class ProjectAnalyzer {
  constructor(...args: any[]);
}
