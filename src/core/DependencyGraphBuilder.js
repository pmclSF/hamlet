/**
 * Builds a dependency graph from a set of files by extracting import/require statements.
 *
 * Extracts ALL import patterns: ES named, default, namespace, require,
 * dynamic import, re-exports, side-effect imports.
 *
 * Resolves relative paths with/without extensions, ../traversal, index file resolution.
 * Ignores node_modules imports (not part of project graph).
 * Detects circular dependencies (warn, don't fail).
 */

import path from "path";

/**
 * All import extraction regexes.
 * Each returns the module specifier in capture group 1.
 */
const IMPORT_PATTERNS = [
  // ES named: import { foo } from './bar'
  // ES default: import foo from './bar'
  // ES namespace: import * as foo from './bar'
  // ES mixed: import foo, { bar } from './bar'
  // ES type-only: import type { Foo } from './bar'
  /import\s+(?:type\s+)?(?:\{[^}]*\}|\*\s+as\s+\w+|\w+(?:\s*,\s*\{[^}]*\})?)\s+from\s+['"]([^'"]+)['"]/g,

  // Side-effect: import './setup'
  /import\s+['"]([^'"]+)['"]/g,

  // Require: const foo = require('./bar')
  // Require destructured: const { foo } = require('./bar')
  /(?:const|let|var)\s+(?:\{[^}]*\}|\w+)\s*=\s*require\s*\(\s*['"]([^'"]+)['"]\s*\)/g,

  // Dynamic import: const foo = await import('./bar')  or  import('./bar')
  /import\s*\(\s*['"]([^'"]+)['"]\s*\)/g,

  // Re-export: export { foo } from './bar'
  // Export all: export * from './bar'
  /export\s+(?:\{[^}]*\}|\*)\s+from\s+['"]([^'"]+)['"]/g,
];

export class DependencyGraphBuilder {
  /**
   * Build a dependency graph from a set of files.
   *
   * @param {Array<{path: string, relativePath: string, content: string}>} files
   *   Each file must have `path` (absolute), `relativePath`, and `content`.
   * @returns {{nodes: string[], edges: Map<string, string[]>, cycles: string[][]}}
   */
  build(files) {
    const fileMap = new Map();
    for (const file of files) {
      fileMap.set(file.path, file);
    }

    const knownPaths = new Set(fileMap.keys());
    const nodes = [...knownPaths];
    const edges = new Map();
    const warnings = [];

    for (const file of files) {
      const imports = this._extractImports(file.content);
      const resolvedDeps = [];

      for (const specifier of imports) {
        // Skip node_modules / bare specifiers
        if (!specifier.startsWith(".") && !specifier.startsWith("/")) continue;

        const resolved = this._resolve(specifier, file.path, knownPaths);
        if (resolved) {
          if (!resolvedDeps.includes(resolved)) {
            resolvedDeps.push(resolved);
          }
        } else {
          warnings.push(
            `Unresolved import '${specifier}' in ${file.relativePath}`,
          );
        }
      }

      edges.set(file.path, resolvedDeps);
    }

    // Detect cycles
    const cycles = this._detectCycles(nodes, edges);

    return { nodes, edges, cycles, warnings };
  }

  /**
   * Extract all import specifiers from content.
   *
   * @param {string} content
   * @returns {string[]} Array of module specifiers (may contain duplicates)
   */
  _extractImports(content) {
    const specifiers = [];
    // Strip comments to avoid matching imports inside comments
    const stripped = this._stripComments(content);

    for (const pattern of IMPORT_PATTERNS) {
      // Reset lastIndex since we reuse the regex
      const regex = new RegExp(pattern.source, pattern.flags);
      let match;
      while ((match = regex.exec(stripped)) !== null) {
        specifiers.push(match[1]);
      }
    }

    // Deduplicate
    return [...new Set(specifiers)];
  }

  /**
   * Strip single-line and multi-line comments from content.
   * @param {string} content
   * @returns {string}
   */
  _stripComments(content) {
    // Remove multi-line comments
    let result = content.replace(/\/\*[\s\S]*?\*\//g, "");
    // Remove single-line comments (but not inside strings)
    result = result.replace(/\/\/.*$/gm, "");
    return result;
  }

  /**
   * Resolve a relative import specifier to an absolute file path.
   *
   * @param {string} specifier - The import specifier (e.g., './bar')
   * @param {string} importerPath - Absolute path of the importing file
   * @param {Set<string>} knownPaths - Set of all known absolute file paths
   * @returns {string|null} Resolved absolute path, or null if not found
   */
  _resolve(specifier, importerPath, knownPaths) {
    const dir = path.dirname(importerPath);
    const resolved = path.resolve(dir, specifier);

    // Direct match (specifier already has extension)
    if (knownPaths.has(resolved)) return resolved;

    // Try common extensions
    const extensions = [".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs"];
    for (const ext of extensions) {
      const withExt = resolved + ext;
      if (knownPaths.has(withExt)) return withExt;
    }

    // Try index file resolution
    for (const ext of extensions) {
      const indexPath = path.join(resolved, `index${ext}`);
      if (knownPaths.has(indexPath)) return indexPath;
    }

    return null;
  }

  /**
   * Detect circular dependencies using DFS.
   *
   * @param {string[]} nodes
   * @param {Map<string, string[]>} edges
   * @returns {string[][]} Array of cycles (each cycle is an array of file paths)
   */
  _detectCycles(nodes, edges) {
    const WHITE = 0;
    const GRAY = 1;
    const BLACK = 2;

    const color = new Map();
    for (const n of nodes) color.set(n, WHITE);

    const cycles = [];
    const stack = [];

    const dfs = (node) => {
      color.set(node, GRAY);
      stack.push(node);

      const deps = edges.get(node) || [];
      for (const dep of deps) {
        if (color.get(dep) === GRAY) {
          // Found a cycle: extract it from the stack
          const cycleStart = stack.indexOf(dep);
          cycles.push([...stack.slice(cycleStart), dep]);
        } else if (color.get(dep) === WHITE) {
          dfs(dep);
        }
      }

      stack.pop();
      color.set(node, BLACK);
    };

    for (const node of nodes) {
      if (color.get(node) === WHITE) {
        dfs(node);
      }
    }

    return cycles;
  }
}
