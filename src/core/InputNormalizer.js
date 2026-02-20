/**
 * Normalizes malformed input content for best-effort conversion.
 *
 * Fixes: mismatched quotes, unterminated strings, unclosed brackets (best-effort).
 * Detects: encoding issues, binary files, empty files.
 * Always returns something usable (or marks file as unconvertible).
 */

export class InputNormalizer {
  /**
   * Normalize content for conversion.
   *
   * @param {string} content - Raw file content
   * @returns {{normalized: string, issues: Array<{type: string, message: string, line?: number}>}}
   */
  normalize(content) {
    const issues = [];

    // Empty content
    if (!content || content.trim().length === 0) {
      issues.push({ type: 'empty', message: 'File is empty' });
      return { normalized: content || '', issues };
    }

    // Binary detection
    if (this._isBinary(content)) {
      issues.push({ type: 'binary', message: 'File appears to be binary' });
      return { normalized: '', issues };
    }

    let normalized = content;

    // Fix encoding issues — remove BOM
    if (normalized.charCodeAt(0) === 0xfeff) {
      normalized = normalized.slice(1);
      issues.push({ type: 'encoding', message: 'Removed BOM character' });
    }

    // Normalize line endings to LF
    if (normalized.includes('\r\n') || normalized.includes('\r')) {
      normalized = normalized.replace(/\r\n/g, '\n').replace(/\r/g, '\n');
      issues.push({
        type: 'encoding',
        message: 'Normalized line endings to LF',
      });
    }

    // Fix mismatched quotes
    normalized = this._fixMismatchedQuotes(normalized, issues);

    // Detect and report unclosed brackets
    this._checkBrackets(normalized, issues);

    return { normalized, issues };
  }

  /**
   * @param {string} content
   * @returns {boolean}
   */
  _isBinary(content) {
    const sample = content.slice(0, 1024);
    for (let i = 0; i < sample.length; i++) {
      if (sample.charCodeAt(i) === 0) return true;
    }
    let nonPrintable = 0;
    for (let i = 0; i < sample.length; i++) {
      const code = sample.charCodeAt(i);
      if (code < 32 && code !== 9 && code !== 10 && code !== 13) {
        nonPrintable++;
      }
    }
    return sample.length > 0 && nonPrintable / sample.length > 0.1;
  }

  /**
   * Fix mismatched quotes on a best-effort basis.
   * Counts unescaped quotes that are not inside the other quote type.
   *
   * @param {string} content
   * @param {Array} issues
   * @returns {string}
   */
  _fixMismatchedQuotes(content, issues) {
    const lines = content.split('\n');
    const fixed = [];

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      const counts = this._countFreeQuotes(line);

      let fixedLine = line;

      if (
        counts.single % 2 !== 0 &&
        counts.double % 2 === 0 &&
        counts.backtick % 2 === 0
      ) {
        fixedLine = fixedLine + "'";
        issues.push({
          type: 'quote',
          message: `Mismatched single quote on line ${i + 1}`,
          line: i + 1,
        });
      } else if (
        counts.double % 2 !== 0 &&
        counts.single % 2 === 0 &&
        counts.backtick % 2 === 0
      ) {
        fixedLine = fixedLine + '"';
        issues.push({
          type: 'quote',
          message: `Mismatched double quote on line ${i + 1}`,
          line: i + 1,
        });
      }

      fixed.push(fixedLine);
    }

    return fixed.join('\n');
  }

  /**
   * Count quotes that are not inside another string type.
   * @param {string} line
   * @returns {{single: number, double: number, backtick: number}}
   */
  _countFreeQuotes(line) {
    let single = 0;
    let double = 0;
    let backtick = 0;
    let inString = false;
    let stringChar = '';

    for (let i = 0; i < line.length; i++) {
      const ch = line[i];
      const escaped = i > 0 && line[i - 1] === '\\';

      if (escaped) continue;

      if (!inString) {
        if (ch === "'") {
          single++;
          inString = true;
          stringChar = "'";
        } else if (ch === '"') {
          double++;
          inString = true;
          stringChar = '"';
        } else if (ch === '`') {
          backtick++;
          inString = true;
          stringChar = '`';
        }
      } else if (ch === stringChar) {
        if (ch === "'") single++;
        else if (ch === '"') double++;
        else if (ch === '`') backtick++;
        inString = false;
      }
    }

    return { single, double, backtick };
  }

  /**
   * Check for unclosed brackets/parens/braces.
   *
   * @param {string} content
   * @param {Array} issues
   */
  _checkBrackets(content, issues) {
    const pairs = { '(': ')', '[': ']', '{': '}' };
    const openers = new Set(Object.keys(pairs));
    const closers = new Map(Object.entries(pairs).map(([k, v]) => [v, k]));
    const stack = [];
    let inString = false;
    let stringChar = '';

    for (let i = 0; i < content.length; i++) {
      const ch = content[i];
      const prev = i > 0 ? content[i - 1] : '';

      // Handle string boundaries
      if (
        !inString &&
        (ch === "'" || ch === '"' || ch === '`') &&
        prev !== '\\'
      ) {
        inString = true;
        stringChar = ch;
        continue;
      }
      if (inString && ch === stringChar && prev !== '\\') {
        inString = false;
        continue;
      }
      if (inString) continue;

      // Handle single-line comments
      if (ch === '/' && i + 1 < content.length && content[i + 1] === '/') {
        const nlIndex = content.indexOf('\n', i);
        if (nlIndex !== -1) {
          i = nlIndex;
        } else {
          break;
        }
        continue;
      }

      // Handle multi-line comments
      if (ch === '/' && i + 1 < content.length && content[i + 1] === '*') {
        const endIndex = content.indexOf('*/', i + 2);
        if (endIndex !== -1) {
          i = endIndex + 1;
        } else {
          break;
        }
        continue;
      }

      if (openers.has(ch)) {
        stack.push({ char: ch, line: content.slice(0, i).split('\n').length });
      } else if (closers.has(ch)) {
        if (
          stack.length > 0 &&
          stack[stack.length - 1].char === closers.get(ch)
        ) {
          stack.pop();
        }
      }
    }

    for (const unclosed of stack) {
      issues.push({
        type: 'bracket',
        message: `Unclosed '${unclosed.char}' at line ${unclosed.line}`,
        line: unclosed.line,
      });
    }
  }
}
