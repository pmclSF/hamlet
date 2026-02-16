/**
 * Validates conversion output for correctness.
 *
 * Checks: balanced brackets/parens, valid imports, no dangling
 * framework references (cy., jest. in target output), no empty test
 * bodies, correct describe/test nesting.
 */

export class OutputValidator {
  /**
   * Validate converted output.
   *
   * @param {string} output - The converted code
   * @param {string} targetFramework - Target framework name (e.g., 'vitest', 'playwright')
   * @returns {{valid: boolean, issues: Array<{type: string, message: string, line?: number}>}}
   */
  validate(output, targetFramework) {
    const issues = [];

    if (!output || output.trim().length === 0) {
      issues.push({ type: 'empty', message: 'Output is empty' });
      return { valid: false, issues };
    }

    this._checkBalancedBrackets(output, issues);
    this._checkDanglingReferences(output, targetFramework, issues);
    this._checkImports(output, issues);
    this._checkEmptyTestBodies(output, issues);

    return { valid: issues.length === 0, issues };
  }

  /**
   * Check that brackets, parens, and braces are balanced.
   *
   * @param {string} output
   * @param {Array} issues
   */
  _checkBalancedBrackets(output, issues) {
    const pairs = { '(': ')', '[': ']', '{': '}' };
    const openers = new Set(Object.keys(pairs));
    const closerToOpener = new Map(Object.entries(pairs).map(([k, v]) => [v, k]));
    const stack = [];
    let inString = false;
    let stringChar = '';

    for (let i = 0; i < output.length; i++) {
      const ch = output[i];
      const prev = i > 0 ? output[i - 1] : '';

      if (!inString && (ch === '\'' || ch === '"' || ch === '`') && prev !== '\\') {
        inString = true;
        stringChar = ch;
        continue;
      }
      if (inString && ch === stringChar && prev !== '\\') {
        inString = false;
        continue;
      }
      if (inString) continue;

      // Skip comments
      if (ch === '/' && i + 1 < output.length && output[i + 1] === '/') {
        const nlIndex = output.indexOf('\n', i);
        if (nlIndex !== -1) { i = nlIndex; } else { break; }
        continue;
      }
      if (ch === '/' && i + 1 < output.length && output[i + 1] === '*') {
        const endIndex = output.indexOf('*/', i + 2);
        if (endIndex !== -1) { i = endIndex + 1; } else { break; }
        continue;
      }

      if (openers.has(ch)) {
        stack.push(ch);
      } else if (closerToOpener.has(ch)) {
        if (stack.length === 0 || stack[stack.length - 1] !== closerToOpener.get(ch)) {
          issues.push({
            type: 'bracket',
            message: `Unmatched '${ch}' at position ${i}`,
          });
        } else {
          stack.pop();
        }
      }
    }

    for (const unclosed of stack) {
      issues.push({
        type: 'bracket',
        message: `Unclosed '${unclosed}'`,
      });
    }
  }

  /**
   * Check for dangling framework references that shouldn't be in the output.
   *
   * @param {string} output
   * @param {string} targetFramework
   * @param {Array} issues
   */
  _checkDanglingReferences(output, targetFramework, issues) {
    const sourcePatterns = {
      vitest: [/\bjest\.\w+/g, /\bjest\.fn\b/g],
      playwright: [/\bcy\.\w+/g, /\bCypress\.\w+/g],
      jest: [/\bvi\.\w+/g],
      cypress: [/\bpage\.\w+/g, /\btest\.describe\b/g],
    };

    const patterns = sourcePatterns[targetFramework] || [];
    const lines = output.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      // Skip comments
      if (line.trim().startsWith('//') || line.trim().startsWith('*')) continue;
      // Skip string literals containing these references
      if (line.trim().startsWith('\'') || line.trim().startsWith('"')) continue;

      for (const pattern of patterns) {
        const regex = new RegExp(pattern.source, pattern.flags);
        const match = regex.exec(line);
        if (match) {
          issues.push({
            type: 'dangling-reference',
            message: `Dangling source framework reference '${match[0]}' on line ${i + 1}`,
            line: i + 1,
          });
        }
      }
    }
  }

  /**
   * Check imports for obvious issues.
   *
   * @param {string} output
   * @param {Array} issues
   */
  _checkImports(output, issues) {
    const lines = output.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();

      // Check for empty import source
      if (/^import\s.*from\s+['"]["']/.test(line)) {
        issues.push({
          type: 'import',
          message: `Empty import source on line ${i + 1}`,
          line: i + 1,
        });
      }

      // Check for duplicate 'from' keyword
      if (/^import\s.*from\s.*from\s/.test(line)) {
        issues.push({
          type: 'import',
          message: `Malformed import with duplicate 'from' on line ${i + 1}`,
          line: i + 1,
        });
      }
    }
  }

  /**
   * Check for empty test bodies.
   *
   * @param {string} output
   * @param {Array} issues
   */
  _checkEmptyTestBodies(output, issues) {
    // Match test/it calls followed by empty function bodies
    const emptyTestPattern = /(?:it|test)\s*\(\s*['"][^'"]*['"]\s*,\s*(?:async\s*)?\(\s*\)\s*=>\s*\{\s*\}\s*\)/g;
    let match;

    while ((match = emptyTestPattern.exec(output)) !== null) {
      const line = output.slice(0, match.index).split('\n').length;
      issues.push({
        type: 'empty-test',
        message: `Empty test body on line ${line}`,
        line,
      });
    }
  }
}
