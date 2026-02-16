/**
 * Error recovery utilities for graceful degradation during conversion.
 *
 * Wraps functions to catch errors and return partial results.
 * Recovers from parse errors with line-by-line fallback.
 * Produces HAMLET-WARNING comments for recovered sections.
 */

import { TodoFormatter } from "./TodoFormatter.js";

export class ErrorRecovery {
  constructor() {
    this.formatter = new TodoFormatter("javascript");
  }

  /**
   * Wrap a function so that it catches errors and returns a partial result.
   *
   * @param {Function} fn - The function to wrap
   * @param {*} fallbackValue - Value to return on error
   * @returns {Function} Wrapped function that returns { result, error }
   */
  wrap(fn, fallbackValue = null) {
    return (...args) => {
      try {
        const result = fn(...args);
        // Handle async functions
        if (result && typeof result.then === "function") {
          return result
            .then((r) => ({ result: r, error: null }))
            .catch((err) => ({ result: fallbackValue, error: err }));
        }
        return { result, error: null };
      } catch (err) {
        return { result: fallbackValue, error: err };
      }
    };
  }

  /**
   * Attempt to recover from a parse error by processing content line-by-line.
   *
   * Lines that fail individually are replaced with HAMLET-WARNING comments.
   *
   * @param {string} content - The original content that failed to parse
   * @param {Error} error - The original parse error
   * @param {Function} lineProcessor - Function to process individual lines: (line) => string
   * @returns {{recovered: string, warnings: string[]}}
   */
  recoverFromParseError(content, error, lineProcessor) {
    const lines = content.split("\n");
    const recoveredLines = [];
    const warnings = [];

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];

      // Skip empty lines / whitespace-only lines
      if (line.trim().length === 0) {
        recoveredLines.push(line);
        continue;
      }

      try {
        const processed = lineProcessor(line);
        recoveredLines.push(processed);
      } catch (lineError) {
        const warning = this.formatter.formatWarning({
          description: `Line ${i + 1} could not be converted: ${lineError.message}`,
          original: line.trim(),
        });
        recoveredLines.push(warning);
        recoveredLines.push(line);
        warnings.push(`Line ${i + 1}: ${lineError.message}`);
      }
    }

    if (warnings.length === 0) {
      // Add a top-level warning about the original error
      const topWarning = this.formatter.formatWarning({
        description: `File recovered from parse error: ${error.message}`,
        original: null,
      });
      recoveredLines.unshift(topWarning);
      warnings.push(`Parse error: ${error.message}`);
    }

    return {
      recovered: recoveredLines.join("\n"),
      warnings,
    };
  }
}
