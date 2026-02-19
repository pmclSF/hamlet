/**
 * pytest framework definition.
 *
 * Provides detect, parse, and emit for the pytest testing framework.
 * parse() builds an IR tree from pytest source code for scoring.
 * emit() converts unittest or nose2 source code to pytest by applying
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

const formatter = new TodoFormatter("python");

/**
 * Detect whether source code is pytest.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong pytest signals
  if (/import\s+pytest\b/.test(source)) score += 30;
  if (/@pytest\.mark\./.test(source)) score += 25;
  if (/@pytest\.fixture\b/.test(source)) score += 25;
  if (/pytest\.raises\s*\(/.test(source)) score += 20;
  if (/pytest\.approx\s*\(/.test(source)) score += 15;

  // pytest-specific patterns
  if (/@pytest\.mark\.parametrize\b/.test(source)) score += 15;
  if (/@pytest\.mark\.skip\b/.test(source)) score += 10;
  if (/@pytest\.mark\.skipif\b/.test(source)) score += 10;
  if (/@pytest\.mark\.xfail\b/.test(source)) score += 10;

  // Bare test functions (no self) with bare assert
  const hasBareTestFunc =
    /^def\s+test_\w+\s*\([^)]*\)\s*:/m.test(source) &&
    !/def\s+test_\w+\s*\(\s*self/.test(source);
  if (hasBareTestFunc) score += 15;

  // Bare assert (without self.)
  if (/^\s+assert\s+/m.test(source) && !/self\.assert/.test(source))
    score += 10;

  // Weak signals
  if (/def\s+test_\w+/.test(source)) score += 5;

  // Negative signals: NOT pytest
  if (/import\s+unittest\b/.test(source)) score -= 30;
  if (/self\.assertEqual\b/.test(source)) score -= 25;
  if (/class\s+\w+\s*\(\s*(?:unittest\.)?TestCase\s*\)/.test(source))
    score -= 25;
  if (/from\s+nose/.test(source)) score -= 30;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse pytest source code into an IR tree.
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

    // @pytest.fixture
    if (/@pytest\.fixture\b/.test(trimmed)) {
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

    // @pytest.mark decorators
    if (/@pytest\.mark\./.test(trimmed)) {
      let modType = "tag";
      if (/@pytest\.mark\.skip\b/.test(trimmed)) modType = "skip";
      else if (/@pytest\.mark\.skipif\b/.test(trimmed)) modType = "skip";
      else if (/@pytest\.mark\.xfail\b/.test(trimmed))
        modType = "expectedFailure";
      else if (/@pytest\.mark\.parametrize\b/.test(trimmed))
        modType = "parameterized";

      allNodes.push(
        new Modifier({
          modifierType: modType,
          sourceLocation: loc,
          originalSource: line,
          confidence: "converted",
        }),
      );
      continue;
    }

    // Test functions (no self)
    if (/def\s+test_\w+\s*\(/.test(trimmed) && !/\bself\b/.test(trimmed)) {
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

    // Bare assert
    if (/^\s*assert\s+/.test(line)) {
      let kind = "truthy";
      if (/assert\s+.+==/.test(trimmed)) kind = "equal";
      else if (/assert\s+.+!=/.test(trimmed)) kind = "notEqual";
      else if (/assert\s+not\s+/.test(trimmed)) kind = "falsy";
      else if (/assert\s+.+is\s+None/.test(trimmed)) kind = "isNull";
      else if (/assert\s+.+is\s+not\s+None/.test(trimmed)) kind = "isDefined";
      else if (/assert\s+.+\s+in\s+/.test(trimmed)) kind = "contains";
      else if (/assert\s+.+\s+not\s+in\s+/.test(trimmed)) kind = "notContains";
      else if (/assert\s+isinstance/.test(trimmed)) kind = "isInstance";

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

    // pytest.raises
    if (/pytest\.raises\s*\(/.test(trimmed)) {
      allNodes.push(
        new Assertion({
          kind: "throws",
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

// ───────────────────────────────────────────────────────────
// Emit helpers
// ───────────────────────────────────────────────────────────

/**
 * Split arguments at depth 0, respecting nested parens, brackets, and strings.
 */
function splitArgs(argsStr) {
  const args = [];
  let depth = 0;
  let current = "";
  let inString = null;

  for (let i = 0; i < argsStr.length; i++) {
    const ch = argsStr[i];
    const prev = i > 0 ? argsStr[i - 1] : "";

    if (inString) {
      current += ch;
      if (ch === inString && prev !== "\\") inString = null;
      continue;
    }

    if (ch === '"' || ch === "'") {
      inString = ch;
      current += ch;
      continue;
    }

    if (ch === "(" || ch === "[" || ch === "{") {
      depth++;
      current += ch;
      continue;
    }

    if (ch === ")" || ch === "]" || ch === "}") {
      depth--;
      current += ch;
      continue;
    }

    if (ch === "," && depth === 0) {
      args.push(current.trim());
      current = "";
      continue;
    }

    current += ch;
  }

  if (current.trim()) args.push(current.trim());
  return args;
}

/**
 * Extract content between outermost parentheses starting from an index.
 */
function extractParenContent(str, startIndex) {
  let depth = 0;
  let start = -1;

  for (let i = startIndex; i < str.length; i++) {
    if (str[i] === "(") {
      if (depth === 0) start = i + 1;
      depth++;
    } else if (str[i] === ")") {
      depth--;
      if (depth === 0) {
        return { content: str.substring(start, i), end: i };
      }
    }
  }
  return null;
}

/**
 * Detect if source is unittest code.
 */
function isUnittestSource(source) {
  if (/import\s+unittest\b/.test(source)) return true;
  if (/from\s+unittest\s+import\b/.test(source)) return true;
  if (/class\s+\w+\s*\(\s*(?:unittest\.)?TestCase\s*\)/.test(source))
    return true;
  if (/self\.assert\w+\s*\(/.test(source)) return true;
  return false;
}

/**
 * Detect if source is nose2 code.
 */
function isNose2Source(source) {
  if (/from\s+nose\.tools\s+import\b/.test(source)) return true;
  if (/from\s+nose2\.tools\s+import\b/.test(source)) return true;
  if (/import\s+nose2?\b/.test(source)) return true;
  if (/\bassert_equal\s*\(/.test(source) || /\bassert_true\s*\(/.test(source))
    return true;
  return false;
}

// ───────────────────────────────────────────────────────────
// Emit: unittest → pytest conversion (Phases 1-9)
// ───────────────────────────────────────────────────────────

/**
 * Convert self.assert* calls to bare assert statements.
 */
function convertUnittestAssertions(result) {
  // self.assertEqual(a, b) -> assert a == b
  result = result.replace(
    /^(\s*)self\.assertEqual\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} == ${args[1]}`;
      return match;
    },
  );

  // self.assertNotEqual(a, b) -> assert a != b
  result = result.replace(
    /^(\s*)self\.assertNotEqual\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} != ${args[1]}`;
      return match;
    },
  );

  // self.assertTrue(x) -> assert x
  result = result.replace(
    /^(\s*)self\.assertTrue\((.+)\)\s*$/gm,
    "$1assert $2",
  );

  // self.assertFalse(x) -> assert not x
  result = result.replace(
    /^(\s*)self\.assertFalse\((.+)\)\s*$/gm,
    "$1assert not $2",
  );

  // self.assertIsNone(x) -> assert x is None
  result = result.replace(
    /^(\s*)self\.assertIsNone\((.+)\)\s*$/gm,
    "$1assert $2 is None",
  );

  // self.assertIsNotNone(x) -> assert x is not None
  result = result.replace(
    /^(\s*)self\.assertIsNotNone\((.+)\)\s*$/gm,
    "$1assert $2 is not None",
  );

  // self.assertIn(a, b) -> assert a in b
  result = result.replace(
    /^(\s*)self\.assertIn\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} in ${args[1]}`;
      return match;
    },
  );

  // self.assertNotIn(a, b) -> assert a not in b
  result = result.replace(
    /^(\s*)self\.assertNotIn\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}assert ${args[0]} not in ${args[1]}`;
      return match;
    },
  );

  // self.assertIsInstance(x, Y) -> assert isinstance(x, Y)
  result = result.replace(
    /^(\s*)self\.assertIsInstance\((.+)\)\s*$/gm,
    "$1assert isinstance($2)",
  );

  // self.assertGreater(a, b) -> assert a > b
  result = result.replace(
    /^(\s*)self\.assertGreater\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} > ${args[1]}`;
      return match;
    },
  );

  // self.assertGreaterEqual(a, b) -> assert a >= b
  result = result.replace(
    /^(\s*)self\.assertGreaterEqual\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} >= ${args[1]}`;
      return match;
    },
  );

  // self.assertLess(a, b) -> assert a < b
  result = result.replace(
    /^(\s*)self\.assertLess\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} < ${args[1]}`;
      return match;
    },
  );

  // self.assertLessEqual(a, b) -> assert a <= b
  result = result.replace(
    /^(\s*)self\.assertLessEqual\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} <= ${args[1]}`;
      return match;
    },
  );

  // self.assertAlmostEqual(a, b) -> assert a == pytest.approx(b)
  result = result.replace(
    /^(\s*)self\.assertAlmostEqual\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}assert ${args[0]} == pytest.approx(${args[1]})`;
      return match;
    },
  );

  // self.assertNotAlmostEqual(a, b) -> assert a != pytest.approx(b)
  result = result.replace(
    /^(\s*)self\.assertNotAlmostEqual\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}assert ${args[0]} != pytest.approx(${args[1]})`;
      return match;
    },
  );

  // self.assertNotIsInstance(x, Y) -> assert not isinstance(x, Y)
  result = result.replace(
    /^(\s*)self\.assertNotIsInstance\((.+)\)\s*$/gm,
    "$1assert not isinstance($2)",
  );

  // self.assertCountEqual(a, b) -> assert sorted(a) == sorted(b)
  result = result.replace(
    /^(\s*)self\.assertCountEqual\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}assert sorted(${args[0]}) == sorted(${args[1]})`;
      return match;
    },
  );

  // self.assertRegex(text, regex) -> assert re.search(regex, text)
  result = result.replace(
    /^(\s*)self\.assertRegex\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}assert re.search(${args[1]}, ${args[0]})`;
      return match;
    },
  );

  // self.assertNotRegex(text, regex) -> assert not re.search(regex, text)
  result = result.replace(
    /^(\s*)self\.assertNotRegex\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}assert not re.search(${args[1]}, ${args[0]})`;
      return match;
    },
  );

  // self.assertDictEqual/assertListEqual/assertSetEqual/assertTupleEqual(a, b) -> assert a == b
  result = result.replace(
    /^(\s*)self\.assert(?:Dict|List|Set|Tuple)Equal\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} == ${args[1]}`;
      return match;
    },
  );

  // self.assertMultiLineEqual(a, b) -> assert a == b
  result = result.replace(
    /^(\s*)self\.assertMultiLineEqual\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} == ${args[1]}`;
      return match;
    },
  );

  // self.assertSequenceEqual(a, b) -> assert list(a) == list(b)
  result = result.replace(
    /^(\s*)self\.assertSequenceEqual\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}assert list(${args[0]}) == list(${args[1]})`;
      return match;
    },
  );

  // self.assertWarns(Warning) context manager -> pytest.warns(Warning)
  result = result.replace(
    /^(\s*)with\s+self\.assertWarns\((.+)\)\s*:/gm,
    "$1with pytest.warns($2):",
  );

  // self.assertWarnsRegex(Warning, "pattern") context manager -> pytest.warns(Warning, match="pattern")
  result = result.replace(
    /^(\s*)with\s+self\.assertWarnsRegex\((.+)\)\s*:/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}with pytest.warns(${args[0]}, match=${args[1]}):`;
      return match;
    },
  );

  // self.assertRaisesRegex(E, "pattern") context manager -> pytest.raises(E, match="pattern")
  result = result.replace(
    /^(\s*)with\s+self\.assertRaisesRegex\((.+)\)\s*:/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}with pytest.raises(${args[0]}, match=${args[1]}):`;
      return match;
    },
  );

  // self.assertRaises(E) context manager -> pytest.raises(E)
  result = result.replace(
    /^(\s*)with\s+self\.assertRaises\((.+)\)\s*:/gm,
    "$1with pytest.raises($2):",
  );

  // self.assertRaisesRegex(E, pattern, callable, *args) inline -> pytest.raises(E, match=pattern) with call
  result = result.replace(
    /^(\s*)self\.assertRaisesRegex\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 3) {
        const callArgs = args.slice(2).join(', ');
        return `${indent}with pytest.raises(${args[0]}, match=${args[1]}):\n${indent}    ${callArgs.includes(',') ? args[2] + '(' + args.slice(3).join(', ') + ')' : args[2] + '()'}`;
      }
      return match;
    },
  );

  // self.assertRaises(E, callable, *args) inline -> with pytest.raises(E): callable(*args)
  result = result.replace(
    /^(\s*)self\.assertRaises\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) {
        const callable = args[1];
        const callArgs = args.slice(2).join(', ');
        return `${indent}with pytest.raises(${args[0]}):\n${indent}    ${callable}(${callArgs})`;
      }
      return match;
    },
  );

  // self.fail("msg") -> pytest.fail("msg")
  result = result.replace(
    /^(\s*)self\.fail\((.+)\)\s*$/gm,
    "$1pytest.fail($2)",
  );

  return result;
}

/**
 * Strip class wrapper, dedent methods, remove self parameter.
 */
function unwrapClass(result) {
  const lines = result.split("\n");
  const output = [];
  let inClass = false;
  let classIndent = 0;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();

    // Detect class declaration
    const classMatch = line.match(
      /^(\s*)class\s+\w+\s*\(\s*(?:unittest\.)?TestCase\s*\)\s*:/,
    );
    if (classMatch) {
      inClass = true;
      classIndent = classMatch[1].length;
      continue; // Skip the class line
    }

    if (inClass) {
      // Check if we've left the class (line at same or lesser indent that's not blank)
      if (trimmed !== "" && !line.startsWith(" ".repeat(classIndent + 1))) {
        // Check if it's another class or top-level
        const lineIndent = line.match(/^(\s*)/)[1].length;
        if (lineIndent <= classIndent) {
          inClass = false;
          output.push(line);
          continue;
        }
      }

      // Dedent by 4 spaces (class body indent)
      let dedented = line;
      if (line.startsWith(" ".repeat(classIndent + 4))) {
        dedented = line.substring(classIndent + 4);
      } else if (trimmed === "") {
        dedented = "";
      }

      // Remove self parameter from function definitions
      dedented = dedented.replace(/^(\s*def\s+\w+\s*)\(\s*self\s*,\s*/, "$1(");
      dedented = dedented.replace(/^(\s*def\s+\w+\s*)\(\s*self\s*\)/, "$1()");

      // Remove cls parameter from class methods
      dedented = dedented.replace(/^(\s*def\s+\w+\s*)\(\s*cls\s*,\s*/, "$1(");
      dedented = dedented.replace(/^(\s*def\s+\w+\s*)\(\s*cls\s*\)/, "$1()");

      output.push(dedented);
    } else {
      output.push(line);
    }
  }

  return output.join("\n");
}

/**
 * Convert setUp/tearDown to @pytest.fixture.
 */
function convertSetUpTearDown(result) {
  const lines = result.split("\n");
  const output = [];
  let i = 0;

  // First pass: find setUp and tearDown pairs
  let setUpStart = -1;
  let setUpBody = [];
  let tearDownStart = -1;
  let tearDownBody = [];

  // Collect setUp and tearDown bodies
  const parsed = {
    setUp: null,
    tearDown: null,
    setUpClass: null,
    tearDownClass: null,
  };

  for (let j = 0; j < lines.length; j++) {
    const trimmed = lines[j].trim();
    if (/^def\s+setUp\s*\(/.test(trimmed)) {
      parsed.setUp = { start: j, body: [] };
      const funcIndent = lines[j].match(/^(\s*)/)[1].length + 4;
      let k = j + 1;
      while (
        k < lines.length &&
        (lines[k].trim() === "" ||
          lines[k].match(/^(\s*)/)[1].length >= funcIndent)
      ) {
        parsed.setUp.body.push(lines[k]);
        k++;
      }
      parsed.setUp.end = k;
    }
    if (/^def\s+tearDown\s*\(\s*\)/.test(trimmed)) {
      parsed.tearDown = { start: j, body: [] };
      const funcIndent = lines[j].match(/^(\s*)/)[1].length + 4;
      let k = j + 1;
      while (
        k < lines.length &&
        (lines[k].trim() === "" ||
          lines[k].match(/^(\s*)/)[1].length >= funcIndent)
      ) {
        parsed.tearDown.body.push(lines[k]);
        k++;
      }
      parsed.tearDown.end = k;
    }
  }

  // If both setUp and tearDown exist, merge into single fixture with yield
  if (parsed.setUp && parsed.tearDown) {
    const skipLines = new Set();
    for (let j = parsed.setUp.start; j < parsed.setUp.end; j++)
      skipLines.add(j);
    for (let j = parsed.tearDown.start; j < parsed.tearDown.end; j++)
      skipLines.add(j);

    let insertedFixture = false;
    for (let j = 0; j < lines.length; j++) {
      if (skipLines.has(j)) {
        if (!insertedFixture && j === parsed.setUp.start) {
          insertedFixture = true;
          output.push("@pytest.fixture(autouse=True)");
          output.push("def setup_teardown():");
          for (const bl of parsed.setUp.body) {
            if (bl.trim() !== "") output.push(bl);
          }
          output.push("    yield");
          for (const bl of parsed.tearDown.body) {
            if (bl.trim() !== "") output.push(bl);
          }
          output.push("");
        }
        continue;
      }
      output.push(lines[j]);
    }
    return output.join("\n");
  }

  // If only setUp, convert to fixture
  if (parsed.setUp) {
    for (let j = 0; j < lines.length; j++) {
      if (j === parsed.setUp.start) {
        output.push("@pytest.fixture(autouse=True)");
        output.push("def setup():");
        for (const bl of parsed.setUp.body) {
          output.push(bl);
        }
        j = parsed.setUp.end - 1;
        continue;
      }
      output.push(lines[j]);
    }
    return output.join("\n");
  }

  // If only tearDown, convert to fixture with yield
  if (parsed.tearDown) {
    for (let j = 0; j < lines.length; j++) {
      if (j === parsed.tearDown.start) {
        output.push("@pytest.fixture(autouse=True)");
        output.push("def teardown():");
        output.push("    yield");
        for (const bl of parsed.tearDown.body) {
          output.push(bl);
        }
        j = parsed.tearDown.end - 1;
        continue;
      }
      output.push(lines[j]);
    }
    return output.join("\n");
  }

  return result;
}

/**
 * Convert unittest decorators to pytest markers.
 */
function convertUnittestDecorators(result) {
  // @unittest.skip("reason") -> @pytest.mark.skip(reason="reason")
  result = result.replace(
    /@unittest\.skip\(\s*("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')\s*\)/g,
    "@pytest.mark.skip(reason=$1)",
  );

  // @unittest.skip -> @pytest.mark.skip
  result = result.replace(/@unittest\.skip\b(?!\s*\()/g, "@pytest.mark.skip");

  // @unittest.skipIf(cond, "reason") -> @pytest.mark.skipif(cond, reason="reason")
  result = result.replace(
    /@unittest\.skipIf\(([^,]+),\s*("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')\s*\)/g,
    "@pytest.mark.skipif($1, reason=$2)",
  );

  // @unittest.skipUnless(cond, "reason") -> @pytest.mark.skipif(not cond, reason="reason")
  result = result.replace(
    /@unittest\.skipUnless\(([^,]+),\s*("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')\s*\)/g,
    "@pytest.mark.skipif(not $1, reason=$2)",
  );

  // @unittest.expectedFailure -> @pytest.mark.xfail
  result = result.replace(
    /@unittest\.expectedFailure\b/g,
    "@pytest.mark.xfail",
  );

  return result;
}

/**
 * Mark unconvertible unittest patterns.
 */
function markUnconvertibleUnittest(result) {
  // setUpModule / tearDownModule
  result = result.replace(
    /^(\s*)(def\s+(?:setUp|tearDown)Module\s*\(.*)$/gm,
    (match, indent, code) => {
      return `${indent}${formatter
        .formatTodo({
          id: "UNCONVERTIBLE-MODULE-SETUP",
          description:
            "Module-level setup/teardown has no direct pytest equivalent in-file",
          original: code.trim(),
          action: "Move to conftest.py as a session/module-scoped fixture",
        })
        .split("\n")
        .join(`\n${indent}`)}\n${match}`;
    },
  );

  // self.addCleanup
  result = result.replace(
    /^(\s*)(.*self\.addCleanup\(.*)$/gm,
    (match, indent, code) => {
      if (code.trim().startsWith("#")) return match;
      return `${indent}${formatter
        .formatTodo({
          id: "UNCONVERTIBLE-ADDCLEANUP",
          description: "self.addCleanup has no direct pytest equivalent",
          original: code.trim(),
          action: "Use a fixture with yield or request.addfinalizer",
        })
        .split("\n")
        .join(`\n${indent}`)}\n${match}`;
    },
  );

  // self.maxDiff / self.longMessage
  result = result.replace(
    /^(\s*)(.*self\.(?:maxDiff|longMessage)\s*=.*)$/gm,
    (match, indent, code) => {
      if (code.trim().startsWith("#")) return match;
      return `${indent}${formatter
        .formatTodo({
          id: "UNCONVERTIBLE-TESTCONFIG",
          description: "unittest test configuration has no pytest equivalent",
          original: code.trim(),
          action:
            "pytest handles diff display automatically; remove or configure via pytest options",
        })
        .split("\n")
        .join(`\n${indent}`)}\n${match}`;
    },
  );

  return result;
}

// ───────────────────────────────────────────────────────────
// Emit: nose2 → pytest conversion (Phases 10-14)
// ───────────────────────────────────────────────────────────

/**
 * Convert nose assertion functions to bare assert.
 */
function convertNoseAssertions(result) {
  // assert_equal(a, b) -> assert a == b
  result = result.replace(
    /^(\s*)assert_equal\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} == ${args[1]}`;
      return match;
    },
  );

  // assert_not_equal(a, b) -> assert a != b
  result = result.replace(
    /^(\s*)assert_not_equal\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} != ${args[1]}`;
      return match;
    },
  );

  // assert_true(x) -> assert x
  result = result.replace(/^(\s*)assert_true\((.+)\)\s*$/gm, "$1assert $2");

  // assert_false(x) -> assert not x
  result = result.replace(
    /^(\s*)assert_false\((.+)\)\s*$/gm,
    "$1assert not $2",
  );

  // assert_is_none(x) -> assert x is None
  result = result.replace(
    /^(\s*)assert_is_none\((.+)\)\s*$/gm,
    "$1assert $2 is None",
  );

  // assert_is_not_none(x) -> assert x is not None
  result = result.replace(
    /^(\s*)assert_is_not_none\((.+)\)\s*$/gm,
    "$1assert $2 is not None",
  );

  // assert_in(a, b) -> assert a in b
  result = result.replace(
    /^(\s*)assert_in\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2) return `${indent}assert ${args[0]} in ${args[1]}`;
      return match;
    },
  );

  // assert_not_in(a, b) -> assert a not in b
  result = result.replace(
    /^(\s*)assert_not_in\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}assert ${args[0]} not in ${args[1]}`;
      return match;
    },
  );

  // assert_raises(E) -> pytest.raises(E)
  result = result.replace(
    /^(\s*)assert_raises\((.+)\)\s*$/gm,
    "$1with pytest.raises($2):",
  );

  // assert_is_instance(x, Y) -> assert isinstance(x, Y)
  result = result.replace(
    /^(\s*)assert_is_instance\((.+)\)\s*$/gm,
    "$1assert isinstance($2)",
  );

  // assert_raises_regex(E, pattern) -> with pytest.raises(E, match=pattern):
  result = result.replace(
    /^(\s*)assert_raises_regex\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      const args = splitArgs(argsStr);
      if (args.length >= 2)
        return `${indent}with pytest.raises(${args[0]}, match=${args[1]}):`;
      return match;
    },
  );

  return result;
}

/**
 * Convert nose2 decorators to pytest equivalents.
 */
function convertNoseDecorators(result) {
  // @params(...) -> @pytest.mark.parametrize("params", [...])
  result = result.replace(
    /^(\s*)@params\((.+)\)\s*$/gm,
    (match, indent, argsStr) => {
      return `${indent}@pytest.mark.parametrize("params", [${argsStr}])`;
    },
  );

  // @attr('tag') -> @pytest.mark.tag
  result = result.replace(/@attr\(\s*(['"])(\w+)\1\s*\)/g, "@pytest.mark.$2");

  return result;
}

/**
 * Mark unconvertible nose2 patterns.
 */
function markUnconvertibleNose(result) {
  // nose2 plugin usage
  result = result.replace(
    /^(\s*)(.*nose2\.(?:tools|plugins)\..*)$/gm,
    (match, indent, code) => {
      if (code.trim().startsWith("#")) return match;
      return `${indent}${formatter
        .formatTodo({
          id: "UNCONVERTIBLE-NOSE-PLUGIN",
          description: "nose2 plugin has no direct pytest equivalent",
          original: code.trim(),
          action:
            "Find a pytest plugin or built-in feature that provides equivalent functionality",
        })
        .split("\n")
        .join(`\n${indent}`)}\n${match}`;
    },
  );

  // such DSL
  result = result.replace(
    /^(\s*)(.*\bsuch\.\w+.*)$/gm,
    (match, indent, code) => {
      if (code.trim().startsWith("#")) return match;
      return `${indent}${formatter
        .formatTodo({
          id: "UNCONVERTIBLE-SUCH-DSL",
          description: "nose2 such DSL has no direct pytest equivalent",
          original: code.trim(),
          action: "Rewrite using standard pytest test functions or classes",
        })
        .split("\n")
        .join(`\n${indent}`)}\n${match}`;
    },
  );

  return result;
}

// ───────────────────────────────────────────────────────────
// Main emit function
// ───────────────────────────────────────────────────────────

/**
 * Emit pytest code from IR + original source.
 * Converts unittest or nose2 source code to pytest.
 */
function emit(_ir, source) {
  let result = source;
  const isUnitTest = isUnittestSource(source);
  const isNose = isNose2Source(source);

  // ── unittest → pytest (Phases 1-9) ──
  if (isUnitTest) {
    // Phase 1: Remove unittest imports
    result = result.replace(/^import\s+unittest\s*\n/gm, "");
    result = result.replace(/^from\s+unittest\s+import\s+TestCase\s*\n/gm, "");
    result = result.replace(/^from\s+unittest\s+import\s+\*\s*\n/gm, "");
    result = result.replace(/^from\s+unittest\s+import\s+.+\n/gm, "");

    // Phase 2: Add import pytest (if pytest features will be needed)
    const needsPytest =
      /self\.assertRaises/.test(source) ||
      /self\.assertAlmostEqual/.test(source) ||
      /self\.assertNotAlmostEqual/.test(source) ||
      /self\.assertWarns/.test(source) ||
      /@unittest\.skip/.test(source) ||
      /self\.fail\(/.test(source) ||
      /def\s+setUp\b/.test(source) ||
      /def\s+tearDown\b/.test(source);

    if (needsPytest && !/import\s+pytest\b/.test(result)) {
      // Add at the top after any remaining imports
      const firstImportMatch = result.match(/^(?:import|from)\s+.+$/m);
      if (firstImportMatch) {
        result = result.replace(
          firstImportMatch[0],
          `import pytest\n${firstImportMatch[0]}`,
        );
      } else {
        result = `import pytest\n\n${result}`;
      }
    }

    // Phase 3: Strip class wrapper, dedent, remove self
    result = unwrapClass(result);

    // Phase 4: Convert setUp/tearDown to fixtures
    result = convertSetUpTearDown(result);

    // Phase 5: Convert assertions
    result = convertUnittestAssertions(result);

    // Phase 6: Convert decorators
    result = convertUnittestDecorators(result);

    // Phase 7: Add import re if re.search was introduced
    if (/\bre\.search\(/.test(result) && !/^import\s+re\b/m.test(result)) {
      const firstImportMatch = result.match(/^(?:import|from)\s+.+$/m);
      if (firstImportMatch) {
        result = result.replace(
          firstImportMatch[0],
          `import re\n${firstImportMatch[0]}`,
        );
      } else {
        result = `import re\n\n${result}`;
      }
    }

    // Phase 7b: Remove import unittest if somehow still present
    result = result.replace(/^import\s+unittest\s*\n/gm, "");

    // Phase 8: Cleanup
    // Remove multiple consecutive blank lines (max 2)
    result = result.replace(/\n{4,}/g, "\n\n\n");
    // Trim leading blank lines
    result = result.replace(/^\n+/, "");

    // Phase 9: Mark unconvertible
    result = markUnconvertibleUnittest(result);
  }

  // ── nose2 → pytest (Phases 10-14) ──
  if (isNose) {
    // Phase 10: Remove nose2 imports
    result = result.replace(/^from\s+nose\.tools\s+import\s+.+\n/gm, "");
    result = result.replace(/^from\s+nose2\.tools\s+import\s+.+\n/gm, "");
    result = result.replace(/^import\s+nose2?\s*\n/gm, "");
    result = result.replace(/^from\s+nose2?\s+import\s+.+\n/gm, "");

    // Phase 11: Convert decorators
    result = convertNoseDecorators(result);

    // Phase 12: Convert assertions
    result = convertNoseAssertions(result);

    // Phase 13: Add import pytest if needed
    if (/pytest\./.test(result) && !/^import\s+pytest\b/m.test(result)) {
      result = `import pytest\n\n${result}`;
    }

    // Phase 14: Mark unconvertible
    result = markUnconvertibleNose(result);
  }

  // Final cleanup
  // Remove multiple consecutive blank lines (max 2)
  result = result.replace(/\n{4,}/g, "\n\n\n");

  // Ensure file ends with newline
  if (result.length > 0 && !result.endsWith("\n")) {
    result += "\n";
  }

  return result;
}

export default {
  name: "pytest",
  language: "python",
  paradigm: "function",
  detect,
  parse,
  emit,
  imports: {
    packages: ["pytest"],
  },
};
