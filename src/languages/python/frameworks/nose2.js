/**
 * nose2 framework definition.
 *
 * Provides detect, parse, and emit for the nose2 testing framework.
 * parse() builds an IR tree from nose2 source code for scoring.
 * emit() is a stub — nose2 is only used as a source framework.
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
  Modifier,
} from "../../../core/ir.js";

/**
 * Detect whether source code is nose2.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong nose signals
  if (/from\s+nose\.tools\s+import\b/.test(source)) score += 30;
  if (/from\s+nose2\.tools\s+import\b/.test(source)) score += 30;
  if (/import\s+nose2\b/.test(source)) score += 30;
  if (/import\s+nose\b/.test(source)) score += 25;

  // nose-specific assertion functions
  if (/\bassert_equal\s*\(/.test(source)) score += 15;
  if (/\bassert_true\s*\(/.test(source)) score += 15;
  if (/\bassert_false\s*\(/.test(source)) score += 15;
  if (/\bassert_raises\s*\(/.test(source)) score += 15;
  if (/\bassert_in\s*\(/.test(source)) score += 15;
  if (/\bassert_not_equal\s*\(/.test(source)) score += 15;
  if (/\bassert_is_none\s*\(/.test(source)) score += 15;
  if (/\bassert_is_not_none\s*\(/.test(source)) score += 15;
  if (/\bassert_is_instance\s*\(/.test(source)) score += 15;

  // nose2-specific decorators
  if (/@params\s*\(/.test(source)) score += 15;
  if (/@attr\s*\(/.test(source)) score += 10;
  if (/@such\.it\b/.test(source)) score += 15;

  // Weak signals (shared with other frameworks)
  if (/def\s+test_\w+/.test(source)) score += 5;
  if (/class\s+\w+.*TestCase/.test(source)) score += 5;

  // Negative signals: NOT nose
  if (/import\s+pytest\b/.test(source)) score -= 30;
  if (/@pytest\./.test(source)) score -= 30;
  if (/import\s+unittest\b/.test(source) && !/import\s+nose/.test(source))
    score -= 20;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse nose2 source code into an IR tree.
 */
function parse(source) {
  const lines = source.split("\n");
  const imports = [];
  const allNodes = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    if (!trimmed) continue;

    // Comments
    if (trimmed.startsWith("#")) {
      const isLicense =
        /license|copyright|MIT|Apache|BSD/i.test(trimmed) && i < 5;
      allNodes.push(
        new Comment({
          text: line,
          commentKind: isLicense ? "license" : "inline",
          preserveExact: isLicense,
          sourceLocation: loc,
          originalSource: line,
        }),
      );
      continue;
    }

    // Import statements
    if (/^(?:import|from)\s/.test(trimmed)) {
      const sourceMatch = trimmed.match(
        /(?:from\s+(\S+)\s+import|import\s+(\S+))/,
      );
      allNodes.push(
        new ImportStatement({
          kind: "library",
          source: sourceMatch ? sourceMatch[1] || sourceMatch[2] : "",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      imports.push(allNodes[allNodes.length - 1]);
      continue;
    }

    // Class declaration
    if (/^\s*class\s+\w+/.test(trimmed)) {
      allNodes.push(
        new TestSuite({
          name: (trimmed.match(/class\s+(\w+)/) || [])[1] || "",
          modifiers: [],
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // setUp / tearDown
    if (/def\s+setUp\s*\(/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: "beforeEach",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }
    if (/def\s+tearDown\s*\(/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: "afterEach",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }
    if (/def\s+setUpClass\s*\(/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: "beforeAll",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }
    if (/def\s+tearDownClass\s*\(/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: "afterAll",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // @params decorator
    if (/@params\s*\(/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: "parameterized",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // @attr decorator
    if (/@attr\s*\(/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: "tag",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // Test functions/methods
    if (/def\s+test_\w+\s*\(/.test(trimmed)) {
      allNodes.push(
        new TestCase({
          name: (trimmed.match(/def\s+(test_\w+)\s*\(/) || [])[1] || "",
          isAsync: /async\s+def/.test(trimmed),
          modifiers: [],
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // nose assertion functions
    if (/\bassert_\w+\s*\(/.test(trimmed)) {
      let kind = "equal";
      if (/assert_equal/.test(trimmed)) kind = "equal";
      else if (/assert_not_equal/.test(trimmed)) kind = "notEqual";
      else if (/assert_true/.test(trimmed)) kind = "truthy";
      else if (/assert_false/.test(trimmed)) kind = "falsy";
      else if (/assert_is_none/.test(trimmed)) kind = "isNull";
      else if (/assert_is_not_none/.test(trimmed)) kind = "isDefined";
      else if (/assert_in/.test(trimmed)) kind = "contains";
      else if (/assert_not_in/.test(trimmed)) kind = "notContains";
      else if (/assert_raises/.test(trimmed)) kind = "throws";
      else if (/assert_is_instance/.test(trimmed)) kind = "isInstance";

      allNodes.push(
        new Assertion({
          kind,
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // Everything else
    allNodes.push(
      new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
      }),
    );
  }

  return new TestFile({
    language: "python",
    imports,
    body: allNodes.filter((n) => !imports.includes(n)),
  });
}

/**
 * Emit nose2 code from IR + original source.
 *
 * Stub — nose2 is only used as a source framework.
 */
function emit(_ir, source) {
  return source;
}

export default {
  name: "nose2",
  language: "python",
  paradigm: "xunit",
  detect,
  parse,
  emit,
  imports: {
    packages: ["nose2", "nose.tools"],
  },
};
