import fs from "fs/promises";
import path from "path";
import chalk from "chalk";
import { TestValidator } from "./validator.js";
import { PluginConverter } from "./plugins.js";
import { VisualComparison } from "./visual.js";
import { TypeScriptConverter } from "./typescript.js";
import { TestMapper } from "./mapper.js";

/**
 * Orchestrates the entire Cypress to Playwright conversion process
 */
export class ConversionOrchestrator {
  constructor(options = {}) {
    this.options = {
      converter: options.converter,
      configConverter: options.configConverter,
      validateTests: options.validateTests ?? true,
      compareVisuals: options.compareVisuals ?? false,
      generateTypes: options.generateTypes ?? true,
      ...options,
    };

    // Initialize components
    this.validator = new TestValidator();
    this.pluginConverter = new PluginConverter();
    this.visualComparison = new VisualComparison();
    this.typeScriptConverter = new TypeScriptConverter();
    this.testMapper = new TestMapper();

    // Conversion statistics
    this.stats = {
      totalFiles: 0,
      convertedFiles: 0,
      skippedFiles: 0,
      errors: [],
      startTime: null,
      endTime: null,
    };
  }

  /**
   * Convert an entire Cypress project to Playwright
   * @param {string} cypressPath - Path to Cypress project
   * @param {string} outputPath - Path for Playwright output
   * @returns {Object} - Conversion report
   */
  async convertProject(cypressPath, outputPath) {
    try {
      this.stats.startTime = new Date();
      console.log(chalk.blue("\nStarting Cypress to Playwright conversion..."));

      // Create output directory
      await fs.mkdir(outputPath, { recursive: true });

      // Phase 1: Project Analysis
      console.log(chalk.blue("\nPhase 1: Analyzing project structure..."));
      const projectStructure = await this.analyzeProject(cypressPath);

      // Phase 2: Convert Configuration
      console.log(chalk.blue("\nPhase 2: Converting configuration..."));
      await this.convertConfiguration(cypressPath, outputPath);

      // Phase 3: Convert Tests
      console.log(chalk.blue("\nPhase 3: Converting test files..."));
      await this.convertTests(projectStructure.testFiles, outputPath);

      // Phase 4: Convert Support Files
      console.log(chalk.blue("\nPhase 4: Converting support files..."));
      await this.convertSupportFiles(projectStructure.supportFiles, outputPath);

      // Phase 5: Convert Plugins
      console.log(chalk.blue("\nPhase 5: Converting plugins..."));
      await this.convertPlugins(projectStructure.plugins, outputPath);

      // Phase 6: Post-Processing
      console.log(chalk.blue("\nPhase 6: Running post-processing..."));
      await this.runPostProcessing(outputPath);

      this.stats.endTime = new Date();
      const report = this.generateReport();

      // Save report
      await this.saveReport(report, outputPath);

      console.log(chalk.green("\n✓ Conversion completed successfully"));
      console.log(
        `Converted ${this.stats.convertedFiles} files in ${this.getExecutionTime()}`,
      );

      return report;
    } catch (error) {
      console.error(chalk.red("\n✗ Conversion failed:"), error);
      throw error;
    }
  }

  /**
   * Analyze Cypress project structure
   * @param {string} projectPath - Path to Cypress project
   * @returns {Object} - Project structure analysis
   */
  async analyzeProject(projectPath) {
    const structure = {
      testFiles: [],
      supportFiles: [],
      plugins: [],
      configs: [],
      fixtures: [],
    };

    async function scanDirectory(dir, structure) {
      const entries = await fs.readdir(dir, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);

        if (entry.isDirectory()) {
          // Skip node_modules
          if (entry.name === "node_modules") continue;
          await scanDirectory(fullPath, structure);
        } else {
          // Categorize files
          if (/\.(spec|test|cy)\.(js|ts)$/.test(entry.name)) {
            structure.testFiles.push(fullPath);
          } else if (/support\/.*\.(js|ts)$/.test(fullPath)) {
            structure.supportFiles.push(fullPath);
          } else if (/plugins\/.*\.(js|ts)$/.test(fullPath)) {
            structure.plugins.push(fullPath);
          } else if (/cypress\.(json|config)\.(js|ts)$/.test(entry.name)) {
            structure.configs.push(fullPath);
          } else if (/fixtures\/.*\.(json|js|ts)$/.test(fullPath)) {
            structure.fixtures.push(fullPath);
          }
        }
      }
    }

    await scanDirectory(projectPath, structure);
    return structure;
  }

  /**
   * Convert Cypress configuration to Playwright
   * @param {string} cypressPath - Path to Cypress project
   * @param {string} outputPath - Path for Playwright output
   */
  async convertConfiguration(cypressPath, outputPath) {
    const configPath = path.join(cypressPath, "cypress.json");

    try {
      if (this.options.configConverter) {
        const playwrightConfig = await this.options.configConverter(configPath);
        if (playwrightConfig) {
          await fs.writeFile(
            path.join(outputPath, "playwright.config.js"),
            playwrightConfig,
          );
          console.log(chalk.green("✓ Configuration converted successfully"));
        }
      }
    } catch (error) {
      console.warn(
        chalk.yellow("Warning: Could not convert configuration:"),
        error.message,
      );
      this.stats.errors.push({
        type: "config",
        message: error.message,
      });
    }
  }

  /**
   * Convert test files to Playwright format
   * @param {string[]} testFiles - Array of test file paths
   * @param {string} outputPath - Path for Playwright output
   */
  async convertTests(testFiles, outputPath) {
    this.stats.totalFiles = testFiles.length;

    for (const testFile of testFiles) {
      try {
        const relativePath = path.relative(process.cwd(), testFile);
        const outputFile = path.join(
          outputPath,
          "tests",
          relativePath.replace(/\.cy\.(js|ts)$/, ".spec.$1"),
        );

        // Ensure output directory exists
        await fs.mkdir(path.dirname(outputFile), { recursive: true });

        // Read and convert test content
        const content = await fs.readFile(testFile, "utf8");
        const converted = await this.options.converter(content);

        // Write converted test
        await fs.writeFile(outputFile, converted);

        // Track mapping
        this.testMapper.addMapping(testFile, outputFile);

        this.stats.convertedFiles++;
        console.log(chalk.green(`✓ Converted ${path.basename(testFile)}`));
      } catch (error) {
        this.stats.skippedFiles++;
        this.stats.errors.push({
          type: "test",
          file: testFile,
          message: error.message,
        });
        console.error(
          chalk.red(`✗ Failed to convert ${testFile}:`),
          error.message,
        );
      }
    }
  }

  /**
   * Convert support files
   * @param {string[]} supportFiles - Array of support file paths
   * @param {string} outputPath - Path for Playwright output
   */
  async convertSupportFiles(supportFiles, outputPath) {
    for (const file of supportFiles) {
      try {
        const relativePath = path.relative(process.cwd(), file);
        const outputFile = path.join(outputPath, "support", relativePath);

        // Ensure output directory exists
        await fs.mkdir(path.dirname(outputFile), { recursive: true });

        // Read and convert content
        const content = await fs.readFile(file, "utf8");
        const converted = await this.options.converter(content);

        // Write converted file
        await fs.writeFile(outputFile, converted);

        console.log(
          chalk.green(`✓ Converted support file ${path.basename(file)}`),
        );
      } catch (error) {
        this.stats.errors.push({
          type: "support",
          file: file,
          message: error.message,
        });
        console.error(
          chalk.red(`✗ Failed to convert support file ${file}:`),
          error.message,
        );
      }
    }
  }

  /**
   * Convert Cypress plugins to Playwright equivalents
   * @param {string[]} plugins - Array of plugin file paths
   * @param {string} outputPath - Path for Playwright output
   */
  async convertPlugins(plugins, outputPath) {
    for (const plugin of plugins) {
      try {
        const converted = await this.pluginConverter.convertPlugin(plugin);
        if (converted) {
          const outputFile = path.join(
            outputPath,
            "plugins",
            path.basename(plugin),
          );
          await fs.writeFile(outputFile, converted);
        }
      } catch (error) {
        this.stats.errors.push({
          type: "plugin",
          file: plugin,
          message: error.message,
        });
      }
    }
  }

  /**
   * Run post-processing tasks
   * @param {string} outputPath - Path to processed files
   */
  async runPostProcessing(outputPath) {
    // Validate converted tests
    if (this.options.validateTests) {
      await this.validator.validateConvertedTests(outputPath);
    }

    // Run visual comparisons
    if (this.options.compareVisuals) {
      await this.visualComparison.compareProjects(outputPath);
    }

    // Generate TypeScript types
    if (this.options.generateTypes) {
      await this.typeScriptConverter.generateTypes(outputPath);
    }

    // Save test mappings
    await this.testMapper.saveMappings(
      path.join(outputPath, "test-mappings.json"),
    );
  }

  /**
   * Generate conversion report
   * @returns {Object} - Detailed conversion report
   */
  generateReport() {
    return {
      statistics: {
        totalFiles: this.stats.totalFiles,
        convertedFiles: this.stats.convertedFiles,
        skippedFiles: this.stats.skippedFiles,
        successRate: `${((this.stats.convertedFiles / this.stats.totalFiles) * 100).toFixed(2)}%`,
        executionTime: this.getExecutionTime(),
      },
      errors: this.stats.errors,
      mappings: this.testMapper.getMappings(),
      validation: this.validator.getResults(),
      timestamp: new Date().toISOString(),
    };
  }

  /**
   * Save conversion report
   * @param {Object} report - Conversion report
   * @param {string} outputPath - Output directory path
   */
  async saveReport(report, outputPath) {
    const reportPath = path.join(outputPath, "conversion-report.json");
    await fs.writeFile(reportPath, JSON.stringify(report, null, 2));
    console.log(chalk.blue(`\nDetailed report saved to: ${reportPath}`));
  }

  /**
   * Calculate execution time
   * @returns {string} - Formatted execution time
   */
  getExecutionTime() {
    if (!this.stats.startTime || !this.stats.endTime) return "N/A";
    const duration = this.stats.endTime - this.stats.startTime;
    return `${(duration / 1000).toFixed(2)} seconds`;
  }
}
