#!/usr/bin/env node

/**
 * Release version-alignment check.
 *
 * Asserts that every version-bearing file in the repo carries the same
 * version, so a release tag is never cut against a half-bumped tree:
 *
 *   - package.json (source of truth)
 *   - package-lock.json (top-level and packages[""])
 *   - extension/vscode/package.json + package-lock.json
 *   - vscode/package.json
 *   - the newest released heading in CHANGELOG.md
 *   - the version pin in scripts/test-release-sanity.mjs
 *
 * Also verifies cheap wiring invariants the release path depends on:
 * the Homebrew workflow keeps its tag-push trigger, and every make
 * target release.yml invokes still exists in the Makefile.
 *
 * Modes:
 *   node scripts/release-alignment.mjs              consistency only
 *   node scripts/release-alignment.mjs --tag vX.Y.Z additionally require
 *                                                   the tag to match
 *
 * On a tag ref (GITHUB_REF=refs/tags/v*), tag mode engages automatically.
 * Exits 0 with a version table on success; exits 1 listing every mismatch.
 */

import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '..');

async function readJSON(relPath) {
  return JSON.parse(await fs.readFile(path.join(rootDir, relPath), 'utf8'));
}

async function readText(relPath) {
  return fs.readFile(path.join(rootDir, relPath), 'utf8');
}

const rows = [];
const errors = [];

const rootPackage = await readJSON('package.json');
const expected = rootPackage.version;
rows.push(['package.json', expected]);

if (!/^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$/.test(expected)) {
  errors.push(`package.json version "${expected}" is not a release version`);
}

// ── Manifest versions ───────────────────────────────────────

const rootLock = await readJSON('package-lock.json');
const extensionPackage = await readJSON('extension/vscode/package.json');
const extensionLock = await readJSON('extension/vscode/package-lock.json');
const legacyVSCodePackage = await readJSON('vscode/package.json');

const manifests = [
  ['package-lock.json', rootLock.version],
  ['package-lock.json packages[""]', rootLock.packages?.['']?.version],
  ['extension/vscode/package.json', extensionPackage.version],
  ['extension/vscode/package-lock.json', extensionLock.version],
  [
    'extension/vscode/package-lock.json packages[""]',
    extensionLock.packages?.['']?.version,
  ],
  ['vscode/package.json', legacyVSCodePackage.version],
];

// ── CHANGELOG: newest released heading ──────────────────────
// "## [Unreleased]" is skipped by the digit-only match; the first
// "## [x.y.z]" heading is the newest released version.

const changelog = await readText('CHANGELOG.md');
const heading = changelog.match(/^## \[(\d+\.\d+\.\d+[^\]]*)\]/m);
if (heading) {
  manifests.push(['CHANGELOG.md newest heading', heading[1]]);
} else {
  errors.push('CHANGELOG.md: no "## [x.y.z]" release heading found');
}

// ── Release-sanity pin ──────────────────────────────────────
// The sanity suite pins the release version; the pin must move in the
// same commit as the manifests. If the pin's shape changes, update the
// pattern here rather than deleting the check.

const sanity = await readText('scripts/test-release-sanity.mjs');
const pin = sanity.match(
  /assert\.equal\(\s*rootPackage\.version,\s*'([^']+)'\s*\)/
);
if (pin) {
  manifests.push(['scripts/test-release-sanity.mjs pin', pin[1]]);
} else {
  errors.push(
    'scripts/test-release-sanity.mjs: version pin not found ' +
      "(expected assert.equal(rootPackage.version, 'x.y.z')); " +
      'update scripts/release-alignment.mjs if the pin moved'
  );
}

for (const [source, version] of manifests) {
  rows.push([source, version ?? '(missing)']);
  if (version !== expected) {
    errors.push(`${source}: "${version}" does not match "${expected}"`);
  }
}

// ── Tag mode ────────────────────────────────────────────────

const tagFlag = process.argv.indexOf('--tag');
let tag = tagFlag !== -1 ? process.argv[tagFlag + 1] : undefined;
if (!tag && process.env.GITHUB_REF?.startsWith('refs/tags/')) {
  tag = process.env.GITHUB_REF.slice('refs/tags/'.length);
}
if (tagFlag !== -1 && !tag) {
  errors.push('--tag requires a value (e.g. --tag v1.2.3)');
}
if (tag) {
  const tagVersion = tag.replace(/^v/, '');
  rows.push([`tag ${tag}`, tagVersion]);
  if (tagVersion !== expected) {
    errors.push(`tag ${tag}: "${tagVersion}" does not match "${expected}"`);
  }
}

// ── Release wiring ──────────────────────────────────────────

const homebrewWorkflow = await readText(
  '.github/workflows/homebrew-update.yml'
);
if (!/push:\s*\n\s+tags:\s*\n\s+-\s+'v\*'/.test(homebrewWorkflow)) {
  errors.push(
    '.github/workflows/homebrew-update.yml: tag-push trigger ' +
      "(push.tags: 'v*') not found"
  );
}

const releaseWorkflow = await readText('.github/workflows/release.yml');
const makefile = await readText('Makefile');
for (const [, target] of releaseWorkflow.matchAll(/\bmake ([a-z][a-z-]*)/g)) {
  if (!new RegExp(`^${target}:`, 'm').test(makefile)) {
    errors.push(
      `release.yml runs "make ${target}" but Makefile has no "${target}:" target`
    );
  }
}

// ── Report ──────────────────────────────────────────────────

const width = Math.max(...rows.map(([source]) => source.length));
for (const [source, version] of rows) {
  console.log(`  ${source.padEnd(width)}  ${version}`);
}

if (errors.length > 0) {
  console.error('\nRelease alignment FAILED:');
  for (const error of errors) {
    console.error(`  - ${error}`);
  }
  process.exit(1);
}

console.log('\nRelease alignment OK.');
