/**
 * Command conversion patterns from Cypress to Playwright
 */
export const commandPatterns = {
    /**
     * Navigation commands
     */
    navigation: {
      'visit\\(': 'goto(',
      'go\\(': 'goto(',
      'reload\\(': 'reload(',
      'back\\(': 'goBack(',
      'forward\\(': 'goForward('
    },
  
    /**
     * Element interaction commands
     */
    interaction: {
      'click\\(': 'click(',
      'dblclick\\(': 'dblclick(',
      'rightclick\\(': 'click({ button: "right" })',
      'type\\(': 'fill(',
      'clear\\(': 'clear(',
      'focus\\(': 'focus(',
      'blur\\(': 'blur(',
      'submit\\(': 'evaluate(form => form.submit())',
      'select\\(': 'selectOption(',
      'scrollIntoView\\(': 'scrollIntoViewIfNeeded(',
      'scrollTo\\(': 'scroll(',
      'trigger\\(': 'dispatchEvent(',
      'check\\(': 'check(',
      'uncheck\\(': 'uncheck('
    },
  
    /**
     * Mouse event commands
     */
    mouse: {
      'hover\\(': 'hover(',
      'mousedown\\(': 'mouseDown(',
      'mouseup\\(': 'mouseUp(',
      'mousemove\\(': 'moveBy(',
      'mouseover\\(': 'hover()',
      'mouseout\\(': 'hover({ force: false })',
      'mouseenter\\(': 'hover()',
      'mouseleave\\(': 'hover({ force: false })'
    },
  
    /**
     * Keyboard event commands
     */
    keyboard: {
      'type\\(': {
        pattern: 'type(',
        options: {
          delay: 50
        }
      },
      'press\\(': 'press(',
      'keydown\\(': 'press(',
      'keyup\\(': 'press(',
      'keypress\\(': 'press('
    },
  
    /**
     * Element selection commands
     */
    selectors: {
      'get\\(': 'locator(',
      'find\\(': 'locator(',
      'contains\\(': 'getByText(',
      'first\\(': 'first(',
      'last\\(': 'last(',
      'eq\\(': 'nth(',
      'closest\\(': 'locator(',
      'parents\\(': 'locator(..',
      'children\\(': 'locator(>',
      'siblings\\(': 'locator(~',
      'prev\\(': 'locator(:prev)',
      'next\\(': 'locator(:next)'
    },
  
    /**
     * Element state commands
     */
    state: {
      'focus\\(': 'focus(',
      'blur\\(': 'blur(',
      'submit\\(': 'evaluate(el => el.submit())',
      'trigger\\(': 'dispatchEvent('
    },
  
    /**
     * Window/Document commands
     */
    window: {
      'window\\(\\)': 'context.newPage()',
      'document\\(\\)': 'page.document()',
      'title\\(\\)': 'page.title()',
      'viewport\\(': 'setViewportSize(',
      'scrollTo\\(': 'scroll('
    },
  
    /**
     * Storage commands
     */
    storage: {
      'clearLocalStorage\\(': 'evaluate(() => localStorage.clear())',
      'clearCookies\\(': 'context.clearCookies()',
      'setCookie\\(': 'context.addCookies(',
      'getCookie\\(': 'context.cookies('
    },
  
    /**
     * Network commands
     */
    network: {
      'request\\(': 'request.fetch(',
      'intercept\\(': 'route(',
      'wait\\(@([^)]+)\\)': 'waitForResponse(response => response.url().includes($1))',
      'server\\(': '// Use page.route() instead of cy.server()'
    },
  
    /**
     * File handling commands
     */
    files: {
      'readFile\\(': 'fs.readFile(',
      'writeFile\\(': 'fs.writeFile(',
      'fixture\\(': 'fs.readFile(path.join(\'fixtures\', ',
      'upload\\(': 'setInputFiles('
    },
  
    /**
     * Iframe handling commands
     */
    iframe: {
      'iframe\\(\\)': 'frameLocator()',
      'within\\(': 'locator(',
      'wrap\\(': 'locator('
    },
  
    /**
     * Custom command mapping helpers
     */
    helpers: {
      /**
       * Transform mouse event options
       * @param {Object} options - Original options
       * @returns {Object} - Transformed options
       */
      transformMouseOptions(options) {
        return {
          button: options.button || 'left',
          clickCount: options.detail || 1,
          modifiers: options.modifiers || []
        };
      },
  
      /**
       * Transform keyboard event options
       * @param {Object} options - Original options
       * @returns {Object} - Transformed options
       */
      transformKeyboardOptions(options) {
        return {
          delay: options.delay || 50,
          noWaitAfter: options.noWaitAfter || false
        };
      },
  
      /**
       * Transform selector
       * @param {string} selector - Original selector
       * @returns {string} - Transformed selector
       */
      transformSelector(selector) {
        // Handle Cypress-specific selectors
        if (selector.startsWith('cy-')) {
          return selector.replace('cy-', 'data-');
        }
        // Handle :contains pseudoselector
        if (selector.includes(':contains')) {
          return selector.replace(/:contains\((.*?)\)/, 'text=$1');
        }
        return selector;
      }
    }
  };
  
  export default commandPatterns;