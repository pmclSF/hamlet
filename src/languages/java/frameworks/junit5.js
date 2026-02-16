/**
 * JUnit 5 framework definition.
 *
 * Provides detect, parse, and emit for the JUnit 5 (Jupiter) testing framework.
 * parse() builds an IR tree from JUnit 5 source code for scoring.
 * emit() converts JUnit 4 or TestNG source code to JUnit 5 by applying
 * source-framework-specific regex phases.
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
import { TodoFormatter } from "../../../core/TodoFormatter.js";

const formatter = new TodoFormatter("java");

/**
 * Detect whether source code is JUnit 5.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong JUnit 5 signals
  if (/import\s+org\.junit\.jupiter/.test(source)) score += 40;
  if (/@BeforeEach\b/.test(source)) score += 15;
  if (/@AfterEach\b/.test(source)) score += 15;
  if (/@BeforeAll\b/.test(source)) score += 15;
  if (/@AfterAll\b/.test(source)) score += 15;
  if (/@DisplayName\b/.test(source)) score += 20;
  if (/@Nested\b/.test(source)) score += 20;
  if (/Assertions\.assertEquals\b/.test(source)) score += 15;
  if (/assertThrows\s*\(/.test(source)) score += 15;
  if (/assertTimeout\s*\(/.test(source)) score += 10;
  if (/@ParameterizedTest\b/.test(source)) score += 15;
  if (/@ExtendWith\b/.test(source)) score += 15;
  if (/@Tag\s*\(/.test(source)) score += 10;
  if (/@Disabled\b/.test(source)) score += 10;

  // Weak signals
  if (/@Test\b/.test(source)) score += 5;

  // Negative signals: NOT JUnit 5
  if (
    /import\s+org\.junit\.Test\b/.test(source) &&
    !/import\s+org\.junit\.jupiter/.test(source)
  )
    score -= 30;
  if (/import\s+org\.testng/.test(source)) score -= 40;
  if (/@RunWith\b/.test(source)) score -= 20;
  if (/@Rule\b/.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse JUnit 5 source code into an IR tree.
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
    if (
      trimmed.startsWith("//") ||
      trimmed.startsWith("/*") ||
      trimmed.startsWith("*")
    ) {
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
    if (/^import\s/.test(trimmed)) {
      const sourceMatch = trimmed.match(/import\s+(?:static\s+)?([^\s;]+)/);
      allNodes.push(
        new ImportStatement({
          kind: "library",
          source: sourceMatch ? sourceMatch[1] : "",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      imports.push(allNodes[allNodes.length - 1]);
      continue;
    }

    // Class declaration
    if (/\bclass\s+\w+/.test(trimmed)) {
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

    // @Test annotation
    if (/@Test\b/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: "test",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // Lifecycle annotations
    if (/@BeforeEach\b/.test(trimmed)) {
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
    if (/@AfterEach\b/.test(trimmed)) {
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
    if (/@BeforeAll\b/.test(trimmed)) {
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
    if (/@AfterAll\b/.test(trimmed)) {
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

    // @Disabled
    if (/@Disabled\b/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: "skip",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // @DisplayName
    if (/@DisplayName\b/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: "displayName",
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // Test methods
    if (/(?:public\s+|protected\s+|private\s+)?void\s+\w+\s*\(/.test(trimmed)) {
      allNodes.push(
        new TestCase({
          name: (trimmed.match(/void\s+(\w+)\s*\(/) || [])[1] || "",
          isAsync: false,
          modifiers: [],
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // Assertions
    if (
      /\bAssertions\.\w+\s*\(/.test(trimmed) ||
      /\bassertThrows\s*\(/.test(trimmed) ||
      /\bassertTimeout\s*\(/.test(trimmed)
    ) {
      let kind = "equal";
      if (/assertEquals/.test(trimmed)) kind = "equal";
      else if (/assertTrue/.test(trimmed)) kind = "truthy";
      else if (/assertFalse/.test(trimmed)) kind = "falsy";
      else if (/assertNull/.test(trimmed)) kind = "isNull";
      else if (/assertNotNull/.test(trimmed)) kind = "isDefined";
      else if (/assertThrows/.test(trimmed)) kind = "throws";
      else if (/assertTimeout/.test(trimmed)) kind = "timeout";

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
    language: "java",
    imports,
    body: allNodes.filter((n) => !imports.includes(n)),
  });
}

/**
 * Extract the method body from source code starting at a given method signature.
 * Uses brace-counting that skips braces inside string literals and line comments.
 *
 * @param {string[]} lines - Source lines
 * @param {number} startIdx - Index of the line containing the method signature
 * @returns {{ bodyLines: string[], startLine: number, endLine: number } | null}
 */
function extractMethodBody(lines, startIdx) {
  let depth = 0;
  let foundOpen = false;
  let bodyStart = -1;
  let bodyEnd = -1;

  for (let i = startIdx; i < lines.length; i++) {
    const line = lines[i];
    let inString = false;
    let stringChar = "";
    let inLineComment = false;

    for (let j = 0; j < line.length; j++) {
      const ch = line[j];
      const next = j + 1 < line.length ? line[j + 1] : "";

      if (inLineComment) break;

      if (inString) {
        if (ch === "\\") {
          j++;
          continue;
        }
        if (ch === stringChar) inString = false;
        continue;
      }

      if (ch === "/" && next === "/") {
        inLineComment = true;
        continue;
      }
      if (ch === '"' || ch === "'") {
        inString = true;
        stringChar = ch;
        continue;
      }

      if (ch === "{") {
        if (!foundOpen) {
          foundOpen = true;
          bodyStart = i;
        }
        depth++;
      } else if (ch === "}") {
        depth--;
        if (depth === 0 && foundOpen) {
          bodyEnd = i;
          return {
            bodyLines: lines.slice(bodyStart, bodyEnd + 1),
            startLine: bodyStart,
            endLine: bodyEnd,
          };
        }
      }
    }
  }

  return null;
}

/**
 * Split a Java argument list, respecting nested parentheses, strings, and generics.
 *
 * @param {string} argStr - The argument string (without outer parens)
 * @returns {string[]} Array of argument strings
 */
function splitArgs(argStr) {
  const args = [];
  let depth = 0;
  let current = "";
  let inString = false;
  let stringChar = "";

  for (let i = 0; i < argStr.length; i++) {
    const ch = argStr[i];

    if (inString) {
      current += ch;
      if (ch === "\\") {
        i++;
        current += argStr[i] || "";
        continue;
      }
      if (ch === stringChar) inString = false;
      continue;
    }

    if (ch === '"' || ch === "'") {
      inString = true;
      stringChar = ch;
      current += ch;
      continue;
    }

    if (ch === "(" || ch === "<" || ch === "{" || ch === "[") {
      depth++;
      current += ch;
    } else if (ch === ")" || ch === ">" || ch === "}" || ch === "]") {
      depth--;
      current += ch;
    } else if (ch === "," && depth === 0) {
      args.push(current.trim());
      current = "";
    } else {
      current += ch;
    }
  }

  if (current.trim()) args.push(current.trim());
  return args;
}

/**
 * Reorder assertion arguments: move first arg (message) to last position.
 * Only when there are exactly `expectedArity` args for message-first forms.
 *
 * Handles: Assert.method("msg", arg1, arg2) → Assertions.method(arg1, arg2, "msg")
 * And:     Assert.method("msg", arg1) → Assertions.method(arg1, "msg")
 *
 * @param {string} source - Source code
 * @param {string} methodName - e.g., 'assertEquals'
 * @param {number} arityWithMessage - Number of args that indicates message is present (3 for assertEquals, 2 for assertTrue)
 * @returns {string}
 */
function reorderAssertionArgs(source, methodName, arityWithMessage) {
  const pattern = new RegExp(`Assertions\\.${methodName}\\(`, "g");

  let result = "";
  let lastIndex = 0;
  let match;

  while ((match = pattern.exec(source)) !== null) {
    const start = match.index + match[0].length;
    // Find matching closing paren
    let depth = 1;
    let pos = start;
    let inStr = false;
    let strCh = "";

    while (pos < source.length && depth > 0) {
      const ch = source[pos];
      if (inStr) {
        if (ch === "\\") {
          pos++;
        } else if (ch === strCh) {
          inStr = false;
        }
      } else {
        if (ch === '"' || ch === "'") {
          inStr = true;
          strCh = ch;
        } else if (ch === "(") {
          depth++;
        } else if (ch === ")") {
          depth--;
        }
      }
      if (depth > 0) pos++;
    }

    const argsStr = source.substring(start, pos);
    const args = splitArgs(argsStr);

    if (args.length === arityWithMessage) {
      // Move first arg to last
      const message = args[0];
      const rest = args.slice(1);
      result += source.substring(lastIndex, match.index);
      result += `Assertions.${methodName}(${[...rest, message].join(", ")})`;
      lastIndex = pos + 1;
    }
    // If arity doesn't match, leave unchanged
  }

  result += source.substring(lastIndex);
  return result || source;
}

/**
 * Emit JUnit 5 code from IR + original source.
 *
 * Converts JUnit 4 or TestNG source code to JUnit 5 by applying
 * source-framework-specific regex phases. Each phase only matches
 * patterns from its source framework.
 *
 * @param {TestFile} _ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original source code
 * @returns {string} Converted JUnit 5 source code
 */
function emit(_ir, source) {
  let result = source;

  // Detect source framework to apply only relevant phases
  const isJUnit4Source =
    /import\s+org\.junit\./.test(source) &&
    !/import\s+org\.junit\.jupiter/.test(source);
  const isTestNGSource = /import\s+org\.testng/.test(source);

  // ========================================
  // JUnit 4 → JUnit 5 Phases (1-7)
  // ========================================

  if (isJUnit4Source) {
    // --- Phase 1: JUnit 4 annotation renames ---

    // @Before → @BeforeEach (negative lookahead for @BeforeClass/@BeforeEach/@BeforeAll/@BeforeMethod)
    result = result.replace(
      /@Before\b(?!Class|Each|All|Method)/g,
      "@BeforeEach",
    );

    // @After → @AfterEach (negative lookahead for @AfterClass/@AfterEach/@AfterAll/@AfterMethod)
    result = result.replace(/@After\b(?!Class|Each|All|Method)/g, "@AfterEach");

    // @BeforeClass → @BeforeAll
    result = result.replace(/@BeforeClass\b/g, "@BeforeAll");

    // @AfterClass → @AfterAll
    result = result.replace(/@AfterClass\b/g, "@AfterAll");

    // @Ignore → @Disabled
    result = result.replace(/@Ignore\b/g, "@Disabled");

    // @RunWith(X.class) → @ExtendWith(X.class)
    result = result.replace(/@RunWith\s*\(/g, "@ExtendWith(");

    // @Category(X.class) → @Tag("X")
    result = result.replace(
      /@Category\s*\(\s*(\w+)\.class\s*\)/g,
      '@Tag("$1")',
    );

    // --- Phase 2: JUnit 4 import rewrites ---

    result = result.replace(
      /import\s+org\.junit\.Test\s*;/g,
      "import org.junit.jupiter.api.Test;",
    );
    result = result.replace(
      /import\s+org\.junit\.Before\s*;/g,
      "import org.junit.jupiter.api.BeforeEach;",
    );
    result = result.replace(
      /import\s+org\.junit\.After\s*;/g,
      "import org.junit.jupiter.api.AfterEach;",
    );
    result = result.replace(
      /import\s+org\.junit\.BeforeClass\s*;/g,
      "import org.junit.jupiter.api.BeforeAll;",
    );
    result = result.replace(
      /import\s+org\.junit\.AfterClass\s*;/g,
      "import org.junit.jupiter.api.AfterAll;",
    );
    result = result.replace(
      /import\s+org\.junit\.Ignore\s*;/g,
      "import org.junit.jupiter.api.Disabled;",
    );
    result = result.replace(
      /import\s+org\.junit\.Assert\s*;/g,
      "import org.junit.jupiter.api.Assertions;",
    );
    result = result.replace(
      /import\s+static\s+org\.junit\.Assert\.\*\s*;/g,
      "import static org.junit.jupiter.api.Assertions.*;",
    );
    result = result.replace(
      /import\s+static\s+org\.junit\.Assert\.(\w+)\s*;/g,
      "import static org.junit.jupiter.api.Assertions.$1;",
    );
    result = result.replace(
      /import\s+org\.junit\.runner\.RunWith\s*;/g,
      "import org.junit.jupiter.api.extension.ExtendWith;",
    );
    result = result.replace(
      /import\s+org\.junit\.experimental\.categories\.Category\s*;/g,
      "import org.junit.jupiter.api.Tag;",
    );
    result = result.replace(
      /import\s+org\.junit\.Rule\s*;/g,
      "import org.junit.jupiter.api.Rule;",
    );
    result = result.replace(
      /import\s+org\.junit\.\*\s*;/g,
      "import org.junit.jupiter.api.*;",
    );
    result = result.replace(
      /import\s+org\.junit\.Assume\s*;/g,
      "import org.junit.jupiter.api.Assumptions;",
    );

    // --- Phase 3: JUnit 4 assertion class renames ---

    result = result.replace(
      /\bAssert\.(assertEquals|assertTrue|assertFalse|assertNull|assertNotNull|assertSame|assertNotSame|assertArrayEquals|assertNotEquals|fail)\b/g,
      "Assertions.$1",
    );
    result = result.replace(
      /\bAssume\.(assumeTrue|assumeFalse|assumeNotNull|assumeThat)\b/g,
      "Assumptions.$1",
    );

    // --- Phase 4: Assertion message parameter reorder ---
    // JUnit 4: Assert.assertEquals("msg", expected, actual) → 3-arg
    // JUnit 5: Assertions.assertEquals(expected, actual, "msg") → message last
    // Only reorder 3-arg calls. 2-arg calls are never reordered.

    result = reorderAssertionArgs(result, "assertEquals", 3);
    result = reorderAssertionArgs(result, "assertNotEquals", 3);
    result = reorderAssertionArgs(result, "assertSame", 3);
    result = reorderAssertionArgs(result, "assertNotSame", 3);
    result = reorderAssertionArgs(result, "assertArrayEquals", 3);

    // For assertTrue/assertFalse/assertNull/assertNotNull, message form is 2-arg
    result = reorderAssertionArgs(result, "assertTrue", 2);
    result = reorderAssertionArgs(result, "assertFalse", 2);
    result = reorderAssertionArgs(result, "assertNull", 2);
    result = reorderAssertionArgs(result, "assertNotNull", 2);

    // --- Phase 5: @Test(expected = X.class) → assertThrows lambda wrapping ---

    result = convertExpectedExceptions(result);

    // --- Phase 6: @Test(timeout = N) → assertTimeout wrapping ---

    result = convertTimeoutAnnotations(result);

    // --- Phase 7: Unconvertible JUnit 4 patterns → HAMLET-TODO ---

    result = markUnconvertibleJUnit4Patterns(result);
  } // end JUnit 4 → JUnit 5

  // ========================================
  // TestNG → JUnit 5 Phases (8-14)
  // ========================================

  if (isTestNGSource) {
    // --- Phase 8: TestNG annotation renames ---

    result = result.replace(/@BeforeMethod\b/g, "@BeforeEach");
    result = result.replace(/@AfterMethod\b/g, "@AfterEach");

    // TestNG @BeforeClass/@AfterClass → JUnit 5 @BeforeAll/@AfterAll
    // Only match if it's a TestNG file (has testng imports or annotations)
    // These are safe since JUnit4 @BeforeClass/@AfterClass were already converted in Phase 1
    // At this point any remaining @BeforeClass/@AfterClass must be from TestNG source
    result = result.replace(/@BeforeClass\b/g, "@BeforeAll");
    result = result.replace(/@AfterClass\b/g, "@AfterAll");

    // --- Phase 9: @Test(enabled = false) → @Disabled + @Test ---

    result = result.replace(
      /@Test\s*\(\s*enabled\s*=\s*false\s*\)/g,
      "@Disabled\n    @Test",
    );

    // --- Phase 10: @Test(groups = {"x"}) → @Tag("x") + @Test ---

    result = result.replace(
      /@Test\s*\(\s*groups\s*=\s*\{\s*"([^"]+)"\s*\}\s*\)/g,
      '@Tag("$1")\n    @Test',
    );

    // --- Phase 11: TestNG import rewrites ---

    result = result.replace(
      /import\s+org\.testng\.annotations\.\*\s*;/g,
      "import org.junit.jupiter.api.*;",
    );
    result = result.replace(
      /import\s+org\.testng\.annotations\.Test\s*;/g,
      "import org.junit.jupiter.api.Test;",
    );
    result = result.replace(
      /import\s+org\.testng\.annotations\.BeforeMethod\s*;/g,
      "import org.junit.jupiter.api.BeforeEach;",
    );
    result = result.replace(
      /import\s+org\.testng\.annotations\.AfterMethod\s*;/g,
      "import org.junit.jupiter.api.AfterEach;",
    );
    result = result.replace(
      /import\s+org\.testng\.annotations\.BeforeClass\s*;/g,
      "import org.junit.jupiter.api.BeforeAll;",
    );
    result = result.replace(
      /import\s+org\.testng\.annotations\.AfterClass\s*;/g,
      "import org.junit.jupiter.api.AfterAll;",
    );
    result = result.replace(
      /import\s+org\.testng\.annotations\.DataProvider\s*;/g,
      "import org.junit.jupiter.params.provider.MethodSource;",
    );
    result = result.replace(
      /import\s+org\.testng\.Assert\s*;/g,
      "import org.junit.jupiter.api.Assertions;",
    );
    result = result.replace(
      /import\s+static\s+org\.testng\.Assert\.\*\s*;/g,
      "import static org.junit.jupiter.api.Assertions.*;",
    );

    // --- Phase 12: TestNG assertion arg order swap ---
    // TestNG: Assert.assertEquals(actual, expected) — actual first
    // JUnit 5: Assertions.assertEquals(expected, actual) — expected first
    // We swap the first two args. Message (if any) stays last.

    result = swapTestNGAssertionArgs(result, "assertEquals");
    result = swapTestNGAssertionArgs(result, "assertNotEquals");

    // Simple renames (no arg swap needed)
    result = result.replace(
      /\bAssert\.(assertTrue|assertFalse|assertNull|assertNotNull|fail)\b/g,
      "Assertions.$1",
    );

    // --- Phase 13: @Test(expectedExceptions = X.class) → assertThrows ---

    result = convertExpectedExceptionsTestNG(result);

    // --- Phase 14: Unconvertible TestNG patterns → HAMLET-TODO ---

    result = markUnconvertibleTestNGPatterns(result);
  } // end TestNG → JUnit 5

  // --- Cleanup ---

  // Clean up multiple blank lines
  result = result.replace(/\n{3,}/g, "\n\n");

  // Ensure trailing newline
  if (!result.endsWith("\n")) result += "\n";

  return result;
}

/**
 * Convert @Test(expected = X.class) → assertThrows() lambda wrapping.
 */
function convertExpectedExceptions(source) {
  const lines = source.split("\n");
  const result = [];
  let i = 0;
  let needsAssertThrowsImport = false;

  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trim();

    // Match @Test(expected = X.class) — possibly with other attributes
    const expectedMatch = trimmed.match(
      /@Test\s*\(\s*expected\s*=\s*([\w.]+)\.class\s*\)/,
    );

    if (expectedMatch) {
      const exceptionClass = expectedMatch[1];

      // Replace the annotation
      result.push(
        line.replace(/@Test\s*\(\s*expected\s*=\s*[\w.]+\.class\s*\)/, "@Test"),
      );
      i++;

      // Find the method signature and body
      const bodyResult = extractMethodBody(lines, i);
      if (bodyResult) {
        const { startLine, endLine } = bodyResult;

        // Copy lines up to the opening brace
        for (let j = i; j <= startLine; j++) {
          result.push(lines[j]);
        }

        // Extract body content (between the opening { and closing })
        const bodyContent = [];
        for (let j = startLine + 1; j < endLine; j++) {
          bodyContent.push(lines[j]);
        }

        // Indent and wrap in assertThrows
        const indent = lines[startLine].match(/^(\s*)/)[1];
        result.push(
          `${indent}    assertThrows(${exceptionClass}.class, () -> {`,
        );
        for (const bodyLine of bodyContent) {
          result.push(`    ${bodyLine}`);
        }
        result.push(`${indent}    });`);
        result.push(lines[endLine]); // closing brace of method

        i = endLine + 1;
        needsAssertThrowsImport = true;
      } else {
        // Brace-counting failed — emit HAMLET-TODO
        result.push(
          line.replace(
            /@Test\s*\(\s*expected\s*=\s*[\w.]+\.class\s*\)/,
            "@Test",
          ),
        );
        const todoComment = formatter.formatTodo({
          id: "UNCONVERTIBLE-EXPECTED-EXCEPTION",
          description: `@Test(expected = ${exceptionClass}.class) could not be auto-wrapped`,
          original: trimmed,
          action: "Manually wrap test body in assertThrows()",
        });
        result.push(todoComment);
      }
    } else {
      result.push(line);
      i++;
    }
  }

  let output = result.join("\n");

  // Add assertThrows import if needed
  if (
    needsAssertThrowsImport &&
    !/import\s+static\s+org\.junit\.jupiter\.api\.Assertions\.assertThrows\s*;/.test(
      output,
    )
  ) {
    output = addImport(
      output,
      "import static org.junit.jupiter.api.Assertions.assertThrows;",
    );
  }

  return output;
}

/**
 * Convert @Test(timeout = N) → assertTimeout() wrapping.
 */
function convertTimeoutAnnotations(source) {
  const lines = source.split("\n");
  const result = [];
  let i = 0;
  let needsTimeoutImport = false;
  let needsDurationImport = false;

  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trim();

    const timeoutMatch = trimmed.match(/@Test\s*\(\s*timeout\s*=\s*(\d+)\s*\)/);

    if (timeoutMatch) {
      const timeoutMs = timeoutMatch[1];

      // Replace the annotation
      result.push(line.replace(/@Test\s*\(\s*timeout\s*=\s*\d+\s*\)/, "@Test"));
      i++;

      // Find the method signature and body
      const bodyResult = extractMethodBody(lines, i);
      if (bodyResult) {
        const { startLine, endLine } = bodyResult;

        // Copy lines up to the opening brace
        for (let j = i; j <= startLine; j++) {
          result.push(lines[j]);
        }

        // Extract body content
        const bodyContent = [];
        for (let j = startLine + 1; j < endLine; j++) {
          bodyContent.push(lines[j]);
        }

        // Indent and wrap in assertTimeout
        const indent = lines[startLine].match(/^(\s*)/)[1];
        result.push(
          `${indent}    assertTimeout(Duration.ofMillis(${timeoutMs}), () -> {`,
        );
        for (const bodyLine of bodyContent) {
          result.push(`    ${bodyLine}`);
        }
        result.push(`${indent}    });`);
        result.push(lines[endLine]); // closing brace of method

        i = endLine + 1;
        needsTimeoutImport = true;
        needsDurationImport = true;
      } else {
        result.push(
          line.replace(/@Test\s*\(\s*timeout\s*=\s*\d+\s*\)/, "@Test"),
        );
        const todoComment = formatter.formatTodo({
          id: "UNCONVERTIBLE-TIMEOUT",
          description: `@Test(timeout = ${timeoutMs}) could not be auto-wrapped`,
          original: trimmed,
          action: "Manually wrap test body in assertTimeout()",
        });
        result.push(todoComment);
      }
    } else {
      result.push(line);
      i++;
    }
  }

  let output = result.join("\n");

  if (
    needsTimeoutImport &&
    !/import\s+static\s+org\.junit\.jupiter\.api\.Assertions\.assertTimeout\s*;/.test(
      output,
    )
  ) {
    output = addImport(
      output,
      "import static org.junit.jupiter.api.Assertions.assertTimeout;",
    );
  }
  if (
    needsDurationImport &&
    !/import\s+java\.time\.Duration\s*;/.test(output)
  ) {
    output = addImport(output, "import java.time.Duration;");
  }

  return output;
}

/**
 * Mark unconvertible JUnit 4 patterns with HAMLET-TODO.
 */
function markUnconvertibleJUnit4Patterns(source) {
  let result = source;

  // @Rule with specific type suggestions
  const rulePatterns = [
    {
      pattern: /ExpectedException/,
      suggestion: "Use `assertThrows()` instead",
    },
    { pattern: /TemporaryFolder/, suggestion: "Use `@TempDir` annotation" },
    { pattern: /TestName/, suggestion: "Use `TestInfo` parameter injection" },
    { pattern: /Timeout/, suggestion: "Use `@Timeout` annotation" },
    {
      pattern: /ExternalResource/,
      suggestion: "Implement `BeforeEachCallback`/`AfterEachCallback`",
    },
    {
      pattern: /ErrorCollector/,
      suggestion: "Use `assertAll()` for grouped assertions",
    },
  ];

  // @Rule annotations — look ahead at following lines for specific type
  result = result.replace(
    /^([ \t]*)@(?:Class)?Rule\b[^\n]*/gm,
    (match, indent, offset, fullStr) => {
      // Check the match line plus the next few lines for rule type
      const context = fullStr.substring(offset, offset + match.length + 200);
      let suggestion = "Manually refactor to JUnit 5 extension";
      for (const { pattern, suggestion: sug } of rulePatterns) {
        if (pattern.test(context)) {
          suggestion = sug;
          break;
        }
      }
      return (
        formatter
          .formatTodo({
            id: "UNCONVERTIBLE-RULE",
            description:
              "JUnit 4 @Rule/@ClassRule has no direct JUnit 5 equivalent",
            original: match.trim(),
            action: suggestion,
          })
          .split("\n")
          .map((l) => indent + l)
          .join("\n") +
        "\n" +
        match
      );
    },
  );

  // Hamcrest assertThat
  result = result.replace(
    /^([ \t]*).*\bassertThat\b.*(?:Matchers|CoreMatchers|is\(|hasItem|containsString).*$/gm,
    (match, indent) => {
      return (
        formatter
          .formatTodo({
            id: "UNCONVERTIBLE-HAMCREST",
            description:
              "Hamcrest assertThat with matchers is not directly convertible",
            original: match.trim(),
            action: "Rewrite using JUnit 5 Assertions methods",
          })
          .split("\n")
          .map((l) => indent + l)
          .join("\n") +
        "\n" +
        match
      );
    },
  );

  return result;
}

/**
 * Swap first two arguments in TestNG Assert calls for JUnit 5 conversion.
 * TestNG: Assert.assertEquals(actual, expected[, message])
 * JUnit 5: Assertions.assertEquals(expected, actual[, message])
 */
function swapTestNGAssertionArgs(source, methodName) {
  // First rename Assert. to Assertions.
  source = source.replace(
    new RegExp(`\\bAssert\\.${methodName}\\(`, "g"),
    `Assertions.${methodName}(`,
  );

  const pattern = new RegExp(`Assertions\\.${methodName}\\(`, "g");
  let result = "";
  let lastIndex = 0;
  let match;

  while ((match = pattern.exec(source)) !== null) {
    const start = match.index + match[0].length;
    let depth = 1;
    let pos = start;
    let inStr = false;
    let strCh = "";

    while (pos < source.length && depth > 0) {
      const ch = source[pos];
      if (inStr) {
        if (ch === "\\") {
          pos++;
        } else if (ch === strCh) {
          inStr = false;
        }
      } else {
        if (ch === '"' || ch === "'") {
          inStr = true;
          strCh = ch;
        } else if (ch === "(") {
          depth++;
        } else if (ch === ")") {
          depth--;
        }
      }
      if (depth > 0) pos++;
    }

    const argsStr = source.substring(start, pos);
    const args = splitArgs(argsStr);

    if (args.length >= 2) {
      // Swap first two args, keep rest in order
      const swapped = [args[1], args[0], ...args.slice(2)];
      result += source.substring(lastIndex, match.index);
      result += `Assertions.${methodName}(${swapped.join(", ")})`;
      lastIndex = pos + 1;
    }
  }

  result += source.substring(lastIndex);
  return result || source;
}

/**
 * Convert TestNG @Test(expectedExceptions = X.class) → assertThrows().
 */
function convertExpectedExceptionsTestNG(source) {
  const lines = source.split("\n");
  const result = [];
  let i = 0;
  let needsAssertThrowsImport = false;

  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trim();

    const expectedMatch = trimmed.match(
      /@Test\s*\(\s*expectedExceptions\s*=\s*([\w.]+)\.class\s*\)/,
    );

    if (expectedMatch) {
      const exceptionClass = expectedMatch[1];

      result.push(
        line.replace(
          /@Test\s*\(\s*expectedExceptions\s*=\s*[\w.]+\.class\s*\)/,
          "@Test",
        ),
      );
      i++;

      const bodyResult = extractMethodBody(lines, i);
      if (bodyResult) {
        const { startLine, endLine } = bodyResult;

        for (let j = i; j <= startLine; j++) {
          result.push(lines[j]);
        }

        const bodyContent = [];
        for (let j = startLine + 1; j < endLine; j++) {
          bodyContent.push(lines[j]);
        }

        const indent = lines[startLine].match(/^(\s*)/)[1];
        result.push(
          `${indent}    assertThrows(${exceptionClass}.class, () -> {`,
        );
        for (const bodyLine of bodyContent) {
          result.push(`    ${bodyLine}`);
        }
        result.push(`${indent}    });`);
        result.push(lines[endLine]);

        i = endLine + 1;
        needsAssertThrowsImport = true;
      } else {
        result.push(line);
        const todoComment = formatter.formatTodo({
          id: "UNCONVERTIBLE-EXPECTED-EXCEPTION",
          description: `@Test(expectedExceptions = ${exceptionClass}.class) could not be auto-wrapped`,
          original: trimmed,
          action: "Manually wrap test body in assertThrows()",
        });
        result.push(todoComment);
      }
    } else {
      result.push(line);
      i++;
    }
  }

  let output = result.join("\n");

  if (
    needsAssertThrowsImport &&
    !/import\s+static\s+org\.junit\.jupiter\.api\.Assertions\.assertThrows\s*;/.test(
      output,
    )
  ) {
    output = addImport(
      output,
      "import static org.junit.jupiter.api.Assertions.assertThrows;",
    );
  }

  return output;
}

/**
 * Mark unconvertible TestNG patterns with HAMLET-TODO.
 */
function markUnconvertibleTestNGPatterns(source) {
  let result = source;

  // dependsOnMethods
  result = result.replace(
    /^([ \t]*).*\bdependsOnMethods\b.*$/gm,
    (match, indent) => {
      return (
        formatter
          .formatTodo({
            id: "UNCONVERTIBLE-DEPENDS-ON-METHODS",
            description: "TestNG dependsOnMethods has no JUnit 5 equivalent",
            original: match.trim(),
            action: "Refactor tests to be independent or use @Order annotation",
          })
          .split("\n")
          .map((l) => indent + l)
          .join("\n") +
        "\n" +
        match
      );
    },
  );

  // priority
  result = result.replace(
    /^([ \t]*).*@Test\s*\(.*\bpriority\s*=.*$/gm,
    (match, indent) => {
      if (/HAMLET-TODO/.test(match)) return match;
      return (
        formatter
          .formatTodo({
            id: "UNCONVERTIBLE-PRIORITY",
            description: "TestNG priority has no direct JUnit 5 equivalent",
            original: match.trim(),
            action:
              "Use @Order annotation with @TestMethodOrder(OrderAnnotation.class)",
          })
          .split("\n")
          .map((l) => indent + l)
          .join("\n") +
        "\n" +
        match
      );
    },
  );

  // @Factory
  result = result.replace(/^([ \t]*)@Factory\b.*$/gm, (match, indent) => {
    return (
      formatter
        .formatTodo({
          id: "UNCONVERTIBLE-FACTORY",
          description: "TestNG @Factory has no direct JUnit 5 equivalent",
          original: match.trim(),
          action: "Use @ParameterizedTest or @TestFactory in JUnit 5",
        })
        .split("\n")
        .map((l) => indent + l)
        .join("\n") +
      "\n" +
      match
    );
  });

  // @Listeners
  result = result.replace(/^([ \t]*)@Listeners\b.*$/gm, (match, indent) => {
    return (
      formatter
        .formatTodo({
          id: "UNCONVERTIBLE-LISTENERS",
          description: "TestNG @Listeners has no direct JUnit 5 equivalent",
          original: match.trim(),
          action: "Use @ExtendWith with JUnit 5 extension instead",
        })
        .split("\n")
        .map((l) => indent + l)
        .join("\n") +
      "\n" +
      match
    );
  });

  return result;
}

/**
 * Add an import statement after existing imports.
 */
function addImport(source, importLine) {
  const lines = source.split("\n");
  let lastImportIdx = -1;

  for (let i = 0; i < lines.length; i++) {
    if (/^\s*import\s/.test(lines[i])) {
      lastImportIdx = i;
    }
  }

  if (lastImportIdx >= 0) {
    lines.splice(lastImportIdx + 1, 0, importLine);
  } else {
    // No imports found, add at top after package declaration
    let insertIdx = 0;
    for (let i = 0; i < lines.length; i++) {
      if (/^\s*package\s/.test(lines[i])) {
        insertIdx = i + 1;
        break;
      }
    }
    lines.splice(insertIdx, 0, "", importLine);
  }

  return lines.join("\n");
}

export default {
  name: "junit5",
  language: "java",
  paradigm: "xunit",
  detect,
  parse,
  emit,
  imports: {
    packages: [
      "org.junit.jupiter.api.Test",
      "org.junit.jupiter.api.Assertions",
      "org.junit.jupiter.api.BeforeEach",
      "org.junit.jupiter.api.AfterEach",
    ],
  },
};
