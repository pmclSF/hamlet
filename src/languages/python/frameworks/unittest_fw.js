/**
 * unittest framework definition.
 *
 * Provides detect, parse, and emit for Python's unittest framework.
 * parse() builds an IR tree from unittest source code for scoring.
 * emit() converts pytest source code to unittest by applying
 * regex-based transformation phases.
 *
 * File named unittest_fw.js to avoid Node reserved word issues.
 * Internal name is 'unittest'.
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
} from '../../../core/ir.js';
import { TodoFormatter } from '../../../core/TodoFormatter.js';

const formatter = new TodoFormatter('python');

/**
 * Detect whether source code is unittest.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong unittest signals
  if (/import\s+unittest\b/.test(source)) score += 30;
  if (/from\s+unittest\s+import\b/.test(source)) score += 30;
  if (/class\s+\w+\s*\(\s*unittest\.TestCase\s*\)/.test(source)) score += 30;
  if (/class\s+\w+\s*\(\s*TestCase\s*\)/.test(source)) score += 25;

  // unittest assertion methods
  if (/self\.assertEqual\s*\(/.test(source)) score += 15;
  if (/self\.assertTrue\s*\(/.test(source)) score += 15;
  if (/self\.assertFalse\s*\(/.test(source)) score += 15;
  if (/self\.assertRaises\s*\(/.test(source)) score += 15;
  if (/self\.assertIn\s*\(/.test(source)) score += 10;
  if (/self\.assertIsNone\s*\(/.test(source)) score += 10;
  if (/self\.assertIsNotNone\s*\(/.test(source)) score += 10;
  if (/self\.assertNotEqual\s*\(/.test(source)) score += 10;
  if (/self\.assertIsInstance\s*\(/.test(source)) score += 10;

  // unittest lifecycle
  if (/def\s+setUp\s*\(\s*self\s*\)/.test(source)) score += 10;
  if (/def\s+tearDown\s*\(\s*self\s*\)/.test(source)) score += 10;
  if (/def\s+setUpClass\s*\(\s*cls\s*\)/.test(source)) score += 10;
  if (/def\s+tearDownClass\s*\(\s*cls\s*\)/.test(source)) score += 10;

  // unittest decorators
  if (/@unittest\.skip\b/.test(source)) score += 10;
  if (/@unittest\.skipIf\b/.test(source)) score += 10;
  if (/@unittest\.skipUnless\b/.test(source)) score += 10;
  if (/@unittest\.expectedFailure\b/.test(source)) score += 10;

  // Weak signals
  if (/def\s+test_\w+\s*\(\s*self/.test(source)) score += 5;

  // Negative signals: NOT unittest
  if (/import\s+pytest\b/.test(source)) score -= 30;
  if (/@pytest\./.test(source)) score -= 30;
  if (/from\s+nose/.test(source)) score -= 30;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse unittest source code into an IR tree.
 */
function parse(source) {
  const lines = source.split('\n');
  const imports = [];
  const allNodes = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    if (!trimmed) continue;

    // Comments
    if (trimmed.startsWith('#')) {
      const isLicense =
        /license|copyright|MIT|Apache|BSD/i.test(trimmed) && i < 5;
      allNodes.push(
        new Comment({
          text: line,
          commentKind: isLicense ? 'license' : 'inline',
          preserveExact: isLicense,
          sourceLocation: loc,
          originalSource: line,
        })
      );
      continue;
    }

    // Import statements
    if (/^(?:import|from)\s/.test(trimmed)) {
      const sourceMatch = trimmed.match(
        /(?:from\s+(\S+)\s+import|import\s+(\S+))/
      );
      allNodes.push(
        new ImportStatement({
          kind: 'library',
          source: sourceMatch ? sourceMatch[1] || sourceMatch[2] : '',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      imports.push(allNodes[allNodes.length - 1]);
      continue;
    }

    // Class declaration
    if (/^\s*class\s+\w+/.test(trimmed)) {
      allNodes.push(
        new TestSuite({
          name: (trimmed.match(/class\s+(\w+)/) || [])[1] || '',
          modifiers: [],
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // setUp / tearDown
    if (/def\s+setUp\s*\(/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: 'beforeEach',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/def\s+tearDown\s*\(\s*self/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: 'afterEach',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/def\s+setUpClass\s*\(/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: 'beforeAll',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/def\s+tearDownClass\s*\(/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: 'afterAll',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // unittest decorators
    if (/@unittest\.skip\b/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: 'skip',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/@unittest\.skipIf\b/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: 'skip',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/@unittest\.skipUnless\b/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: 'skip',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/@unittest\.expectedFailure\b/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: 'expectedFailure',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // Test methods
    if (/def\s+test_\w+\s*\(/.test(trimmed)) {
      allNodes.push(
        new TestCase({
          name: (trimmed.match(/def\s+(test_\w+)\s*\(/) || [])[1] || '',
          isAsync: /async\s+def/.test(trimmed),
          modifiers: [],
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // unittest assertions
    if (/self\.assert\w+\s*\(/.test(trimmed)) {
      let kind = 'equal';
      if (/self\.assertEqual/.test(trimmed)) kind = 'equal';
      else if (/self\.assertNotEqual/.test(trimmed)) kind = 'notEqual';
      else if (/self\.assertTrue/.test(trimmed)) kind = 'truthy';
      else if (/self\.assertFalse/.test(trimmed)) kind = 'falsy';
      else if (/self\.assertIsNone/.test(trimmed)) kind = 'isNull';
      else if (/self\.assertIsNotNone/.test(trimmed)) kind = 'isDefined';
      else if (/self\.assertIn/.test(trimmed)) kind = 'contains';
      else if (/self\.assertNotIn/.test(trimmed)) kind = 'notContains';
      else if (/self\.assertRaises/.test(trimmed)) kind = 'throws';
      else if (/self\.assertIsInstance/.test(trimmed)) kind = 'isInstance';
      else if (/self\.assertGreater\b/.test(trimmed)) kind = 'greaterThan';
      else if (/self\.assertGreaterEqual/.test(trimmed))
        kind = 'greaterThanOrEqual';
      else if (/self\.assertLess\b/.test(trimmed)) kind = 'lessThan';
      else if (/self\.assertLessEqual/.test(trimmed)) kind = 'lessThanOrEqual';
      else if (/self\.assertAlmostEqual/.test(trimmed)) kind = 'closeTo';

      allNodes.push(
        new Assertion({
          kind,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // self.fail
    if (/self\.fail\s*\(/.test(trimmed)) {
      allNodes.push(
        new Assertion({
          kind: 'fail',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // Everything else
    allNodes.push(
      new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
      })
    );
  }

  return new TestFile({
    language: 'python',
    imports,
    body: allNodes.filter((n) => !imports.includes(n)),
  });
}

// ───────────────────────────────────────────────────────────
// Emit: pytest → unittest conversion
// ───────────────────────────────────────────────────────────

/**
 * Split arguments at depth 0, respecting nested parens, brackets, and strings.
 */
function splitArgs(argsStr) {
  const args = [];
  let depth = 0;
  let current = '';
  let inString = null;

  for (let i = 0; i < argsStr.length; i++) {
    const ch = argsStr[i];
    const prev = i > 0 ? argsStr[i - 1] : '';

    if (inString) {
      current += ch;
      if (ch === inString && prev !== '\\') inString = null;
      continue;
    }

    if (ch === '"' || ch === "'") {
      inString = ch;
      current += ch;
      continue;
    }

    if (ch === '(' || ch === '[' || ch === '{') {
      depth++;
      current += ch;
      continue;
    }

    if (ch === ')' || ch === ']' || ch === '}') {
      depth--;
      current += ch;
      continue;
    }

    if (ch === ',' && depth === 0) {
      args.push(current.trim());
      current = '';
      continue;
    }

    current += ch;
  }

  if (current.trim()) args.push(current.trim());
  return args;
}

/**
 * Extract the content between the outermost parentheses.
 */
function extractParenContent(str, startIndex) {
  let depth = 0;
  let start = -1;

  for (let i = startIndex; i < str.length; i++) {
    if (str[i] === '(') {
      if (depth === 0) start = i + 1;
      depth++;
    } else if (str[i] === ')') {
      depth--;
      if (depth === 0) {
        return { content: str.substring(start, i), end: i };
      }
    }
  }
  return null;
}

/**
 * Detect if source is pytest code.
 */
function isPytestSource(source) {
  if (/import\s+pytest\b/.test(source)) return true;
  if (/@pytest\./.test(source)) return true;
  // Bare test functions without self and with bare assert
  if (
    /^def\s+test_\w+\s*\([^)]*\)\s*:/m.test(source) &&
    !/def\s+test_\w+\s*\(\s*self/.test(source)
  ) {
    if (/^\s+assert\s+/m.test(source)) return true;
  }
  return false;
}

/**
 * Convert pytest assertions to unittest self.assert* calls.
 */
function convertAssertions(result) {
  // Pre-pass: join multi-line with pytest.raises( into single line
  // Handle 1-arg: with pytest.raises(\n    Exc\n):
  result = result.replace(
    /^(\s*with\s+pytest\.raises\s*\()[ \t]*\n\s*(.+?)\n\s*(\)\s*(?:as\s+\w+\s*)?:)/gm,
    '$1$2$3'
  );
  // Handle multi-arg: with pytest.raises(\n    Exc, match="pattern"\n):
  result = result.replace(
    /^(\s*with\s+pytest\.raises\s*\()[ \t]*\n\s*(.+?)(?:,\s*(.+?))?\n\s*(\)\s*(?:as\s+\w+\s*)?:)/gm,
    (match, prefix, arg1, arg2, suffix) => {
      return arg2
        ? `${prefix}${arg1}, ${arg2}${suffix}`
        : `${prefix}${arg1}${suffix}`;
    }
  );

  const lines = result.split('\n');
  const converted = [];

  for (const line of lines) {
    const trimmed = line.trim();

    // Skip non-assert lines
    if (!trimmed.startsWith('assert ') && !trimmed.startsWith('assert(')) {
      // Handle pytest.raises context manager
      if (/with\s+pytest\.raises\s*\(/.test(trimmed)) {
        converted.push(line.replace(/pytest\.raises/, 'self.assertRaises'));
        continue;
      }
      converted.push(line);
      continue;
    }

    const indent = line.match(/^(\s*)/)[1];
    let expr = trimmed.replace(/^assert\s+/, '');

    // Separate assertion message: assert X == Y, "msg" → expr=X == Y
    let assertMsg = '';
    const topParts = splitArgs(expr);
    if (topParts.length > 1) {
      const last = topParts[topParts.length - 1].trim();
      if (/^["'(f"]/.test(last)) {
        assertMsg = last;
        expr = topParts.slice(0, -1).join(', ');
      }
    }
    const msgAnnotation = assertMsg ? `  # @hamlet:msg(${assertMsg})` : '';

    // Strip string contents for safe operator detection (avoid matching > inside strings)
    const safe = expr.replace(/(["'])(?:(?!\1|\\).|\\.)*\1/g, (_m, q) => q + q);

    // Order matters: most specific patterns first

    // assert x is not None
    if (/^(.+?)\s+is\s+not\s+None$/.test(safe)) {
      const isNotNoneMatch = expr.match(/^(.+?)\s+is\s+not\s+None$/);
      if (isNotNoneMatch) {
        converted.push(
          `${indent}self.assertIsNotNone(${isNotNoneMatch[1]})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert x is None
    if (/^(.+?)\s+is\s+None$/.test(safe)) {
      const isNoneMatch = expr.match(/^(.+?)\s+is\s+None$/);
      if (isNoneMatch) {
        converted.push(
          `${indent}self.assertIsNone(${isNoneMatch[1]})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert isinstance(x, Y)
    const isInstanceMatch = expr.match(/^isinstance\((.+)\)$/);
    if (isInstanceMatch) {
      converted.push(
        `${indent}self.assertIsInstance(${isInstanceMatch[1]})${msgAnnotation}`
      );
      continue;
    }

    // assert not x
    const notMatch = expr.match(/^not\s+(.+)$/);
    if (notMatch) {
      converted.push(
        `${indent}self.assertFalse(${notMatch[1]})${msgAnnotation}`
      );
      continue;
    }

    // assert x not in y (greedy left to skip 'not in' inside strings)
    if (/^(.+)\s+not\s+in\s+(.+)$/.test(safe)) {
      const notInMatch = expr.match(/^(.+)\s+not\s+in\s+(.+)$/);
      if (notInMatch) {
        converted.push(
          `${indent}self.assertNotIn(${notInMatch[1]}, ${notInMatch[2]})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert x in y (greedy left to skip 'in' inside strings)
    if (/^(.+)\s+in\s+(.+)$/.test(safe)) {
      const inMatch = expr.match(/^(.+)\s+in\s+(.+)$/);
      if (inMatch) {
        converted.push(
          `${indent}self.assertIn(${inMatch[1]}, ${inMatch[2]})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert a == b (match against safe to avoid operators inside strings)
    if (/^(.+?)\s*==\s*(.+)$/.test(safe)) {
      const eqMatch = expr.match(/^(.+?)\s*==\s*(.+)$/);
      if (eqMatch) {
        converted.push(
          `${indent}self.assertEqual(${eqMatch[1].trim()}, ${eqMatch[2].trim()})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert a != b
    if (/^(.+?)\s*!=\s*(.+)$/.test(safe)) {
      const neqMatch = expr.match(/^(.+?)\s*!=\s*(.+)$/);
      if (neqMatch) {
        converted.push(
          `${indent}self.assertNotEqual(${neqMatch[1].trim()}, ${neqMatch[2].trim()})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert a >= b
    if (/^(.+?)\s*>=\s*(.+)$/.test(safe)) {
      const gteMatch = expr.match(/^(.+?)\s*>=\s*(.+)$/);
      if (gteMatch) {
        converted.push(
          `${indent}self.assertGreaterEqual(${gteMatch[1].trim()}, ${gteMatch[2].trim()})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert a <= b
    if (/^(.+?)\s*<=\s*(.+)$/.test(safe)) {
      const lteMatch = expr.match(/^(.+?)\s*<=\s*(.+)$/);
      if (lteMatch) {
        converted.push(
          `${indent}self.assertLessEqual(${lteMatch[1].trim()}, ${lteMatch[2].trim()})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert a > b
    if (/^(.+?)\s*>\s*(.+)$/.test(safe)) {
      const gtMatch = expr.match(/^(.+?)\s*>\s*(.+)$/);
      if (gtMatch) {
        converted.push(
          `${indent}self.assertGreater(${gtMatch[1].trim()}, ${gtMatch[2].trim()})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert a < b
    if (/^(.+?)\s*<\s*(.+)$/.test(safe)) {
      const ltMatch = expr.match(/^(.+?)\s*<\s*(.+)$/);
      if (ltMatch) {
        converted.push(
          `${indent}self.assertLess(${ltMatch[1].trim()}, ${ltMatch[2].trim()})${msgAnnotation}`
        );
        continue;
      }
    }

    // assert x (simple truthy)
    converted.push(`${indent}self.assertTrue(${expr})${msgAnnotation}`);
  }

  return converted.join('\n');
}

/**
 * Wrap bare test functions in a unittest.TestCase class.
 */
function wrapInClass(result) {
  const lines = result.split('\n');
  const output = [];
  let hasClass = false;

  // Check if already has a class
  for (const line of lines) {
    if (/^\s*class\s+\w+/.test(line)) {
      hasClass = true;
      break;
    }
  }
  if (hasClass) return result;

  // Derive class name from first test function
  let className = 'TestConversion';
  for (const line of lines) {
    const testMatch = line.match(/^def\s+(test_\w+)\s*\(/);
    if (testMatch) {
      const parts = testMatch[1].replace(/^test_/, '').split('_');
      className =
        'Test' +
        parts.map((p) => p.charAt(0).toUpperCase() + p.slice(1)).join('');
      break;
    }
  }

  // Separate: imports/top-level, non-test functions, test functions
  const preContent = [];
  const testFunctions = [];
  let i = 0;

  // Gather imports and non-function top-level content, plus non-test functions
  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trim();

    // Stop at test functions, setUp/tearDown, or pytest decorators
    if (/^(?:@pytest\.|def\s+test_|def\s+setUp|def\s+tearDown)/.test(trimmed)) {
      break;
    }

    // Non-test functions: keep at module level (collect entire function body)
    if (/^def\s+(?!test_|setUp|tearDown)\w+/.test(trimmed)) {
      preContent.push(line);
      i++;
      while (i < lines.length) {
        // A non-empty line at column 0 means we've left the function body
        if (lines[i].trim() !== '' && /^\S/.test(lines[i])) break;
        preContent.push(lines[i]);
        i++;
      }
      continue;
    }

    preContent.push(line);
    i++;
  }

  // Everything remaining is test functions (and their bodies)
  while (i < lines.length) {
    testFunctions.push(lines[i]);
    i++;
  }

  if (testFunctions.length === 0) return result;

  // Build output
  for (const line of preContent) {
    output.push(line);
  }

  if (
    preContent.length > 0 &&
    preContent[preContent.length - 1].trim() !== ''
  ) {
    output.push('');
  }
  output.push('');
  output.push(`class ${className}(unittest.TestCase):`);

  for (const line of testFunctions) {
    if (!line.trim()) {
      output.push('');
      continue;
    }

    let indented = '    ' + line;

    // Add self parameter to function definitions
    const defMatch = indented.match(/^(\s*def\s+\w+\s*)\(\s*\)/);
    if (defMatch) {
      indented = indented.replace(/\(\s*\)/, '(self)');
    } else {
      const defParamMatch = indented.match(/^(\s*def\s+\w+\s*)\((.+)\)/);
      if (defParamMatch && !/\bself\b/.test(defParamMatch[2])) {
        indented = indented.replace(/\((.+)\)/, '(self, $1)');
      }
    }

    output.push(indented);
  }

  return output.join('\n');
}

/**
 * Convert pytest markers to unittest decorators.
 */
function convertMarkers(result) {
  // @pytest.mark.skip -> @unittest.skip
  result = result.replace(/@pytest\.mark\.skip\b(?!\s*\()/g, '@unittest.skip');

  // @pytest.mark.skip(reason="...") -> @unittest.skip("...")
  result = result.replace(
    /@pytest\.mark\.skip\(\s*reason\s*=\s*("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')\s*\)/g,
    '@unittest.skip($1)'
  );

  // @pytest.mark.skip("...") -> @unittest.skip("...")
  result = result.replace(
    /@pytest\.mark\.skip\(("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')\)/g,
    '@unittest.skip($1)'
  );

  // @pytest.mark.skipif(cond, reason="...") -> @unittest.skipIf(cond, "...")
  result = result.replace(
    /@pytest\.mark\.skipif\(([^,]+),\s*reason\s*=\s*("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')\s*\)/g,
    '@unittest.skipIf($1, $2)'
  );

  // @pytest.mark.xfail(args) -> @unittest.expectedFailure  # @hamlet:xfail(args)
  result = result.replace(
    /@pytest\.mark\.xfail\(([^)]*)\)/g,
    '@unittest.expectedFailure  # @hamlet:xfail($1)'
  );

  // @pytest.mark.xfail (bare) -> @unittest.expectedFailure
  result = result.replace(
    /@pytest\.mark\.xfail\b(?!\s*\()/g,
    '@unittest.expectedFailure'
  );

  return result;
}

/**
 * Convert pytest fixtures to setUp/tearDown.
 */
function convertFixtures(result) {
  const lines = result.split('\n');
  const output = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trim();

    // Match @pytest.fixture(autouse=True)
    if (/^(\s*)@pytest\.fixture\s*\(\s*autouse\s*=\s*True\s*\)/.test(line)) {
      const indent = line.match(/^(\s*)/)[1];
      i++;

      // Next line should be the function def
      if (i < lines.length && /def\s+\w+\s*\(/.test(lines[i].trim())) {
        const funcLine = lines[i];
        i++;

        // Collect the function body
        const bodyLines = [];
        const funcIndent = indent + '    ';
        while (i < lines.length) {
          const bodyLine = lines[i];
          if (
            bodyLine.trim() === '' ||
            bodyLine.startsWith(funcIndent) ||
            /^\s+/.test(bodyLine)
          ) {
            if (
              bodyLine.trim() !== '' &&
              !bodyLine.startsWith(funcIndent) &&
              bodyLine.trim().length > 0
            ) {
              // Check if this line is at a lesser indent (end of function)
              const lineIndent = bodyLine.match(/^(\s*)/)[1];
              if (lineIndent.length <= indent.length && bodyLine.trim() !== '')
                break;
            }
            bodyLines.push(bodyLine);
            i++;
          } else {
            break;
          }
        }

        // Check for yield in body
        const yieldIdx = bodyLines.findIndex(
          (l) => l.trim() === 'yield' || l.trim().startsWith('yield ')
        );

        if (yieldIdx >= 0) {
          // Split into setUp (pre-yield) and tearDown (post-yield)
          const setUpBody = bodyLines.slice(0, yieldIdx);
          const tearDownBody = bodyLines
            .slice(yieldIdx + 1)
            .filter((l) => l.trim() !== '');

          output.push(`${indent}def setUp(self):`);
          if (
            setUpBody.length === 0 ||
            setUpBody.every((l) => l.trim() === '')
          ) {
            output.push(`${indent}    pass`);
          } else {
            for (const bl of setUpBody) {
              output.push(bl);
            }
          }

          if (tearDownBody.length > 0) {
            output.push('');
            output.push(`${indent}def tearDown(self):`);
            for (const bl of tearDownBody) {
              output.push(bl);
            }
          }
        } else {
          // No yield — just setUp
          output.push(`${indent}def setUp(self):`);
          if (
            bodyLines.length === 0 ||
            bodyLines.every((l) => l.trim() === '')
          ) {
            output.push(`${indent}    pass`);
          } else {
            for (const bl of bodyLines) {
              output.push(bl);
            }
          }
        }
        continue;
      }
    }

    // Mark non-autouse fixtures as unconvertible
    if (
      /^(\s*)@pytest\.fixture\b/.test(line) &&
      !/autouse\s*=\s*True/.test(line)
    ) {
      const indent = line.match(/^(\s*)/)[1];
      output.push(
        `${indent}${formatter
          .formatTodo({
            id: 'UNCONVERTIBLE-FIXTURE',
            description:
              'pytest fixture without autouse=True has no direct unittest equivalent',
            original: trimmed,
            action:
              'Manually convert this fixture to setUp/tearDown or pass the value directly',
          })
          .split('\n')
          .join(`\n${indent}`)}`
      );
      output.push(line);
      i++;
      continue;
    }

    output.push(line);
    i++;
  }

  return output.join('\n');
}

/**
 * Mark unconvertible pytest patterns.
 */
function markUnconvertible(result) {
  // monkeypatch fixture usage
  result = result.replace(
    /^(\s*)(.*\bmonkeypatch\b.*)$/gm,
    (match, indent, code) => {
      if (code.trim().startsWith('#')) return match;
      return `${indent}${formatter
        .formatTodo({
          id: 'UNCONVERTIBLE-MONKEYPATCH',
          description: 'monkeypatch fixture has no direct unittest equivalent',
          original: code.trim(),
          action: 'Use unittest.mock.patch() or similar mocking approach',
        })
        .split('\n')
        .join(`\n${indent}`)}\n${match}`;
    }
  );

  // capfd/capsys/caplog fixture usage
  result = result.replace(
    /^(\s*)(.*\b(?:capfd|capsys|caplog)\b.*)$/gm,
    (match, indent, code) => {
      if (code.trim().startsWith('#')) return match;
      return `${indent}${formatter
        .formatTodo({
          id: 'UNCONVERTIBLE-CAPTURE',
          description:
            'pytest capture fixtures have no direct unittest equivalent',
          original: code.trim(),
          action:
            'Use contextlib.redirect_stdout or unittest.mock to capture output',
        })
        .split('\n')
        .join(`\n${indent}`)}\n${match}`;
    }
  );

  // tmp_path fixture usage
  result = result.replace(
    /^(\s*)(.*\btmp_path\b.*)$/gm,
    (match, indent, code) => {
      if (code.trim().startsWith('#')) return match;
      return `${indent}${formatter
        .formatTodo({
          id: 'UNCONVERTIBLE-TMPPATH',
          description: 'tmp_path fixture has no direct unittest equivalent',
          original: code.trim(),
          action: 'Use tempfile.mkdtemp() in setUp/tearDown',
        })
        .split('\n')
        .join(`\n${indent}`)}\n${match}`;
    }
  );

  return result;
}

/**
 * Emit unittest code from IR + original source.
 * Converts pytest source to unittest.
 */
function emit(_ir, source) {
  if (!isPytestSource(source)) {
    return source;
  }

  let result = source;

  // Phase 1: Add import unittest (if not already present)
  if (!/import\s+unittest\b/.test(result)) {
    // Add after any existing imports, or at the top
    const importMatch = result.match(/^(?:(?:import|from)\s+.+\n)+/m);
    if (importMatch) {
      const lastImportEnd = importMatch.index + importMatch[0].length;
      result =
        result.substring(0, lastImportEnd) +
        'import unittest\n' +
        result.substring(lastImportEnd);
    } else {
      result = 'import unittest\n\n' + result;
    }
  }

  // Phase 2: Convert class wrapping (bare functions -> TestCase class)
  result = wrapInClass(result);

  // Phase 3: Convert fixtures
  result = convertFixtures(result);

  // Phase 4: Convert assertions
  result = convertAssertions(result);

  // Phase 5: Convert markers
  result = convertMarkers(result);

  // Phase 6: @pytest.mark.parametrize -> HAMLET-TODO (complex conversion)
  result = result.replace(
    /^(\s*)@pytest\.mark\.parametrize\((.+)\)$/gm,
    (match, indent, args) => {
      return `${indent}${formatter
        .formatTodo({
          id: 'UNCONVERTIBLE-PARAMETRIZE',
          description:
            '@pytest.mark.parametrize has no direct unittest equivalent',
          original: match.trim(),
          action:
            'Use subTest() or create individual test methods for each parameter set',
        })
        .split('\n')
        .join(`\n${indent}`)}\n${match}`;
    }
  );

  // Phase 7: Remove import pytest (if no remaining references)
  if (!/pytest\./.test(result.replace(/#.*$/gm, ''))) {
    result = result.replace(/^import\s+pytest\s*\n/gm, '');
  }

  // Phase 8: Cleanup
  // Remove multiple consecutive blank lines (max 2)
  result = result.replace(/\n{4,}/g, '\n\n\n');

  // Ensure file ends with newline
  if (!result.endsWith('\n')) {
    result += '\n';
  }

  // Phase 9: Mark unconvertible patterns
  result = markUnconvertible(result);

  return result;
}

export default {
  name: 'unittest',
  language: 'python',
  paradigm: 'xunit',
  detect,
  parse,
  emit,
  imports: {
    packages: ['unittest'],
  },
};
