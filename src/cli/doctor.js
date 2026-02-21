/**
 * Hamlet Doctor — environment and project diagnostic checks.
 *
 * Pure logic module: returns structured results, no console/chalk/process.exit.
 */

import fs from 'fs/promises';
import { constants as fsConstants } from 'fs';
import path from 'path';

/**
 * @typedef {{ id: string, label: string, status: 'PASS'|'WARN'|'FAIL',
 *   detail: string, verbose: string|null, remediation: string|null }} Check
 *
 * @typedef {{ pass: number, warn: number, fail: number, total: number }} Summary
 *
 * @typedef {{ checks: Check[], summary: Summary, hasFail: boolean }} DoctorResult
 */

/**
 * Run all diagnostic checks against `targetPath`.
 *
 * @param {string} targetPath - Directory to diagnose
 * @param {object} [options]
 * @returns {Promise<DoctorResult>}
 */
export async function runDoctor(targetPath, _options = {}) {
  const resolved = path.resolve(targetPath);

  const checks = [];

  // 1 — Node version (no fs)
  checks.push(checkNodeVersion());

  // 2 — Target path accessibility
  const pathCheck = await checkPath(resolved);
  checks.push(pathCheck);

  // Gate: remaining checks only run when the path is valid
  if (pathCheck.status !== 'FAIL') {
    const results = await Promise.all([
      checkProjectType(resolved),
      checkTestFiles(resolved),
      checkTypeScript(resolved),
      checkOutputPermissions(resolved),
      checkJestEsmWarning(resolved),
    ]);
    for (const r of results) {
      if (r !== null) checks.push(r);
    }
  }

  const summary = { pass: 0, warn: 0, fail: 0, total: checks.length };
  for (const c of checks) {
    if (c.status === 'PASS') summary.pass++;
    else if (c.status === 'WARN') summary.warn++;
    else if (c.status === 'FAIL') summary.fail++;
  }

  return { checks, summary, hasFail: summary.fail > 0 };
}

// ── Individual checks ────────────────────────────────────────────────

function checkNodeVersion() {
  const major = parseInt(process.versions.node.split('.')[0], 10);
  if (major >= 22) {
    return {
      id: 'node-version',
      label: 'Node.js version',
      status: 'PASS',
      detail: `v${process.versions.node} (>= 22 required)`,
      verbose: null,
      remediation: null,
    };
  }
  return {
    id: 'node-version',
    label: 'Node.js version',
    status: 'FAIL',
    detail: `v${process.versions.node} (>= 22 required)`,
    verbose: null,
    remediation: 'Upgrade to Node.js 22 or later',
  };
}

async function checkPath(resolved) {
  let stat;
  try {
    stat = await fs.stat(resolved);
  } catch {
    return {
      id: 'target-path',
      label: 'Target path',
      status: 'FAIL',
      detail: `${resolved} does not exist`,
      verbose: null,
      remediation: 'Provide a valid directory path',
    };
  }

  if (!stat.isDirectory()) {
    return {
      id: 'target-path',
      label: 'Target path',
      status: 'FAIL',
      detail: `${resolved} is not a directory`,
      verbose: null,
      remediation: 'Provide a directory, not a file',
    };
  }

  try {
    await fs.readdir(resolved);
  } catch {
    return {
      id: 'target-path',
      label: 'Target path',
      status: 'FAIL',
      detail: `${resolved} is not readable`,
      verbose: null,
      remediation: 'Check directory permissions',
    };
  }

  return {
    id: 'target-path',
    label: 'Target path',
    status: 'PASS',
    detail: resolved,
    verbose: null,
    remediation: null,
  };
}

async function checkProjectType(resolved) {
  const pkgPath = path.join(resolved, 'package.json');
  let pkg;
  try {
    const raw = await fs.readFile(pkgPath, 'utf8');
    pkg = JSON.parse(raw);
  } catch (err) {
    if (err.code === 'ENOENT') {
      return {
        id: 'project-type',
        label: 'Project type',
        status: 'WARN',
        detail: 'No package.json found',
        verbose: null,
        remediation: 'Run npm init or create a package.json',
      };
    }
    return {
      id: 'project-type',
      label: 'Project type',
      status: 'WARN',
      detail: 'package.json is not valid JSON',
      verbose: null,
      remediation: 'Fix the JSON syntax in package.json',
    };
  }

  const allDeps = {
    ...(pkg.dependencies || {}),
    ...(pkg.devDependencies || {}),
  };

  const KNOWN_FRAMEWORKS = [
    'jest',
    'vitest',
    'mocha',
    'jasmine',
    'cypress',
    '@playwright/test',
    'playwright',
    'selenium-webdriver',
    'webdriverio',
    'puppeteer',
    'testcafe',
  ];

  const detected = KNOWN_FRAMEWORKS.filter((f) => f in allDeps);
  const detail =
    detected.length > 0
      ? `Detected: ${detected.join(', ')}`
      : 'No known test frameworks in dependencies';

  return {
    id: 'project-type',
    label: 'Project type',
    status: 'PASS',
    detail,
    verbose: `package.json has ${Object.keys(allDeps).length} total dependencies`,
    remediation: null,
  };
}

async function checkTestFiles(resolved) {
  const { Scanner } = await import('../core/Scanner.js');
  const { FileClassifier } = await import('../core/FileClassifier.js');

  const scanner = new Scanner();
  const classifier = new FileClassifier();

  const files = await scanner.scan(resolved);
  const frameworks = {};
  let testFileCount = 0;
  let scannedCount = 0;

  for (const file of files) {
    if (file.size > 500_000) continue;
    scannedCount++;
    try {
      const content = await fs.readFile(file.path, 'utf8');
      const classification = classifier.classify(file.path, content);
      if (classification.type === 'test') {
        testFileCount++;
        if (classification.framework) {
          frameworks[classification.framework] =
            (frameworks[classification.framework] || 0) + 1;
        }
      }
    } catch {
      // Skip unreadable files
    }
  }

  const fwSummary = Object.entries(frameworks)
    .sort((a, b) => b[1] - a[1])
    .map(([fw, n]) => `${fw} (${n})`)
    .join(', ');

  if (testFileCount === 0) {
    return {
      id: 'test-files',
      label: 'Test files',
      status: 'WARN',
      detail: 'No test files found',
      verbose: `Scanned ${scannedCount} files`,
      remediation: 'Add test files to enable conversion',
    };
  }

  return {
    id: 'test-files',
    label: 'Test files',
    status: 'PASS',
    detail: `${testFileCount} test file${testFileCount !== 1 ? 's' : ''} found${fwSummary ? ': ' + fwSummary : ''}`,
    verbose: `Scanned ${scannedCount} files`,
    remediation: null,
  };
}

async function checkTypeScript(resolved) {
  let entries;
  try {
    entries = await fs.readdir(resolved);
  } catch {
    return null;
  }

  const tsConfigs = entries.filter(
    (e) => e === 'tsconfig.json' || /^tsconfig\..*\.json$/.test(e)
  );

  if (tsConfigs.length > 0) {
    return {
      id: 'typescript',
      label: 'TypeScript',
      status: 'PASS',
      detail: `Found: ${tsConfigs.join(', ')}`,
      verbose: null,
      remediation: null,
    };
  }

  return {
    id: 'typescript',
    label: 'TypeScript',
    status: 'PASS',
    detail: 'No tsconfig found (not required)',
    verbose: null,
    remediation: null,
  };
}

async function checkOutputPermissions(resolved) {
  try {
    await fs.access(resolved, fsConstants.W_OK);
    return {
      id: 'output-permissions',
      label: 'Output permissions',
      status: 'PASS',
      detail: 'Directory is writable',
      verbose: null,
      remediation: null,
    };
  } catch {
    return {
      id: 'output-permissions',
      label: 'Output permissions',
      status: 'WARN',
      detail: 'Directory is not writable',
      verbose: null,
      remediation: 'Check write permissions on the target directory',
    };
  }
}

async function checkJestEsmWarning(resolved) {
  const pkgPath = path.join(resolved, 'package.json');
  let pkg;
  try {
    const raw = await fs.readFile(pkgPath, 'utf8');
    pkg = JSON.parse(raw);
  } catch {
    return null;
  }

  const allDeps = {
    ...(pkg.dependencies || {}),
    ...(pkg.devDependencies || {}),
  };

  const jestVersion = allDeps.jest;
  if (!jestVersion) return null;

  // Check for ^29 or ~29 or 29.x patterns
  const is29 = /(?:^|\^|~|>=?\s*)29(?:\.|$)/.test(jestVersion);
  if (!is29) return null;

  return {
    id: 'jest-esm',
    label: 'Jest ESM compatibility',
    status: 'WARN',
    detail: `Jest ${jestVersion} requires --experimental-vm-modules for ESM`,
    verbose: null,
    remediation: 'See docs/adr/004-jest-esm-strategy.md for details',
  };
}
