#!/usr/bin/env node

import { program } from 'commander';
import chalk from 'chalk';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { createRequire } from 'module';
import { 
  convertFile, 
  convertRepository,
  validateTests,
  generateReport 
} from '../src/index.js';

const require = createRequire(import.meta.url);
const version = require('../package.json').version;

program
  .version(version)
  .description('To be or not to be... in Playwright. A test converter from Cypress to Playwright.');

program
  .command('convert')
  .description('Convert Cypress tests to Playwright format')
  .argument('<source>', 'Source Cypress test file, directory, or repository URL')
  .option('-o, --output <path>', 'Output path for converted tests')
  .option('-c, --config <path>', 'Custom configuration file path')
  .option('-t, --test-type <type>', 'Specify test type (e2e, component, api, etc.)')
  .option('--validate', 'Validate converted tests')
  .option('--report <format>', 'Generate conversion report (html, json, markdown)')
  .option('--preserve-structure', 'Maintain original directory structure')
  .option('--batch-size <number>', 'Number of tests to process in parallel')
  .option('--ignore <pattern>', 'Files to ignore (glob pattern)')
  .option('--hooks', 'Convert test hooks and configurations')
  .option('--plugins', 'Convert Cypress plugins')
  .action(async (source, options) => {
    try {
      console.log(chalk.blue(`Starting conversion process...`));
      
      const isRepository = source.includes('github.com') || source.includes('gitlab.com');
      
      if (isRepository) {
        await convertRepository(source, options.output, options);
      } else {
        await convertFile(source, options.output, options);
      }

      if (options.validate) {
        console.log(chalk.blue('\nValidating converted tests...'));
        await validateTests(options.output);
      }

      if (options.report) {
        console.log(chalk.blue('\nGenerating report...'));
        await generateReport(options.output, options.report);
      }

      console.log(chalk.green('\nConversion completed successfully! 🎭'));
    } catch (error) {
      console.error(chalk.red('Error during conversion:'), error.message);
      if (error.details) {
        console.error(chalk.yellow('Details:'), error.details);
      }
      process.exit(1);
    }
  });

program
  .command('validate')
  .description('Validate converted Playwright tests')
  .argument('<path>', 'Path to converted tests')
  .option('--report <format>', 'Validation report format')
  .action(async (path, options) => {
    try {
      console.log(chalk.blue('Validating Playwright tests...'));
      const results = await validateTests(path);
      if (options.report) {
        await generateReport(results, options.report);
      }
      console.log(chalk.green('Validation completed!'));
    } catch (error) {
      console.error(chalk.red('Validation error:'), error.message);
      process.exit(1);
    }
  });

program
  .command('init')
  .description('Initialize Hamlet configuration')
  .option('-f, --force', 'Overwrite existing configuration')
  .action(async (options) => {
    try {
      console.log(chalk.blue('Initializing Hamlet configuration...'));
      // Add initialization logic here
      console.log(chalk.green('Configuration initialized!'));
    } catch (error) {
      console.error(chalk.red('Initialization error:'), error.message);
      process.exit(1);
    }
  });

program.parse();