/**
 * Cypress framework definition.
 *
 * Provides detect, parse, and emit for the Cypress E2E testing framework.
 * parse() builds an IR tree from Cypress source code for scoring.
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  MockCall,
  ImportStatement,
  RawCode,
  Comment,
  Modifier,
} from "../../../core/ir.js";

function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  if (/\bcy\./.test(source)) score += 30;
  if (/\bcy\.visit\s*\(/.test(source)) score += 15;
  if (/\bcy\.get\s*\(/.test(source)) score += 15;
  if (/\bcy\.contains\s*\(/.test(source)) score += 10;
  if (/\bcy\.intercept\s*\(/.test(source)) score += 10;
  if (/\bcy\.request\s*\(/.test(source)) score += 5;
  if (/\.should\s*\(/.test(source)) score += 10;
  if (/\bCypress\./.test(source)) score += 10;
  if (/\bdescribe\s*\(/.test(source)) score += 3;
  if (/\bit\s*\(/.test(source)) score += 3;

  // Negative: Playwright
  if (/from\s+['"]@playwright\/test['"]/.test(source)) score -= 40;
  if (/\bpage\./.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

function parse(source) {
  const lines = source.split("\n");
  const imports = [];
  const body = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    if (!trimmed) continue;

    if (
      trimmed.startsWith("//") ||
      trimmed.startsWith("/*") ||
      trimmed.startsWith("*")
    ) {
      body.push(
        new Comment({ text: line, sourceLocation: loc, originalSource: line }),
      );
      continue;
    }

    if (/^import\s/.test(trimmed) || /^const\s.*=\s*require\(/.test(trimmed)) {
      imports.push(
        new ImportStatement({
          source: trimmed,
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    if (/\bdescribe\s*\(/.test(trimmed)) {
      body.push(
        new TestSuite({
          name: "",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    if (/\b(?:it|test)\s*\(/.test(trimmed)) {
      body.push(
        new TestCase({
          name: "",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    if (
      /\b(?:beforeEach|afterEach|beforeAll|afterAll|before|after)\s*\(/.test(
        trimmed,
      )
    ) {
      body.push(
        new Hook({
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    if (/\.should\s*\(/.test(trimmed) || /\bexpect\s*\(/.test(trimmed)) {
      body.push(
        new Assertion({
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    if (/\bcy\./.test(trimmed)) {
      body.push(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    body.push(
      new RawCode({ code: line, sourceLocation: loc, originalSource: line }),
    );
  }

  return new TestFile({ language: "javascript", imports, body });
}

function emit(_ir, _source) {
  throw new Error("Cypress emit not yet implemented (X→Cypress direction)");
}

export default {
  name: "cypress",
  language: "javascript",
  paradigm: "bdd-e2e",
  detect,
  parse,
  emit,
  imports: {
    globals: [
      "describe",
      "it",
      "context",
      "specify",
      "before",
      "after",
      "beforeEach",
      "afterEach",
      "cy",
      "Cypress",
      "expect",
    ],
    mockNamespace: "cy",
  },
};
