/**
 * Rewrites import paths in file content based on a rename map.
 *
 * Handles ALL import patterns: ES named, default, namespace, mixed,
 * require, destructured require, dynamic import, re-export, export-all,
 * type-only, side-effect.
 *
 * Must NOT rewrite: node_modules imports, imports in comments,
 * imports in string literals, substring matches.
 */

export class ImportRewriter {
  /**
   * Rewrite import paths in content based on a rename map.
   *
   * @param {string} content - File content
   * @param {Map<string, string>} renames - Map<oldPath, newPath>
   * @returns {string} Content with rewritten imports
   */
  rewrite(content, renames) {
    if (!renames || renames.size === 0) return content;

    const lines = content.split('\n');
    const result = [];
    let inBlockComment = false;
    let inMultilineImport = false;
    let multilineBuffer = '';

    for (let i = 0; i < lines.length; i++) {
      let line = lines[i];

      // Track block comments
      if (!inMultilineImport) {
        if (inBlockComment) {
          const endIdx = line.indexOf('*/');
          if (endIdx !== -1) {
            inBlockComment = false;
          }
          result.push(line);
          continue;
        }

        const trimmed = line.trim();

        // Skip single-line comments
        if (trimmed.startsWith('//')) {
          result.push(line);
          continue;
        }

        // Check for block comment start
        if (trimmed.startsWith('/*')) {
          if (!trimmed.includes('*/')) {
            inBlockComment = true;
          }
          result.push(line);
          continue;
        }
      }

      // Handle multiline imports
      if (inMultilineImport) {
        multilineBuffer += '\n' + line;
        if (this._hasClosingQuote(multilineBuffer)) {
          inMultilineImport = false;
          const rewritten = this._rewriteLine(multilineBuffer, renames);
          result.push(rewritten);
          multilineBuffer = '';
        }
        continue;
      }

      // Check if this line starts an import/require/export-from
      if (this._isImportLine(line)) {
        if (this._isComplete(line)) {
          result.push(this._rewriteLine(line, renames));
        } else {
          // Multiline import
          inMultilineImport = true;
          multilineBuffer = line;
        }
      } else {
        result.push(line);
      }
    }

    // If we ended in a multiline import (shouldn't happen), flush it
    if (multilineBuffer) {
      result.push(this._rewriteLine(multilineBuffer, renames));
    }

    return result.join('\n');
  }

  /**
   * Check if a line is an import/require/export-from statement.
   * @param {string} line
   * @returns {boolean}
   */
  _isImportLine(line) {
    const trimmed = line.trim();
    return (
      trimmed.startsWith('import ') ||
      trimmed.startsWith('import(') ||
      /^\s*(?:const|let|var)\s+.*=\s*require\s*\(/.test(line) ||
      /^\s*(?:const|let|var)\s+.*=\s*await\s+import\s*\(/.test(line) ||
      /^\s*export\s+(?:\{[^}]*\}|\*)\s+from\s/.test(trimmed) ||
      /^\s*import\s*\(/.test(line) ||
      /await\s+import\s*\(/.test(line)
    );
  }

  /**
   * Check if an import statement is complete (has closing quote).
   * @param {string} line
   * @returns {boolean}
   */
  _isComplete(line) {
    // A complete import has a from '...' or require('...') with closing quote
    return /['"][^'"]*['"]\s*\)?\s*;?\s*(?:\/\/.*)?$/.test(line);
  }

  /**
   * Check if a multiline buffer has a closing quote for the import source.
   * @param {string} buffer
   * @returns {boolean}
   */
  _hasClosingQuote(buffer) {
    // Look for the pattern: from '...' or require('...') at the end
    return /from\s+['"][^'"]*['"]\s*;?\s*(?:\/\/.*)?$/.test(buffer) ||
           /require\s*\(\s*['"][^'"]*['"]\s*\)\s*;?\s*(?:\/\/.*)?$/.test(buffer) ||
           /import\s*\(\s*['"][^'"]*['"]\s*\)\s*;?\s*(?:\/\/.*)?$/.test(buffer);
  }

  /**
   * Rewrite a single import statement line (or multiline buffer).
   *
   * @param {string} line
   * @param {Map<string, string>} renames
   * @returns {string}
   */
  _rewriteLine(line, renames) {
    // Extract the module specifier using various patterns
    const patterns = [
      // from '...' or from "..."
      /(from\s+)(['"])([^'"]+)(['"])/,
      // require('...') or require("...")
      /(require\s*\(\s*)(['"])([^'"]+)(['"])/,
      // dynamic import('...') or import("...")
      /(import\s*\(\s*)(['"])([^'"]+)(['"])/,
    ];

    for (const pattern of patterns) {
      const match = line.match(pattern);
      if (match) {
        const [, prefix, openQuote, specifier, closeQuote] = match;

        // Skip node_modules (bare specifiers)
        if (!specifier.startsWith('.') && !specifier.startsWith('/')) {
          return line;
        }

        // Find the matching rename
        const newPath = this._findRename(specifier, renames);
        if (newPath) {
          return line.replace(
            `${prefix}${openQuote}${specifier}${closeQuote}`,
            `${prefix}${openQuote}${newPath}${closeQuote}`
          );
        }

        return line;
      }
    }

    // Side-effect import: import './foo'  (no from keyword, no require)
    const sideEffectMatch = line.match(/(import\s+)(['"])([^'"]+)(['"])/);
    if (sideEffectMatch) {
      const [, prefix, openQuote, specifier, closeQuote] = sideEffectMatch;
      if (specifier.startsWith('.') || specifier.startsWith('/')) {
        const newPath = this._findRename(specifier, renames);
        if (newPath) {
          return line.replace(
            `${prefix}${openQuote}${specifier}${closeQuote}`,
            `${prefix}${openQuote}${newPath}${closeQuote}`
          );
        }
      }
    }

    return line;
  }

  /**
   * Find a matching rename for a specifier, handling extension variations.
   * Must NOT match substrings (e.g., './bar' should not match './bar-utils').
   *
   * @param {string} specifier - The import specifier
   * @param {Map<string, string>} renames
   * @returns {string|null}
   */
  _findRename(specifier, renames) {
    // Direct match
    if (renames.has(specifier)) {
      return renames.get(specifier);
    }

    // Try with/without extensions
    const extensions = ['.js', '.ts', '.jsx', '.tsx', '.mjs', '.cjs'];

    // If specifier has extension, try without
    for (const ext of extensions) {
      if (specifier.endsWith(ext)) {
        const withoutExt = specifier.slice(0, -ext.length);
        if (renames.has(withoutExt)) {
          const newBase = renames.get(withoutExt);
          // Preserve original extension style
          return newBase.endsWith(ext) ? newBase : newBase + ext;
        }
      }
    }

    // If specifier has no extension, try with extensions
    if (!extensions.some(ext => specifier.endsWith(ext))) {
      for (const ext of extensions) {
        const withExt = specifier + ext;
        if (renames.has(withExt)) {
          const newPath = renames.get(withExt);
          // Return without extension to match original style
          for (const e of extensions) {
            if (newPath.endsWith(e)) {
              return newPath.slice(0, -e.length);
            }
          }
          return newPath;
        }
      }
    }

    return null;
  }
}
