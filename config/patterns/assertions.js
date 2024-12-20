/**
 * Assertion conversion patterns from Cypress to Playwright
 */
export const assertionPatterns = {
    /**
     * Visibility assertions
     */
    visibility: {
      'should\\(\'be.visible\'\\)': 'toBeVisible()',
      'should\\(\'not.be.visible\'\\)': 'toBeHidden()',
      'should\\(\'exist\'\\)': 'toBeVisible()',
      'should\\(\'not.exist\'\\)': 'toBeHidden()',
      'should\\(\'be.hidden\'\\)': 'toBeHidden()',
      'should\\(\'be.in.viewport\'\\)': 'toBeInViewport()',
      'should\\(\'be.partially.visible\'\\)': 'toBeVisible({ visible: "partly" })'
    },
  
    /**
     * Text content assertions
     */
    text: {
      'should\\(\'have.text\',\\s*([^)]+)\\)': 'toHaveText($1)',
      'should\\(\'contain\',\\s*([^)]+)\\)': 'toContainText($1)',
      'should\\(\'not.contain\',\\s*([^)]+)\\)': 'not.toContainText($1)',
      'should\\(\'have.html\',\\s*([^)]+)\\)': 'toHaveHTML($1)',
      'should\\(\'include.text\',\\s*([^)]+)\\)': 'toContainText($1)',
      'should\\(\'match\',\\s*([^)]+)\\)': 'toMatch($1)',
      'should\\(\'have.value\',\\s*([^)]+)\\)': 'toHaveValue($1)'
    },
  
    /**
     * Element state assertions
     */
    state: {
      'should\\(\'be.checked\'\\)': 'toBeChecked()',
      'should\\(\'not.be.checked\'\\)': 'not.toBeChecked()',
      'should\\(\'be.disabled\'\\)': 'toBeDisabled()',
      'should\\(\'be.enabled\'\\)': 'toBeEnabled()',
      'should\\(\'be.focused\'\\)': 'toBeFocused()',
      'should\\(\'be.empty\'\\)': 'toBeEmpty()',
      'should\\(\'be.selected\'\\)': 'toBeChecked()',
      'should\\(\'be.readonly\'\\)': 'toHaveAttribute("readonly", "")'
    },
  
    /**
     * Attribute assertions
     */
    attributes: {
      'should\\(\'have.attr\',\\s*([^,]+),\\s*([^)]+)\\)': 'toHaveAttribute($1, $2)',
      'should\\(\'have.class\',\\s*([^)]+)\\)': 'toHaveClass($1)',
      'should\\(\'have.id\',\\s*([^)]+)\\)': 'toHaveAttribute("id", $1)',
      'should\\(\'have.value\',\\s*([^)]+)\\)': 'toHaveValue($1)',
      'should\\(\'have.prop\',\\s*([^,]+),\\s*([^)]+)\\)': 'toHaveJSProperty($1, $2)',
      'should\\(\'have.css\',\\s*([^,]+),\\s*([^)]+)\\)': 'toHaveCSS($1, $2)'
    },
  
    /**
     * Length/count assertions
     */
    length: {
      'should\\(\'have.length\',\\s*([^)]+)\\)': 'toHaveCount($1)',
      'should\\(\'have.lengthOf\',\\s*([^)]+)\\)': 'toHaveCount($1)',
      'should\\(\'not.have.length\'\\)': 'not.toHaveCount(',
      'should\\(\'have.length.greaterThan\',\\s*([^)]+)\\)': 'toHaveCount({ min: $1 })',
      'should\\(\'have.length.lessThan\',\\s*([^)]+)\\)': 'toHaveCount({ max: $1 })'
    },
  
    /**
     * Numerical comparison assertions
     */
    numerical: {
      'should\\(\'be.gt\',\\s*([^)]+)\\)': 'toBeGreaterThan($1)',
      'should\\(\'be.gte\',\\s*([^)]+)\\)': 'toBeGreaterThanOrEqual($1)',
      'should\\(\'be.lt\',\\s*([^)]+)\\)': 'toBeLessThan($1)',
      'should\\(\'be.lte\',\\s*([^)]+)\\)': 'toBeLessThanOrEqual($1)',
      'should\\(\'equal\',\\s*([^)]+)\\)': 'toBe($1)',
      'should\\(\'not.equal\',\\s*([^)]+)\\)': 'not.toBe($1)'
    },
  
    /**
     * URL assertions
     */
    url: {
      'should\\(\'have.property\',\\s*\'href\',\\s*([^)]+)\\)': 'toHaveAttribute("href", $1)',
      'should\\(\'include\',\\s*([^)]+)\\)': 'toContain($1)',
      'url\\(\\)\\.should\\(\'include\',\\s*([^)]+)\\)': 'expect(page.url()).toContain($1)',
      'location\\.pathname\\.should\\(\'eq\',\\s*([^)]+)\\)': 'expect(page.url()).toMatch(new RegExp($1))'
    },
  
    /**
     * Network assertions
     */
    network: {
      'its\\(\'status\'\\)\\.should\\(\'eq\',\\s*([^)]+)\\)': 'expect(response.status()).toBe($1)',
      'its\\(\'body\'\\)\\.should\\(\'contain\',\\s*([^)]+)\\)': 'expect(await response.text()).toContain($1)',
      'should\\(\'have\\.property\',\\s*\'status\',\\s*([^)]+)\\)': 'expect(response.status()).toBe($1)'
    },
  
    /**
     * Element relationship assertions
     */
    relationships: {
      'should\\(\'have.descendants\',\\s*([^)]+)\\)': 'toHaveCount(locator($1), { minimum: 1 })',
      'should\\(\'be.descendant\',\\s*([^)]+)\\)': 'toHaveCount(locator($1).locator("ancestor::*"), { minimum: 1 })',
      'should\\(\'have.parent\',\\s*([^)]+)\\)': 'toHaveCount(locator($1).locator("parent::*"), { minimum: 1 })',
      'should\\(\'be.child\',\\s*([^)]+)\\)': 'toHaveCount(locator($1).locator("child::*"), { minimum: 1 })'
    },
  
    /**
     * Custom assertion helpers
     */
    helpers: {
      /**
       * Transform chai assertion to Playwright expect
       * @param {string} assertion - Original assertion
       * @returns {string} - Transformed assertion
       */
      transformAssertion(assertion) {
        // Handle chained assertions
        if (assertion.includes('.and.')) {
          return assertion.split('.and.').map(this.transformSingleAssertion).join('\n');
        }
        return this.transformSingleAssertion(assertion);
      },
  
      /**
       * Transform a single assertion
       * @param {string} assertion - Single assertion
       * @returns {string} - Transformed assertion
       */
      transformSingleAssertion(assertion) {
        // Remove cy. prefix
        assertion = assertion.replace(/^cy\./, '');
        
        // Handle negation
        const isNegated = assertion.includes('.not.');
        assertion = assertion.replace('.not.', '.');
        
        // Transform the assertion
        for (const [pattern, replacement] of Object.entries(assertionPatterns)) {
          const regex = new RegExp(pattern);
          if (regex.test(assertion)) {
            let transformed = assertion.replace(regex, replacement);
            if (isNegated) {
              transformed = `not.${transformed}`;
            }
            return transformed;
          }
        }
        
        return assertion;
      },
  
      /**
       * Transform assertion options
       * @param {Object} options - Original options
       * @returns {Object} - Transformed options
       */
      transformOptions(options) {
        const transformed = {};
        
        if (options.timeout) {
          transformed.timeout = options.timeout;
        }
        
        if (options.force) {
          transformed.force = options.force;
        }
        
        return transformed;
      }
    }
  };
  
  export default assertionPatterns;