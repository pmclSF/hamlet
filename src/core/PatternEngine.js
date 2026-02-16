/**
 * Pattern matching and replacement engine for test conversion
 * Handles regex-based transformations between frameworks
 */
export class PatternEngine {
  constructor() {
    this.patterns = new Map();
    this.transformers = new Map();
    this.stats = {
      patternsApplied: 0,
      transformersApplied: 0,
      replacements: 0
    };
  }

  /**
   * Register a pattern for conversion
   * @param {string} category - Pattern category (navigation, selector, assertion, etc.)
   * @param {string|RegExp} sourcePattern - Regex pattern for source framework
   * @param {string|Function} targetReplacement - Target framework replacement
   * @param {Object} options - Pattern options
   */
  registerPattern(category, sourcePattern, targetReplacement, options = {}) {
    if (!this.patterns.has(category)) {
      this.patterns.set(category, []);
    }

    const pattern = sourcePattern instanceof RegExp
      ? sourcePattern
      : new RegExp(sourcePattern, options.flags || 'g');

    this.patterns.get(category).push({
      pattern,
      replacement: targetReplacement,
      priority: options.priority || 0,
      description: options.description || ''
    });

    // Sort patterns by priority (higher first)
    this.patterns.get(category).sort((a, b) => b.priority - a.priority);
  }

  /**
   * Register multiple patterns at once
   * @param {string} category - Pattern category
   * @param {Object} patterns - Object mapping source patterns to replacements
   * @param {Object} options - Shared options for all patterns
   */
  registerPatterns(category, patterns, options = {}) {
    for (const [source, replacement] of Object.entries(patterns)) {
      this.registerPattern(category, source, replacement, options);
    }
  }

  /**
   * Apply all patterns to content
   * @param {string} content - Source content
   * @param {string[]|null} categories - Categories to apply (null for all)
   * @returns {string} - Transformed content
   */
  applyPatterns(content, categories = null) {
    let result = content;
    const categoriesToApply = categories || Array.from(this.patterns.keys());

    for (const category of categoriesToApply) {
      const patterns = this.patterns.get(category) || [];

      for (const { pattern, replacement } of patterns) {
        const matches = result.match(pattern);
        if (matches) {
          this.stats.replacements += matches.length;
        }

        if (typeof replacement === 'function') {
          result = result.replace(pattern, replacement);
        } else {
          result = result.replace(pattern, replacement);
        }
        this.stats.patternsApplied++;
      }
    }

    return result;
  }

  /**
   * Apply patterns with detailed tracking
   * @param {string} content - Source content
   * @param {string[]|null} categories - Categories to apply
   * @returns {Object} - { result: string, changes: Array }
   */
  applyPatternsWithTracking(content, categories = null) {
    let result = content;
    const changes = [];
    const categoriesToApply = categories || Array.from(this.patterns.keys());

    for (const category of categoriesToApply) {
      const patterns = this.patterns.get(category) || [];

      for (const { pattern, replacement, description } of patterns) {
        const beforeChange = result;

        if (typeof replacement === 'function') {
          result = result.replace(pattern, replacement);
        } else {
          result = result.replace(pattern, replacement);
        }

        if (beforeChange !== result) {
          changes.push({
            category,
            pattern: pattern.toString(),
            description,
            matches: (beforeChange.match(pattern) || []).length
          });
        }
      }
    }

    return { result, changes };
  }

  /**
   * Register a custom transformer function
   * @param {string} name - Transformer name
   * @param {Function} transformer - Transformer function (content) => content
   * @param {Object} options - Transformer options
   */
  registerTransformer(name, transformer, options = {}) {
    this.transformers.set(name, {
      fn: transformer,
      priority: options.priority || 0,
      description: options.description || ''
    });
  }

  /**
   * Apply a specific transformer
   * @param {string} name - Transformer name
   * @param {string} content - Content to transform
   * @returns {string} - Transformed content
   */
  applyTransformer(name, content) {
    const transformer = this.transformers.get(name);
    if (!transformer) {
      throw new Error(`Unknown transformer: ${name}`);
    }
    this.stats.transformersApplied++;
    return transformer.fn(content);
  }

  /**
   * Apply all transformers in order
   * @param {string} content - Content to transform
   * @param {string[]|null} names - Transformer names (null for all)
   * @returns {string} - Transformed content
   */
  applyTransformers(content, names = null) {
    let result = content;

    const transformersToApply = names
      ? names.map(n => [n, this.transformers.get(n)]).filter(([, t]) => t)
      : Array.from(this.transformers.entries());

    // Sort by priority
    transformersToApply.sort((a, b) => (b[1]?.priority || 0) - (a[1]?.priority || 0));

    for (const [_name, transformer] of transformersToApply) {
      if (transformer) {
        result = transformer.fn(result);
        this.stats.transformersApplied++;
      }
    }

    return result;
  }

  /**
   * Get all registered pattern categories
   * @returns {string[]}
   */
  getCategories() {
    return Array.from(this.patterns.keys());
  }

  /**
   * Get patterns for a specific category
   * @param {string} category - Category name
   * @returns {Array} - Array of pattern objects
   */
  getPatternsForCategory(category) {
    return this.patterns.get(category) || [];
  }

  /**
   * Get all registered transformer names
   * @returns {string[]}
   */
  getTransformerNames() {
    return Array.from(this.transformers.keys());
  }

  /**
   * Clear all patterns and transformers
   */
  clear() {
    this.patterns.clear();
    this.transformers.clear();
    this.resetStats();
  }

  /**
   * Clear patterns for a specific category
   * @param {string} category - Category to clear
   */
  clearCategory(category) {
    this.patterns.delete(category);
  }

  /**
   * Get statistics
   * @returns {Object}
   */
  getStats() {
    return { ...this.stats };
  }

  /**
   * Reset statistics
   */
  resetStats() {
    this.stats = {
      patternsApplied: 0,
      transformersApplied: 0,
      replacements: 0
    };
  }

  /**
   * Create a clone of this engine
   * @returns {PatternEngine}
   */
  clone() {
    const engine = new PatternEngine();

    for (const [category, patterns] of this.patterns) {
      engine.patterns.set(category, [...patterns]);
    }

    for (const [name, transformer] of this.transformers) {
      engine.transformers.set(name, { ...transformer });
    }

    return engine;
  }

  /**
   * Merge patterns from another engine
   * @param {PatternEngine} other - Engine to merge from
   * @param {Object} options - Merge options
   */
  merge(other, options = {}) {
    const overwrite = options.overwrite ?? false;

    for (const [category, patterns] of other.patterns) {
      if (!this.patterns.has(category) || overwrite) {
        this.patterns.set(category, [...patterns]);
      } else {
        this.patterns.get(category).push(...patterns);
        this.patterns.get(category).sort((a, b) => b.priority - a.priority);
      }
    }

    for (const [name, transformer] of other.transformers) {
      if (!this.transformers.has(name) || overwrite) {
        this.transformers.set(name, { ...transformer });
      }
    }
  }
}
