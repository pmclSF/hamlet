/**
 * Registry for framework definitions.
 *
 * Stores and retrieves framework definitions keyed by (language, name).
 * Each definition provides detect, parse, and emit methods for conversion.
 */

const REQUIRED_FIELDS = ['name', 'language', 'detect', 'parse', 'emit', 'imports', 'paradigm'];

export class FrameworkRegistry {
  #frameworks = new Map();

  /**
   * Register a framework definition.
   * @param {Object} definition
   * @param {string} definition.name - Framework name (e.g., 'jest')
   * @param {string} definition.language - Language (e.g., 'javascript')
   * @param {Function} definition.detect - (sourceCode) => confidence 0-100
   * @param {Function} definition.parse - (sourceCode) => IR TestFile node
   * @param {Function} definition.emit - (ir, sourceCode) => converted string
   * @param {Object} definition.imports - Import rewriting rules
   * @param {string} definition.paradigm - e.g., 'bdd', 'xunit', 'function'
   * @throws {Error} If required fields are missing
   */
  register(definition) {
    const missing = REQUIRED_FIELDS.filter(f => !(f in definition));
    if (missing.length > 0) {
      throw new Error(
        `Framework definition for '${definition.name || 'unknown'}' missing required fields: ${missing.join(', ')}`
      );
    }

    const key = `${definition.language}:${definition.name}`;
    this.#frameworks.set(key, definition);
  }

  /**
   * Retrieve a framework definition by name, optionally scoped to a language.
   * @param {string} name - Framework name
   * @param {string} [language] - Language to scope the lookup
   * @returns {Object|null} Framework definition or null
   */
  get(name, language) {
    if (language) {
      return this.#frameworks.get(`${language}:${name}`) || null;
    }

    for (const [key, def] of this.#frameworks) {
      if (key.endsWith(`:${name}`)) return def;
    }
    return null;
  }

  /**
   * Check if a framework is registered.
   * @param {string} name
   * @param {string} [language]
   * @returns {boolean}
   */
  has(name, language) {
    return this.get(name, language) !== null;
  }

  /**
   * List all registered frameworks, optionally filtered by language.
   * @param {string} [language]
   * @returns {Object[]} Array of framework definitions
   */
  list(language) {
    const result = [];
    for (const [key, def] of this.#frameworks) {
      if (!language || key.startsWith(`${language}:`)) {
        result.push(def);
      }
    }
    return result;
  }

  /**
   * Remove all registered frameworks.
   */
  clear() {
    this.#frameworks.clear();
  }
}
