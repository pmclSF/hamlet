import path from 'path';

/**
 * Get the target file extension for a given framework.
 * @param {string} toFramework - Target framework name
 * @param {string} originalExt - Original file extension (e.g., '.js')
 * @returns {string} Target file extension
 */
export function getTargetExtension(toFramework, originalExt) {
  if (originalExt === '.py' || originalExt === '.java') return originalExt;
  if (toFramework === 'cypress') return '.cy' + (originalExt || '.js');
  if (toFramework === 'playwright') return '.spec' + (originalExt || '.js');
  return '.test' + (originalExt || '.js');
}

/**
 * Build an output filename by replacing test-framework suffixes.
 * @param {string} sourceBasename - Source file basename (e.g., 'auth.test.js')
 * @param {string} toFramework - Target framework name
 * @returns {string} Output filename
 */
export function buildOutputFilename(sourceBasename, toFramework) {
  const ext = path.extname(sourceBasename);
  const base = path.basename(sourceBasename, ext);
  const cleanBase = base.replace(/\.(cy|spec|test)$/, '');
  if (ext === '.py' || ext === '.java') return cleanBase + ext;
  if (toFramework === 'cypress') return cleanBase + '.cy.js';
  if (toFramework === 'playwright') return cleanBase + '.spec.js';
  return cleanBase + '.test.js';
}

/**
 * Count HAMLET-TODO markers in converted content.
 * @param {string} content - Converted file content
 * @returns {number} Number of TODO markers
 */
export function countTodos(content) {
  const matches = content.match(/HAMLET-TODO/g);
  return matches ? matches.length : 0;
}
