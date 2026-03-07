import fs from 'fs/promises';
import path from 'path';
import { execFile } from 'child_process';
import { promisify } from 'util';
import glob from 'fast-glob';
import {
  convertFile,
  convertConfig as convertCypressConfig,
  convertConfig,
} from './fileConverter.js';
import { BatchProcessor } from './batchProcessor.js';
import { DependencyAnalyzer } from './dependencyAnalyzer.js';
import { TestMetadataCollector } from './metadataCollector.js';
import { PluginConverter } from './plugins.js';
import { TestMapper } from './mapper.js';
import { ConversionReporter } from '../utils/reporter.js';
import { fileUtils, logUtils } from '../utils/helpers.js';
import { Scanner } from '../core/Scanner.js';
import { FileClassifier } from '../core/FileClassifier.js';
import { buildOutputFilename } from '../cli/outputHelpers.js';

const execFileAsync = promisify(execFile);
const logger = logUtils.createLogger('RepoConverter');

function isRemoteRepository(repoRef) {
  return repoRef.startsWith('https://') || repoRef.startsWith('git@');
}

/**
 * Validate a repository URL to prevent command injection.
 * Allows https:// and git@ (SSH) URLs that look like valid git repos.
 * @param {string} url - Repository URL to validate
 * @throws {Error} If the URL is invalid or contains dangerous characters
 */
export function validateRepoUrl(url) {
  if (typeof url !== 'string' || !url) {
    throw new Error('Invalid repository URL: must be a non-empty string');
  }

  // Reject null bytes, control characters, and shell metacharacters
  // eslint-disable-next-line no-control-regex
  if (/[\0\x01-\x1f]/.test(url) || /[;`|&$(){}[\]!#~]/.test(url)) {
    throw new Error('Invalid repository URL: contains disallowed characters');
  }

  // Must start with a recognized protocol or SSH prefix
  if (!url.startsWith('https://') && !url.startsWith('git@')) {
    throw new Error('Invalid repository URL: must start with https:// or git@');
  }

  // Basic structural check: host/path pattern, optional .git suffix
  if (url.startsWith('git@')) {
    // SSH: git@host:org/repo.git
    if (!/^git@[\w.-]+:[\w./_-]+(\.git)?$/.test(url)) {
      throw new Error(
        'Invalid repository URL: does not match expected SSH pattern'
      );
    }
  } else {
    // HTTPS: protocol://host/path with optional .git
    if (!/^https:\/\/[\w.-]+(:\d+)?\/[\w./_-]+(\.git)?$/.test(url)) {
      throw new Error(
        'Invalid repository URL: does not match expected URL pattern'
      );
    }
  }
}

export class RepositoryConverter {
  constructor(options = {}) {
    this.options = {
      tempDir: '.hamlet-temp',
      batchSize: 5,
      preserveStructure: true,
      ignore: ['node_modules/**', '**/cypress/plugins/**'],
      ...options,
    };

    this.stats = {
      totalFiles: 0,
      converted: 0,
      skipped: 0,
      errors: [],
    };
  }

  /**
   * Analyze repository structure for Cypress tests, configs, support files, and plugins
   * @param {string} repoPath - Path to repository
   * @returns {Promise<Object>} - Repository analysis with testFiles, configs, supportFiles, plugins
   */
  async analyzeRepository(repoPath) {
    const sourceFramework = (this.options.from || 'cypress').toLowerCase();
    const testFiles = await this.findFrameworkTests(repoPath, sourceFramework);

    const frameworkConfigPatterns = {
      cypress: ['**/cypress.json', '**/cypress.config.{js,ts}'],
      playwright: ['**/playwright.config.{js,ts,mjs,cjs,mts,cts}'],
      webdriverio: ['**/wdio.conf.{js,ts,mjs,cjs,mts,cts}'],
      jest: ['**/jest.config.{js,ts,mjs,cjs,mts,cts,json}'],
      vitest: ['**/vitest.config.{js,ts,mjs,cjs,mts,cts}'],
      mocha: ['**/.mocharc.{js,json,yaml,yml,cjs,mjs}', '**/mocha.opts'],
      jasmine: ['**/jasmine.json'],
      pytest: ['**/pytest.ini', '**/pyproject.toml', '**/setup.cfg'],
      nose2: ['**/nose2.cfg', '**/unittest.cfg'],
      testng: ['**/testng.xml', '**/pom.xml', '**/build.gradle*'],
      junit4: ['**/pom.xml', '**/build.gradle*'],
      junit5: ['**/pom.xml', '**/build.gradle*'],
      unittest: [],
      selenium: [],
      puppeteer: [],
      testcafe: [],
    };
    const configPatterns = frameworkConfigPatterns[sourceFramework] || [];
    const configs =
      configPatterns.length > 0
        ? await glob(configPatterns, {
            cwd: repoPath,
            absolute: true,
            ignore: this.options.ignore,
          })
        : [];

    const isCypress = sourceFramework === 'cypress';
    const supportPatterns = isCypress
      ? ['**/cypress/support/**/*.{js,ts}']
      : [];
    const supportFiles =
      supportPatterns.length > 0
        ? await glob(supportPatterns, {
            cwd: repoPath,
            absolute: true,
            ignore: this.options.ignore,
          })
        : [];

    const pluginPatterns = isCypress ? ['**/cypress/plugins/**/*.{js,ts}'] : [];
    const plugins =
      pluginPatterns.length > 0
        ? await glob(pluginPatterns, {
            cwd: repoPath,
            absolute: true,
          })
        : [];

    return { testFiles, configs, supportFiles, plugins };
  }

  /**
   * Convert a repository's Cypress tests to Playwright
   * @param {string} repoUrl - Repository URL or local path
   * @param {string} outputPath - Output directory path
   * @returns {Promise<Object>} - Conversion results
   */
  async convertRepository(repoUrl, outputPath) {
    const fromFramework = (this.options.from || 'cypress').toLowerCase();
    const isRemote = isRemoteRepository(repoUrl);
    let repoPath = repoUrl;

    try {
      if (isRemote) {
        repoPath = await this.cloneRepository(repoUrl);
      }

      logger.info(`Processing repository: ${repoPath}`);

      // Find all source-framework tests
      const testFiles = await this.findFrameworkTests(repoPath, fromFramework);
      this.stats.totalFiles = testFiles.length;

      logger.info(`Found ${testFiles.length} ${fromFramework} test files`);

      // Process tests in batches
      const batches = this.createBatches(testFiles);
      for (const batch of batches) {
        await Promise.all(
          batch.map((file) => this.processTestFile(file, repoPath, outputPath))
        );
      }

      // Convert configuration files
      await this.convertConfigurations(repoPath, outputPath);

      return this.generateReport();
    } catch (error) {
      logger.error(`Repository conversion failed: ${error.message}`);
      throw error;
    } finally {
      if (isRemote && repoPath !== repoUrl) {
        try {
          await fs.rm(repoPath, { recursive: true, force: true });
        } catch (cleanupError) {
          logger.warn(
            `Failed to clean up temporary clone at ${repoPath}: ${cleanupError.message}`
          );
        }
      }
    }
  }

  /**
   * Clone remote repository
   * @param {string} repoUrl - Repository URL
   * @returns {Promise<string>} - Path to cloned repository
   */
  async cloneRepository(repoUrl) {
    validateRepoUrl(repoUrl);

    const tempRoot = path.resolve(process.cwd(), this.options.tempDir);
    await fs.mkdir(tempRoot, { recursive: true });
    const cloneDir = await fs.mkdtemp(path.join(tempRoot, 'repo-'));

    logger.info('Cloning repository...');
    try {
      await execFileAsync('git', ['clone', '--', repoUrl, cloneDir]);
    } catch (error) {
      await fs.rm(cloneDir, { recursive: true, force: true });
      throw error;
    }

    return cloneDir;
  }

  /**
   * Find test files for the requested source framework.
   * Uses content-aware classification rather than framework-specific path globs.
   * @param {string} repoPath - Repository path
   * @param {string} sourceFramework - Source framework name
   * @returns {Promise<string[]>} - Array of test file paths
   */
  async findFrameworkTests(repoPath, sourceFramework) {
    const scanner = new Scanner();
    const classifier = new FileClassifier();
    const framework = (sourceFramework || 'cypress').toLowerCase();
    const frameworkAgnosticUnitRuntimes = new Set([
      'jest',
      'vitest',
      'mocha',
      'jasmine',
    ]);

    const files = await scanner.scan(repoPath);
    const testFiles = [];

    for (const file of files) {
      if (!/\.(js|jsx|ts|tsx|py|java)$/i.test(file.path)) continue;

      try {
        const content = await fs.readFile(file.path, 'utf8');
        const classification = classifier.classify(file.path, content);
        if (classification.type !== 'test') continue;

        const detectedFramework = classification.framework
          ? classification.framework.toLowerCase()
          : null;

        if (detectedFramework === framework) {
          testFiles.push(file.path);
          continue;
        }

        // For unit-test runtimes with overlapping syntax, fall back to test-file
        // inclusion when framework detection is inconclusive.
        if (
          !detectedFramework &&
          frameworkAgnosticUnitRuntimes.has(framework)
        ) {
          testFiles.push(file.path);
        }
      } catch (_err) {
        // Skip unreadable files.
      }
    }

    return testFiles;
  }

  /**
   * Backward-compat helper for existing Cypress-specific callers.
   * @param {string} repoPath - Repository path
   * @returns {Promise<string[]>}
   */
  async findCypressTests(repoPath) {
    return this.findFrameworkTests(repoPath, 'cypress');
  }

  /**
   * Create batches of files for processing
   * @param {string[]} files - Array of file paths
   * @returns {Array<string[]>} - Array of batches
   */
  createBatches(files) {
    const batches = [];
    for (let i = 0; i < files.length; i += this.options.batchSize) {
      batches.push(files.slice(i, i + this.options.batchSize));
    }
    return batches;
  }

  /**
   * Process single test file
   * @param {string} filePath - Test file path
   * @param {string} repoPath - Repository root path
   * @param {string} outputPath - Output directory path
   */
  async processTestFile(filePath, repoPath, outputPath) {
    try {
      const relativePath = path.relative(repoPath, filePath);
      const toFramework = (this.options.to || 'playwright').toLowerCase();
      const fromFramework = (this.options.from || 'cypress').toLowerCase();
      const outputFilename = buildOutputFilename(
        path.basename(filePath),
        toFramework
      );
      const outputFile = this.options.preserveStructure
        ? path.join(outputPath, path.dirname(relativePath), outputFilename)
        : path.join(outputPath, outputFilename);

      await convertFile(filePath, outputFile, {
        ...this.options,
        from: fromFramework,
        to: toFramework,
      });
      this.stats.converted++;
      logger.info(`Converted: ${relativePath}`);
    } catch (error) {
      this.stats.errors.push({
        file: filePath,
        error: error.message,
      });
      this.stats.skipped++;
      logger.error(`Failed to convert ${filePath}: ${error.message}`);
    }
  }

  /**
   * Convert Cypress configuration files
   * @param {string} repoPath - Repository path
   * @param {string} outputPath - Output directory path
   */
  async convertConfigurations(repoPath, outputPath) {
    try {
      const configFiles = [
        'cypress.json',
        'cypress.config.js',
        'cypress.config.ts',
      ];

      for (const configFile of configFiles) {
        const configPath = path.join(repoPath, configFile);
        if (await fileUtils.fileExists(configPath)) {
          // Convert and save configuration
          const converted = await this.convertConfig(configPath);
          await fs.writeFile(
            path.join(outputPath, 'playwright.config.js'),
            converted
          );
          break;
        }
      }
    } catch (error) {
      logger.warn(`Configuration conversion failed: ${error.message}`);
    }
  }

  /**
   * Convert Cypress config to Playwright config
   * @param {string} configPath - Path to Cypress config
   * @returns {Promise<string>} - Playwright config content
   */
  async convertConfig(configPath) {
    return await convertCypressConfig(configPath, this.options);
  }

  /**
   * Generate conversion report
   * @returns {Object} - Conversion report
   */
  generateReport() {
    return {
      stats: this.stats,
      timestamp: new Date().toISOString(),
      duration: process.hrtime(),
      configuration: this.options,
    };
  }
}

export async function convertRepo(repoUrl, outputPath, options = {}) {
  const converter = new RepositoryConverter(options);
  return converter.convertRepository(repoUrl, outputPath);
}

const converterLogger = logUtils.createLogger('Converter');

/**
 * Convert a repository of Cypress tests to Playwright
 * @param {string} repoPath - Path to repository or repository URL
 * @param {string} outputPath - Output directory path
 * @param {Object} options - Conversion options
 */
export async function convertRepository(repoPath, outputPath, options = {}) {
  const fromFramework = (
    options.from ||
    options.fromFramework ||
    'cypress'
  ).toLowerCase();
  const toFramework = (
    options.to ||
    options.toFramework ||
    'playwright'
  ).toLowerCase();

  // Initialize components
  const repoConverter = new RepositoryConverter({
    ...options,
    from: fromFramework,
    to: toFramework,
  });
  const batchProcessor = new BatchProcessor(options);
  const metadataCollector = new TestMetadataCollector();
  const dependencyAnalyzer = new DependencyAnalyzer();
  const reporter = options.reporter || new ConversionReporter();
  const testMapper = new TestMapper();

  const isRemoteRepo = isRemoteRepository(repoPath);
  let workingPath = repoPath;

  try {
    converterLogger.info(`Starting repository conversion: ${repoPath}`);

    // Clone repository if it's a URL
    if (isRemoteRepo) {
      validateRepoUrl(repoPath);
      workingPath = await repoConverter.cloneRepository(repoPath);
    }
    await fs.mkdir(outputPath, { recursive: true });

    // Analyze repository structure
    const structure = await repoConverter.analyzeRepository(workingPath);
    converterLogger.info(`Found ${structure.testFiles.length} test files`);

    // Convert configuration files
    const configs = await Promise.all(
      structure.configs.map(async (config) => {
        const outputConfig = path.join(
          outputPath,
          path
            .basename(config)
            .replace(fromFramework, toFramework)
            .replace('.json', '.config.js')
        );

        try {
          const converted = await convertConfig(config, {
            ...options,
            from: fromFramework,
            to: toFramework,
          });
          await fs.writeFile(outputConfig, converted);
          return { source: config, output: outputConfig, status: 'success' };
        } catch (error) {
          converterLogger.error(`Failed to convert config ${config}:`, error);
          return { source: config, status: 'error', error: error.message };
        }
      })
    );

    // Process tests in batches
    const batchSummary = await batchProcessor.processBatch(
      structure.testFiles,
      async (file) => {
        const relativePath = path.relative(workingPath, file);
        const outputFilename = buildOutputFilename(
          path.basename(file),
          toFramework
        );
        const outputFile = path.join(
          outputPath,
          'tests',
          path.dirname(relativePath),
          outputFilename
        );

        try {
          // Convert individual test file
          const _result = await convertFile(file, outputFile, {
            ...options,
            reporter,
            from: fromFramework,
            to: toFramework,
          });

          // Collect metadata and analyze dependencies
          const metadata = await metadataCollector.collectMetadata(file);
          const dependencies =
            await dependencyAnalyzer.analyzeDependencies(file);

          // Add to test mapper
          await testMapper.addMapping(file, outputFile);

          return {
            source: file,
            output: outputFile,
            status: 'success',
            metadata,
            dependencies,
          };
        } catch (error) {
          converterLogger.error(`Failed to convert ${file}:`, error);
          return {
            source: file,
            status: 'error',
            error: error.message,
          };
        }
      }
    );
    const batchResults = batchSummary.results || [];

    // Convert support files
    const supportResults = await Promise.all(
      structure.supportFiles.map(async (file) => {
        const relativePath = path.relative(workingPath, file);
        const outputFile = path.join(outputPath, 'support', relativePath);

        try {
          await convertFile(file, outputFile, {
            ...options,
            from: fromFramework,
            to: toFramework,
          });
          return { source: file, output: outputFile, status: 'success' };
        } catch (error) {
          converterLogger.error(
            `Failed to convert support file ${file}:`,
            error
          );
          return { source: file, status: 'error', error: error.message };
        }
      })
    );

    // Convert plugins if requested
    let pluginResults = [];
    if (options.convertPlugins) {
      const pluginConverter = new PluginConverter();
      pluginResults = await Promise.all(
        structure.plugins.map(async (plugin) => {
          try {
            const converted = await pluginConverter.convertPlugin(plugin);
            const convertedPlugins = (converted.conversions || []).filter(
              (entry) => entry.status === 'converted'
            );
            if (convertedPlugins.length === 0) {
              return {
                source: plugin,
                status: 'skipped',
                reason: 'No known Cypress plugins detected',
              };
            }

            const outputFile = path.join(
              outputPath,
              'plugins',
              path.basename(plugin)
            );
            await fs.mkdir(path.dirname(outputFile), { recursive: true });
            const pluginOutput = [
              '// Converted plugin helpers generated by Hamlet',
              converted.imports || '',
              '',
              converted.setup || '',
              '',
              `export const pluginConfig = ${JSON.stringify(converted.config || {}, null, 2)};`,
              '',
            ].join('\n');
            await fs.writeFile(outputFile, pluginOutput, 'utf8');
            return { source: plugin, output: outputFile, status: 'success' };
          } catch (error) {
            converterLogger.error(`Failed to convert plugin ${plugin}:`, error);
            return { source: plugin, status: 'error', error: error.message };
          }
        })
      );
    }

    // Generate comprehensive report
    const report = {
      summary: {
        totalFiles: structure.testFiles.length,
        convertedFiles: batchResults.filter((r) => r.status === 'success')
          .length,
        failedFiles: batchResults.filter((r) => r.status === 'error').length,
        configurationFiles: configs.length,
        supportFiles: supportResults.length,
        plugins: pluginResults.length,
      },
      testResults: batchResults,
      configResults: configs,
      supportResults: supportResults,
      pluginResults: pluginResults,
      metadata: metadataCollector.generateReport(),
      dependencies: dependencyAnalyzer.generateReport(),
      mappings: testMapper.getMappings(),
      timestamp: new Date().toISOString(),
      duration: process.hrtime(),
    };

    // Save report
    if (options.report) {
      const reportPath = path.join(outputPath, 'conversion-report.json');
      await fs.writeFile(reportPath, JSON.stringify(report, null, 2));
      converterLogger.info(`Report saved to: ${reportPath}`);

      // Generate HTML report if requested
      if (options.report === 'html') {
        const htmlReport = reporter.generateHtmlReport(report);
        await fs.writeFile(
          path.join(outputPath, 'conversion-report.html'),
          htmlReport
        );
      }
    }

    converterLogger.success('Repository conversion completed successfully');
    return report;
  } catch (error) {
    converterLogger.error('Repository conversion failed:', error);
    throw error;
  } finally {
    if (isRemoteRepo && workingPath !== repoPath) {
      try {
        await fs.rm(workingPath, { recursive: true, force: true });
      } catch (cleanupError) {
        converterLogger.warn(
          `Failed to clean up temporary clone at ${workingPath}: ${cleanupError.message}`
        );
      }
    }
  }
}
