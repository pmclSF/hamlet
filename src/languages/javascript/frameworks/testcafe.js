/**
 * TestCafe framework definition.
 *
 * Provides detect, parse, and emit for the TestCafe E2E testing framework.
 * TestCafe is source-only — no conversions TO TestCafe are supported.
 * emit() is a stub that returns the source unchanged.
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  ImportStatement,
  RawCode,
  Comment,
} from '../../../core/ir.js';

function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // TestCafe imports (strong signals)
  if (/import\s+\{[^}]*Selector[^}]*\}\s+from\s+['"]testcafe['"]/.test(source)) score += 30;
  if (/import\s+\{[^}]*Role[^}]*\}\s+from\s+['"]testcafe['"]/.test(source)) score += 15;
  if (/from\s+['"]testcafe['"]/.test(source)) score += 25;

  // fixture backtick syntax (strong signals)
  if (/\bfixture\s*`/.test(source)) score += 25;
  if (/\.page\s*`/.test(source)) score += 15;

  // TestCafe t.* actions
  if (/await\s+t\.typeText\s*\(/.test(source)) score += 15;
  if (/await\s+t\.click\s*\(/.test(source)) score += 15;
  if (/await\s+t\.doubleClick\s*\(/.test(source)) score += 10;
  if (/await\s+t\.rightClick\s*\(/.test(source)) score += 10;
  if (/await\s+t\.hover\s*\(/.test(source)) score += 10;
  if (/await\s+t\.pressKey\s*\(/.test(source)) score += 10;
  if (/await\s+t\.navigateTo\s*\(/.test(source)) score += 10;
  if (/await\s+t\.takeScreenshot\s*\(/.test(source)) score += 5;
  if (/await\s+t\.resizeWindow\s*\(/.test(source)) score += 5;
  if (/await\s+t\.wait\s*\(/.test(source)) score += 5;

  // TestCafe assertions
  if (/\bt\.expect\s*\(/.test(source)) score += 10;
  if (/Selector\s*\(/.test(source)) score += 10;

  // Negative: Cypress
  if (/\bcy\./.test(source)) score -= 30;
  // Negative: Playwright
  if (/from\s+['"]@playwright\/test['"]/.test(source)) score -= 30;
  if (/\bpage\.goto\s*\(/.test(source)) score -= 30;
  // Negative: WDIO
  if (/\bbrowser\.url\s*\(/.test(source)) score -= 20;
  // Negative: Puppeteer
  if (/\bpuppeteer\.launch/.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

function parse(source) {
  const lines = source.split('\n');
  const imports = [];
  const body = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    if (!trimmed) continue;

    if (trimmed.startsWith('//') || trimmed.startsWith('/*') || trimmed.startsWith('*')) {
      body.push(new Comment({ text: line, sourceLocation: loc, originalSource: line }));
      continue;
    }

    if (/^import\s/.test(trimmed)) {
      imports.push(new ImportStatement({ source: trimmed, sourceLocation: loc, originalSource: line, confidence: 'converted' }));
      continue;
    }

    // fixture`name` -> TestSuite
    if (/\bfixture\s*`/.test(trimmed)) {
      body.push(new TestSuite({ name: '', sourceLocation: loc, originalSource: line, confidence: 'converted' }));
      continue;
    }

    // test('name', async t => { -> TestCase
    if (/\btest\s*\(/.test(trimmed)) {
      body.push(new TestCase({ name: '', isAsync: true, sourceLocation: loc, originalSource: line, confidence: 'converted' }));
      continue;
    }

    // t.expect(...) -> Assertion
    if (/\bt\.expect\s*\(/.test(trimmed)) {
      body.push(new Assertion({ sourceLocation: loc, originalSource: line, confidence: 'converted' }));
      continue;
    }

    // Selector() -> RawCode (converted)
    if (/\bSelector\s*\(/.test(trimmed)) {
      body.push(new RawCode({ code: line, sourceLocation: loc, originalSource: line, confidence: 'converted' }));
      continue;
    }

    // t.* actions -> RawCode (converted)
    if (/\bt\./.test(trimmed)) {
      body.push(new RawCode({ code: line, sourceLocation: loc, originalSource: line, confidence: 'converted' }));
      continue;
    }

    body.push(new RawCode({ code: line, sourceLocation: loc, originalSource: line }));
  }

  return new TestFile({ language: 'javascript', imports, body });
}

/**
 * TestCafe is source-only. No conversions TO TestCafe are supported.
 */
function emit(_ir, source) {
  return source;
}

export default {
  name: 'testcafe',
  language: 'javascript',
  paradigm: 'bdd-e2e',
  detect,
  parse,
  emit,
  imports: {
    explicit: ['Selector', 'Role', 'ClientFunction', 'RequestMock'],
    from: 'testcafe',
    mockNamespace: null,
  },
};
