#!/usr/bin/env node

/**
 * Verify that npm pack produces a valid, installable package with correct exports.
 * Run via: npm run release:verify (after format:check, lint, and test).
 */

import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '..');

let tmpDir;
let tarball;

try {
  console.log('Packing tarball...');
  const packOutput = execFileSync(
    'npm',
    ['pack', '--pack-destination', os.tmpdir()],
    { cwd: rootDir, encoding: 'utf8' }
  ).trim();

  const tarballName = packOutput.split('\n').pop().trim();
  tarball = path.join(os.tmpdir(), tarballName);
  console.log(`  Created: ${tarballName}`);

  // List contents for review
  console.log('\nTarball contents:');
  const listing = execFileSync('tar', ['tzf', tarball], { encoding: 'utf8' });
  const files = listing.trim().split('\n');
  console.log(`  ${files.length} files`);

  // Check for unexpected files
  const unexpected = files.filter(
    (f) =>
      f.includes('node_modules/') ||
      f.includes('.env') ||
      f.includes('.github/') ||
      f.includes('test/')
  );
  if (unexpected.length > 0) {
    console.error('\nUnexpected files in tarball:');
    unexpected.forEach((f) => console.error(`  ${f}`));
    process.exit(1);
  }

  // Install in temp dir and verify imports
  console.log('\nInstalling in temp directory...');
  tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-verify-'));
  execFileSync('npm', ['init', '-y'], { cwd: tmpDir, encoding: 'utf8' });
  execFileSync('npm', ['install', tarball], { cwd: tmpDir, encoding: 'utf8' });

  console.log('Verifying exports...');
  const check = execFileSync(
    'node',
    [
      '--input-type=module',
      '-e',
      `
      import { VERSION, convertFile, convertRepository, BatchProcessor, ConversionReporter } from 'hamlet-converter';
      const pkg = JSON.parse(await import('fs/promises').then(f => f.default.readFile('node_modules/hamlet-converter/package.json', 'utf8')));
      const errors = [];
      if (VERSION !== pkg.version) errors.push('VERSION ' + VERSION + ' !== package.json ' + pkg.version);
      if (typeof convertFile !== 'function') errors.push('convertFile is not a function');
      if (typeof convertRepository !== 'function') errors.push('convertRepository is not a function');
      if (typeof BatchProcessor !== 'function') errors.push('BatchProcessor is not a function');
      if (typeof ConversionReporter !== 'function') errors.push('ConversionReporter is not a function');
      if (errors.length) { console.error(errors.join('\\n')); process.exit(1); }
      console.log('  VERSION=' + VERSION + ' (matches package.json)');
      console.log('  convertFile: ok');
      console.log('  convertRepository: ok');
      console.log('  BatchProcessor: ok');
      console.log('  ConversionReporter: ok');
    `,
    ],
    { cwd: tmpDir, encoding: 'utf8' }
  );
  console.log(check);

  console.log('Release verification passed.');
} catch (error) {
  console.error('Release verification failed:', error.message);
  process.exit(1);
} finally {
  if (tmpDir) await fs.rm(tmpDir, { recursive: true, force: true });
  if (tarball) await fs.rm(tarball, { force: true });
}
