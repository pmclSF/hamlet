/**
 * Lightweight regex-based TypeScript type annotation stripper.
 *
 * Used by the ConversionPipeline to remove TypeScript annotations
 * before cross-language conversions (e.g., TS Cypress → Java Selenium).
 * Does not require the TypeScript compiler.
 */

export class TypeStripper {
  /**
   * Strip TypeScript type annotations from source code.
   *
   * Removes:
   * - `import type` statements entirely
   * - Interface declarations (`interface Foo { ... }`)
   * - Type alias declarations (`type Foo = ...;`)
   * - Parameter type annotations (`param: Type`)
   * - Return type annotations (`): Type =>` / `): Type {`)
   * - `as Type` assertions
   * - Generic type parameters on calls (`fn<Type>()`)
   * - Angle-bracket type assertions (`<Type>expr`)
   * - Non-null assertions (`expr!.prop` → `expr.prop`)
   * - `readonly` modifiers
   *
   * @param {string} source - TypeScript source code
   * @returns {string} - Source with type annotations removed
   */
  static strip(source) {
    let result = source;

    // Remove `import type` statements entirely
    result = result.replace(
      /^import\s+type\s+\{[^}]*\}\s+from\s+['"][^'"]+['"];?\s*$/gm,
      ''
    );
    result = result.replace(
      /^import\s+type\s+\w+\s+from\s+['"][^'"]+['"];?\s*$/gm,
      ''
    );

    // Remove type-only imports from mixed import statements
    // e.g., import { type Foo, Bar } from 'mod' → import { Bar } from 'mod'
    result = result.replace(/,\s*type\s+\w+/g, '');
    result = result.replace(/type\s+\w+\s*,\s*/g, '');

    // Remove interface declarations (single-line and multi-line)
    result = result.replace(
      /^(?:export\s+)?interface\s+\w+(?:\s+extends\s+[^{]+)?\s*\{[^}]*\}\s*$/gm,
      ''
    );
    // Multi-line interfaces: match from `interface` to closing `}`
    result = result.replace(
      /^(?:export\s+)?interface\s+\w+(?:\s+extends\s+[^{]+)?\s*\{[\s\S]*?^\}\s*$/gm,
      ''
    );

    // Remove type alias declarations
    result = result.replace(
      /^(?:export\s+)?type\s+\w+(?:<[^>]+>)?\s*=\s*[^;]+;\s*$/gm,
      ''
    );

    // Remove `as Type` assertions (but not `as const`)
    result = result.replace(/\s+as\s+(?!const\b)[A-Z]\w*(?:<[^>]+>)?/g, '');

    // Remove generic type parameters on function calls: fn<Type>(...)
    result = result.replace(
      /(\w+)\s*<(?:[A-Z]\w*(?:\s*,\s*[A-Z]\w*)*)>\s*\(/g,
      '$1('
    );

    // Remove return type annotations: ): Type => or ): Type {
    result = result.replace(
      /\)\s*:\s*(?:Promise<[^>]+>|[A-Z]\w*(?:<[^>]+>)?(?:\s*\|\s*\w+(?:<[^>]+>)?)*(?:\[\])?)\s*(?=[{=])/g,
      ') '
    );
    result = result.replace(
      /\)\s*:\s*(?:string|number|boolean|void|any|never|unknown|null|undefined)(?:\[\])?\s*(?=[{=])/g,
      ') '
    );

    // Remove parameter type annotations: (param: Type) → (param)
    // Handle multiple params: (a: Type1, b: Type2) → (a, b)
    result = result.replace(
      /(\w+)\s*:\s*(?:string|number|boolean|void|any|never|unknown|null|undefined|object)(?:\[\])?(?=\s*[,)=])/g,
      '$1'
    );
    result = result.replace(
      /(\w+)\s*:\s*[A-Z]\w*(?:<[^>]+>)?(?:\[\])?(?:\s*\|\s*\w+(?:<[^>]+>)?)*(?=\s*[,)=])/g,
      '$1'
    );

    // Remove variable type annotations: const x: Type = ... → const x = ...
    result = result.replace(
      /((?:const|let|var)\s+\w+)\s*:\s*[A-Z]\w*(?:<[^>]+>)?(?:\[\])?(?:\s*\|\s*\w+(?:<[^>]+>)?)*/g,
      '$1'
    );
    result = result.replace(
      /((?:const|let|var)\s+\w+)\s*:\s*(?:string|number|boolean|any|unknown|object)(?:\[\])?/g,
      '$1'
    );

    // Remove non-null assertions: expr!.prop → expr.prop
    result = result.replace(/(\w)!\./g, '$1.');

    // Remove `readonly` modifier from properties
    result = result.replace(/\breadonly\s+/g, '');

    // Clean up empty lines left behind
    result = result.replace(/\n{3,}/g, '\n\n');

    return result;
  }

  /**
   * Check if source code contains TypeScript type annotations.
   *
   * @param {string} source - Source code to check
   * @returns {boolean} - True if TS annotations detected
   */
  static hasTypeAnnotations(source) {
    // Check for common TS-only patterns
    if (/\binterface\s+\w+\s*\{/.test(source)) return true;
    if (/^type\s+\w+\s*=/m.test(source)) return true;
    if (/\bimport\s+type\s/.test(source)) return true;
    if (/:\s*(?:string|number|boolean|void|any|never|unknown)\b/.test(source))
      return true;
    if (/\w+\s*:\s*[A-Z]\w*(?:<[^>]+>)?\s*[,)=]/.test(source)) return true;
    if (/\bas\s+[A-Z]\w*/.test(source)) return true;
    if (/\w+<[A-Z]\w*>\s*\(/.test(source)) return true;
    return false;
  }
}
