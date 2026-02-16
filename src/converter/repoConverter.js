import fs from 'fs/promises';
import path from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';
import glob from 'fast-glob';
import { convertFile, convertConfig as convertCypressConfig } from '../index.js';
import { fileUtils, logUtils } from '../utils/helpers.js';

const execAsync = promisify(exec);
const logger = logUtils.createLogger('RepoConverter');

export class RepositoryConverter {
  constructor(options = {}) {
    this.options = {
      tempDir: '.hamlet-temp',
      batchSize: 5,
      preserveStructure: true,
      ignore: ['node_modules/**', '**/cypress/plugins/**'],
      ...options
    };

    this.stats = {
      totalFiles: 0,
      converted: 0,
      skipped: 0,
      errors: []
    };
  }

  /**
   * Analyze repository structure for Cypress tests, configs, support files, and plugins
   * @param {string} repoPath - Path to repository
   * @returns {Promise<Object>} - Repository analysis with testFiles, configs, supportFiles, plugins
   */
  async analyzeRepository(repoPath) {
    const testFiles = await this.findCypressTests(repoPath);

    const configPatterns = [
      '**/cypress.json',
      '**/cypress.config.{js,ts}',
    ];
    const configs = await glob(configPatterns, {
      cwd: repoPath,
      absolute: true,
      ignore: this.options.ignore,
    });

    const supportPatterns = [
      '**/cypress/support/**/*.{js,ts}',
    ];
    const supportFiles = await glob(supportPatterns, {
      cwd: repoPath,
      absolute: true,
      ignore: this.options.ignore,
    });

    const pluginPatterns = [
      '**/cypress/plugins/**/*.{js,ts}',
    ];
    const plugins = await glob(pluginPatterns, {
      cwd: repoPath,
      absolute: true,
    });

    return { testFiles, configs, supportFiles, plugins };
  }

  /**
   * Convert a repository's Cypress tests to Playwright
   * @param {string} repoUrl - Repository URL or local path
   * @param {string} outputPath - Output directory path
   * @returns {Promise<Object>} - Conversion results
   */
  async convertRepository(repoUrl, outputPath) {
    try {
      const isRemote = repoUrl.startsWith('http') || repoUrl.startsWith('git@');
      const repoPath = isRemote ? await this.cloneRepository(repoUrl) : repoUrl;

      logger.info(`Processing repository: ${repoPath}`);

      // Find all Cypress tests
      const testFiles = await this.findCypressTests(repoPath);
      this.stats.totalFiles = testFiles.length;

      logger.info(`Found ${testFiles.length} Cypress test files`);

      // Process tests in batches
      const batches = this.createBatches(testFiles);
      for (const batch of batches) {
        await Promise.all(batch.map(file => this.processTestFile(file, repoPath, outputPath)));
      }

      // Convert configuration files
      await this.convertConfigurations(repoPath, outputPath);

      // Cleanup if temporary directory was used
      if (isRemote) {
        await fs.rm(repoPath, { recursive: true, force: true });
      }

      return this.generateReport();

    } catch (error) {
      logger.error(`Repository conversion failed: ${error.message}`);
      throw error;
    }
  }

  /**
   * Clone remote repository
   * @param {string} repoUrl - Repository URL
   * @returns {Promise<string>} - Path to cloned repository
   */
  async cloneRepository(repoUrl) {
    const tempDir = path.join(process.cwd(), this.options.tempDir);
    await fs.mkdir(tempDir, { recursive: true });

    logger.info('Cloning repository...');
    await execAsync(`git clone ${repoUrl} ${tempDir}`);

    return tempDir;
  }

  /**
   * Find all Cypress test files in repository
   * @param {string} repoPath - Repository path
   * @returns {Promise<string[]>} - Array of test file paths
   */
  async findCypressTests(repoPath) {
    const patterns = [
      '**/cypress/integration/**/*.{js,jsx,ts,tsx}',
      '**/cypress/e2e/**/*.{js,jsx,ts,tsx}',
      '**/cypress/component/**/*.{js,jsx,ts,tsx}',
      '**/*.cy.{js,jsx,ts,tsx}'
    ];

    const files = await glob(patterns, {
      cwd: repoPath,
      absolute: true,
      ignore: this.options.ignore
    });

    return files;
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
      const outputFile = this.options.preserveStructure
        ? path.join(outputPath, relativePath).replace(/\.cy\.(js|ts)x?$/, '.spec.$1')
        : path.join(outputPath, path.basename(filePath).replace(/\.cy\.(js|ts)x?$/, '.spec.$1'));

      await convertFile(filePath, outputFile);
      this.stats.converted++;
      logger.info(`Converted: ${relativePath}`);

    } catch (error) {
      this.stats.errors.push({
        file: filePath,
        error: error.message
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
        'cypress.config.ts'
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
      configuration: this.options
    };
  }
}

export async function convertRepo(repoUrl, outputPath, options = {}) {
  const converter = new RepositoryConverter(options);
  return converter.convertRepository(repoUrl, outputPath);
}