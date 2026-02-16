/**
 * Formats HAMLET-TODO markers for unconvertible patterns.
 *
 * Produces standardized output in the target language's comment syntax
 * when code cannot be automatically converted.
 */

const COMMENT_STYLES = {
  javascript: { prefix: "//", blockStart: "/*", blockEnd: "*/" },
  python: { prefix: "#", blockStart: '"""', blockEnd: '"""' },
  ruby: { prefix: "#", blockStart: "=begin", blockEnd: "=end" },
  java: { prefix: "//", blockStart: "/*", blockEnd: "*/" },
};

export class TodoFormatter {
  /**
   * @param {string} language - Target language for comment syntax
   */
  constructor(language = "javascript") {
    this.language = language;
    this.style = COMMENT_STYLES[language] || COMMENT_STYLES.javascript;
  }

  /**
   * Format a HAMLET-TODO comment for an unconvertible pattern.
   *
   * @param {Object} options
   * @param {string} options.id - Unconvertible pattern ID (e.g., 'UNCONVERTIBLE-001')
   * @param {string} options.description - What the pattern is
   * @param {string} options.original - Original source code
   * @param {string} options.action - What the developer needs to do
   * @returns {string} Formatted TODO comment
   */
  formatTodo({ id, description, original, action }) {
    const p = this.style.prefix;
    const lines = [];

    lines.push(`${p} HAMLET-TODO [${id}]: ${description}`);

    const originalLines = original.split("\n");
    if (originalLines.length === 1) {
      lines.push(`${p} Original: ${original}`);
    } else {
      lines.push(`${p} Original:`);
      for (const line of originalLines) {
        lines.push(`${p}   ${line}`);
      }
    }

    lines.push(`${p} Manual action required: ${action}`);

    return lines.join("\n");
  }

  /**
   * Format a HAMLET-WARNING comment for patterns that convert but may need review.
   *
   * @param {Object} options
   * @param {string} options.description - What the concern is
   * @param {string} options.original - Original source code
   * @returns {string} Formatted warning comment
   */
  formatWarning({ description, original }) {
    const p = this.style.prefix;
    const lines = [];

    lines.push(`${p} HAMLET-WARNING: ${description}`);
    if (original) {
      lines.push(`${p} Original: ${original}`);
    }

    return lines.join("\n");
  }
}
