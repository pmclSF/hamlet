/**
 * TestNG framework definition.
 *
 * Provides detect, parse, and emit for the TestNG testing framework.
 * parse() builds an IR tree from TestNG source code for scoring.
 * emit() converts JUnit 5 source code to TestNG by applying regex phases.
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
 * Detect whether source code is TestNG.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong TestNG signals
  if (/import\s+org\.testng/.test(source)) score += 40;
  if (/@BeforeMethod\b/.test(source)) score += 20;
  if (/@AfterMethod\b/.test(source)) score += 20;
  if (/@DataProvider\b/.test(source)) score += 20;
  if (/@Test\s*\(\s*groups\s*=/.test(source)) score += 15;
  if (/dependsOnMethods\b/.test(source)) score += 15;
  if (/@Test\s*\(\s*expectedExceptions\s*=/.test(source)) score += 15;
  if (/@Test\s*\(\s*enabled\s*=\s*false/.test(source)) score += 15;
  if (/@Factory\b/.test(source)) score += 15;
  if (/@Listeners\b/.test(source)) score += 15;
  if (
    /\bAssert\.assertEquals\b/.test(source) &&
    /import\s+org\.testng/.test(source)
  )
    score += 10;

  // Weak signals
  if (/@Test\b/.test(source)) score += 5;
  if (/@BeforeClass\b/.test(source)) score += 3;
  if (/@AfterClass\b/.test(source)) score += 3;

  // Negative signals: NOT TestNG
  if (/import\s+org\.junit\.jupiter/.test(source)) score -= 40;
  if (/import\s+org\.junit\.Test\b/.test(source)) score -= 30;
  if (/@BeforeEach\b/.test(source)) score -= 20;
  if (/@AfterEach\b/.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse TestNG source code into an IR tree.
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
    if (/@BeforeMethod\b/.test(trimmed)) {
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
    if (/@AfterMethod\b/.test(trimmed)) {
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
    if (/@BeforeClass\b/.test(trimmed)) {
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
    if (/@AfterClass\b/.test(trimmed)) {
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

    // @DataProvider
    if (/@DataProvider\b/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: "dataProvider",
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

    // Assert calls
    if (/\bAssert\.\w+\s*\(/.test(trimmed)) {
      let kind = "equal";
      if (/assertEquals/.test(trimmed)) kind = "equal";
      else if (/assertTrue/.test(trimmed)) kind = "truthy";
      else if (/assertFalse/.test(trimmed)) kind = "falsy";
      else if (/assertNull/.test(trimmed)) kind = "isNull";
      else if (/assertNotNull/.test(trimmed)) kind = "isDefined";

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
 * Split a Java argument list, respecting nested parentheses, strings, and generics.
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
 * Swap first two arguments in JUnit 5 Assertions calls for TestNG conversion.
 * JUnit 5: Assertions.assertEquals(expected, actual[, message])
 * TestNG:  Assert.assertEquals(actual, expected[, message])
 */
function swapJUnit5AssertionArgs(source, methodName) {
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
      result += `Assert.${methodName}(${swapped.join(", ")})`;
      lastIndex = pos + 1;
    }
  }

  result += source.substring(lastIndex);
  return result || source;
}

/**
 * Extract the method body from source code starting at a given method signature.
 * Uses brace-counting that skips braces inside string literals and line comments.
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
 * Convert assertThrows(X.class, () -> {...}) → @Test(expectedExceptions = X.class) + unwrap.
 */
function convertAssertThrowsToTestNG(source) {
  // Match assertThrows(ExClass.class, () -> { ... });
  const lines = source.split("\n");
  const result = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trim();

    const throwsMatch = trimmed.match(
      /assertThrows\s*\(\s*([\w.]+)\.class\s*,\s*\(\)\s*->\s*\{/,
    );

    if (throwsMatch) {
      const exceptionClass = throwsMatch[1];
      const indent = line.match(/^(\s*)/)[1];

      // Find the matching closing brace of the lambda
      let depth = 0;
      let lambdaEnd = -1;
      let foundOpen = false;

      for (let j = i; j < lines.length; j++) {
        const l = lines[j];
        for (let k = 0; k < l.length; k++) {
          const ch = l[k];
          if (ch === "{") {
            depth++;
            foundOpen = true;
          } else if (ch === "}") {
            depth--;
            if (depth === 0 && foundOpen) {
              lambdaEnd = j;
              break;
            }
          }
        }
        if (lambdaEnd >= 0) break;
      }

      if (lambdaEnd >= 0) {
        // Find the @Test annotation above this line — look backwards for the method
        // Add expectedExceptions to the @Test annotation
        // For now, extract body lines and unwrap
        const bodyLines = [];
        for (let j = i + 1; j < lambdaEnd; j++) {
          // Remove one level of indentation
          const bodyLine = lines[j].replace(/^    /, "");
          bodyLines.push(bodyLine);
        }

        // Replace assertThrows call with unwrapped body
        for (const bodyLine of bodyLines) {
          result.push(bodyLine);
        }

        // Find and update the @Test annotation for this method
        // Look backwards from the current assertion to find @Test
        for (let j = result.length - 1; j >= 0; j--) {
          if (result[j].trim() === "@Test") {
            result[j] = result[j].replace(
              "@Test",
              `@Test(expectedExceptions = ${exceptionClass}.class)`,
            );
            break;
          } else if (/^\s*@Test\s*$/.test(result[j])) {
            result[j] = result[j].replace(
              /@Test\s*$/,
              `@Test(expectedExceptions = ${exceptionClass}.class)`,
            );
            break;
          }
        }

        i = lambdaEnd + 1;
        continue;
      }
    }

    result.push(line);
    i++;
  }

  return result.join("\n");
}

/**
 * Emit TestNG code from IR + original source.
 *
 * Converts JUnit 5 source code to TestNG by applying regex phases.
 *
 * @param {TestFile} _ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original source code (JUnit 5)
 * @returns {string} Converted TestNG source code
 */
function emit(_ir, source) {
  let result = source;

  // --- Phase 1: Annotation renames ---

  result = result.replace(/@BeforeEach\b/g, "@BeforeMethod");
  result = result.replace(/@AfterEach\b/g, "@AfterMethod");
  result = result.replace(/@BeforeAll\b/g, "@BeforeClass");
  result = result.replace(/@AfterAll\b/g, "@AfterClass");

  // --- Phase 2: @Disabled → @Test(enabled = false) ---

  // If @Disabled appears before @Test, merge into @Test(enabled = false)
  result = result.replace(
    /^([ \t]*)@Disabled\s*\n\s*@Test\b(?:\s*\(\s*\))?/gm,
    "$1@Test(enabled = false)",
  );
  // If @Disabled appears alone (without @Test following), still convert
  result = result.replace(
    /@Disabled\b(?!\s*\n\s*@Test)/g,
    "@Test(enabled = false)",
  );

  // --- Phase 3: @Tag("x") → @Test(groups = {"x"}) ---

  // If @Tag appears before @Test, merge into @Test(groups = {"x"})
  result = result.replace(
    /^([ \t]*)@Tag\s*\(\s*"([^"]+)"\s*\)\s*\n\s*@Test\b(?:\s*\(\s*\))?/gm,
    '$1@Test(groups = {"$2"})',
  );
  // Standalone @Tag (no @Test following) — just add comment
  result = result.replace(
    /@Tag\s*\(\s*"([^"]+)"\s*\)(?!\s*\n\s*@Test)/g,
    '@Test(groups = {"$1"})',
  );

  // --- Phase 4: Import rewrites ---

  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.\*\s*;/g,
    "import org.testng.annotations.*;",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.Test\s*;/g,
    "import org.testng.annotations.Test;",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.BeforeEach\s*;/g,
    "import org.testng.annotations.BeforeMethod;",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.AfterEach\s*;/g,
    "import org.testng.annotations.AfterMethod;",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.BeforeAll\s*;/g,
    "import org.testng.annotations.BeforeClass;",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.AfterAll\s*;/g,
    "import org.testng.annotations.AfterClass;",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.Disabled\s*;\n?/g,
    "// TestNG uses @Test(enabled = false) instead of @Disabled\n",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.Tag\s*;\n?/g,
    "// TestNG uses @Test(groups = {...}) instead of @Tag\n",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.DisplayName\s*;\n?/g,
    "",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.Nested\s*;\n?/g,
    "",
  );
  result = result.replace(
    /import\s+org\.junit\.jupiter\.api\.Assertions\s*;/g,
    "import org.testng.Assert;",
  );
  result = result.replace(
    /import\s+static\s+org\.junit\.jupiter\.api\.Assertions\.\*\s*;/g,
    "import static org.testng.Assert.*;",
  );
  result = result.replace(
    /import\s+static\s+org\.junit\.jupiter\.api\.Assertions\.assertThrows\s*;\n?/g,
    "",
  );
  result = result.replace(
    /import\s+static\s+org\.junit\.jupiter\.api\.Assertions\.assertTimeout\s*;\n?/g,
    "",
  );

  // --- Phase 5: Assertion argument order swap ---
  // JUnit 5: Assertions.assertEquals(expected, actual[, message]) — expected first
  // TestNG:  Assert.assertEquals(actual, expected[, message]) — actual first

  result = swapJUnit5AssertionArgs(result, "assertEquals");
  result = swapJUnit5AssertionArgs(result, "assertNotEquals");

  // Simple renames (no arg swap needed)
  result = result.replace(/\bAssertions\.assertTrue\b/g, "Assert.assertTrue");
  result = result.replace(/\bAssertions\.assertFalse\b/g, "Assert.assertFalse");
  result = result.replace(/\bAssertions\.assertNull\b/g, "Assert.assertNull");
  result = result.replace(
    /\bAssertions\.assertNotNull\b/g,
    "Assert.assertNotNull",
  );
  result = result.replace(/\bAssertions\.fail\b/g, "Assert.fail");

  // --- Phase 6: assertThrows → @Test(expectedExceptions) ---

  result = convertAssertThrowsToTestNG(result);

  // --- Phase 7: Unconvertible JUnit 5 patterns → HAMLET-TODO ---

  // @DisplayName
  result = result.replace(
    /^([ \t]*)@DisplayName\s*\(.*\)$/gm,
    (match, indent) => {
      return (
        formatter
          .formatTodo({
            id: "UNCONVERTIBLE-DISPLAY-NAME",
            description: "JUnit 5 @DisplayName has no TestNG equivalent",
            original: match.trim(),
            action:
              'Use test method naming conventions or TestNG @Test(description = "...")',
          })
          .split("\n")
          .map((l) => indent + l)
          .join("\n") +
        "\n" +
        match
      );
    },
  );

  // @Nested
  result = result.replace(/^([ \t]*)@Nested\b.*$/gm, (match, indent) => {
    return (
      formatter
        .formatTodo({
          id: "UNCONVERTIBLE-NESTED",
          description: "JUnit 5 @Nested has no TestNG equivalent",
          original: match.trim(),
          action: "Flatten nested test classes or use separate test classes",
        })
        .split("\n")
        .map((l) => indent + l)
        .join("\n") +
      "\n" +
      match
    );
  });

  // @RepeatedTest
  result = result.replace(/^([ \t]*)@RepeatedTest\b.*$/gm, (match, indent) => {
    return (
      formatter
        .formatTodo({
          id: "UNCONVERTIBLE-REPEATED-TEST",
          description: "JUnit 5 @RepeatedTest has no direct TestNG equivalent",
          original: match.trim(),
          action: "Use @Test(invocationCount = N) in TestNG",
        })
        .split("\n")
        .map((l) => indent + l)
        .join("\n") +
      "\n" +
      match
    );
  });

  // @TestFactory
  result = result.replace(/^([ \t]*)@TestFactory\b.*$/gm, (match, indent) => {
    return (
      formatter
        .formatTodo({
          id: "UNCONVERTIBLE-TEST-FACTORY",
          description: "JUnit 5 @TestFactory has no TestNG equivalent",
          original: match.trim(),
          action: "Use @DataProvider or @Factory in TestNG",
        })
        .split("\n")
        .map((l) => indent + l)
        .join("\n") +
      "\n" +
      match
    );
  });

  // --- Cleanup ---

  result = result.replace(/\n{3,}/g, "\n\n");
  if (!result.endsWith("\n")) result += "\n";

  return result;
}

export default {
  name: "testng",
  language: "java",
  paradigm: "xunit",
  detect,
  parse,
  emit,
  imports: {
    packages: [
      "org.testng.annotations.Test",
      "org.testng.Assert",
      "org.testng.annotations.BeforeMethod",
      "org.testng.annotations.AfterMethod",
    ],
  },
};
