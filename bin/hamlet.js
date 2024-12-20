#!/usr/bin/env node

import { program } from 'commander';
import chalk from 'chalk';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { createRequire } from 'module';
import { convertFile } from '../src/index.js';

const require = createRequire(import.meta.url);
const version = require('../package.json').version;

program
  .version(version)
  .description('To be or not to be... in Playwright. A test converter from Cypress to Playwright.');

program
  .command('convert')
  .description('Convert Cypress tests to Playwright format')
  .argument('<source>', 'Source Cypress test file or directory')
  .option('-o, --output <path>', 'Output file or directory path')
  .action(async (source, options) => {
    try {
      console.log(chalk.blue(`Converting ${source} to Playwright format...`));
      await convertFile(source, options.output);
      console.log(chalk.green('Conversion completed successfully!'));
    } catch (error) {
      console.error(chalk.red('Error during conversion:'), error.message);
      process.exit(1);
    }
  });

program.parse();