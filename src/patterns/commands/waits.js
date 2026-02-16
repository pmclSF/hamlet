/**
 * Wait patterns for all three frameworks
 * Handles explicit waits, timeouts, and synchronization
 */

export const waitPatterns = {
  cypress: {
    patterns: {
      // Fixed waits
      'cy\\.wait\\((\\d+)\\)': 'WAIT($1)',

      // Alias waits
      'cy\\.wait\\([\'"]@([^\'"]+)[\'"]\\)': 'WAIT_FOR_ALIAS($1)',

      // Network waits
      'cy\\.intercept\\(([^)]+)\\)\\.as\\([\'"]([^\'"]+)[\'"]\\)': 'INTERCEPT_AS($1, $2)',

      // Element waits (implicit in Cypress via retry-ability)
      'cy\\.get\\(([^)]+),\\s*\\{\\s*timeout:\\s*(\\d+)\\s*\\}\\)': 'WAIT_FOR_ELEMENT($1, $2)'
    },

    generators: {
      'WAIT': (ms) => `cy.wait(${ms})`,
      'WAIT_FOR_ALIAS': (alias) => `cy.wait('@${alias}')`,
      'INTERCEPT_AS': (route, alias) => `cy.intercept(${route}).as('${alias}')`,
      'WAIT_FOR_ELEMENT': (sel, timeout) => `cy.get(${sel}, { timeout: ${timeout} })`,
      'WAIT_FOR_VISIBLE': (sel, timeout) => `cy.get(${sel}, { timeout: ${timeout} }).should('be.visible')`,
      'WAIT_FOR_HIDDEN': (sel, timeout) => `cy.get(${sel}, { timeout: ${timeout} }).should('not.be.visible')`,
      'WAIT_FOR_ENABLED': (sel, timeout) => `cy.get(${sel}, { timeout: ${timeout} }).should('be.enabled')`,
      'WAIT_FOR_TEXT': (sel, text, timeout) => `cy.get(${sel}, { timeout: ${timeout} }).should('contain', ${text})`,
      'WAIT_FOR_URL': (url) => `cy.url().should('include', ${url})`,
      'WAIT_FOR_NAVIGATION': () => 'cy.location(\'pathname\')',
      'WAIT_FOR_NETWORK_IDLE': () => 'cy.wait(1000)' // Cypress doesn't have built-in network idle
    }
  },

  playwright: {
    patterns: {
      // Fixed waits
      'await page\\.waitForTimeout\\((\\d+)\\)': 'WAIT($1)',

      // Element waits
      'await page\\.waitForSelector\\(([^)]+)\\)': 'WAIT_FOR_ELEMENT($1)',
      'await page\\.waitForSelector\\(([^,]+),\\s*\\{\\s*state:\\s*[\'"]visible[\'"]\\s*\\}\\)': 'WAIT_FOR_VISIBLE($1)',
      'await page\\.waitForSelector\\(([^,]+),\\s*\\{\\s*state:\\s*[\'"]hidden[\'"]\\s*\\}\\)': 'WAIT_FOR_HIDDEN($1)',
      'await page\\.waitForSelector\\(([^,]+),\\s*\\{\\s*state:\\s*[\'"]attached[\'"]\\s*\\}\\)': 'WAIT_FOR_ATTACHED($1)',
      'await page\\.waitForSelector\\(([^,]+),\\s*\\{\\s*state:\\s*[\'"]detached[\'"]\\s*\\}\\)': 'WAIT_FOR_DETACHED($1)',

      // Page waits
      'await page\\.waitForURL\\(([^)]+)\\)': 'WAIT_FOR_URL($1)',
      'await page\\.waitForLoadState\\([\'"]networkidle[\'"]\\)': 'WAIT_FOR_NETWORK_IDLE()',
      'await page\\.waitForLoadState\\([\'"]domcontentloaded[\'"]\\)': 'WAIT_FOR_DOM_LOADED()',
      'await page\\.waitForLoadState\\([\'"]load[\'"]\\)': 'WAIT_FOR_LOAD()',

      // Network waits
      'await page\\.waitForResponse\\(([^)]+)\\)': 'WAIT_FOR_RESPONSE($1)',
      'await page\\.waitForRequest\\(([^)]+)\\)': 'WAIT_FOR_REQUEST($1)',

      // Function waits
      'await page\\.waitForFunction\\(([^)]+)\\)': 'WAIT_FOR_FUNCTION($1)'
    },

    generators: {
      'WAIT': (ms) => `await page.waitForTimeout(${ms})`,
      'WAIT_FOR_ALIAS': (_alias) => `await page.waitForResponse(response => response.url().includes('${_alias}'))`,
      'INTERCEPT_AS': (route, _alias) => `await page.route(${route}, route => { route.continue(); })`,
      'WAIT_FOR_ELEMENT': (sel, timeout) => timeout ? `await page.waitForSelector(${sel}, { timeout: ${timeout} })` : `await page.waitForSelector(${sel})`,
      'WAIT_FOR_VISIBLE': (sel, timeout) => timeout ? `await page.waitForSelector(${sel}, { state: 'visible', timeout: ${timeout} })` : `await page.waitForSelector(${sel}, { state: 'visible' })`,
      'WAIT_FOR_HIDDEN': (sel, timeout) => timeout ? `await page.waitForSelector(${sel}, { state: 'hidden', timeout: ${timeout} })` : `await page.waitForSelector(${sel}, { state: 'hidden' })`,
      'WAIT_FOR_ATTACHED': (sel) => `await page.waitForSelector(${sel}, { state: 'attached' })`,
      'WAIT_FOR_DETACHED': (sel) => `await page.waitForSelector(${sel}, { state: 'detached' })`,
      'WAIT_FOR_ENABLED': (sel, _timeout) => `await page.locator(${sel}).waitFor({ state: 'visible' }); await expect(page.locator(${sel})).toBeEnabled()`,
      'WAIT_FOR_TEXT': (sel, text, timeout) => `await expect(page.locator(${sel})).toContainText(${text}, { timeout: ${timeout || 30000} })`,
      'WAIT_FOR_URL': (url) => `await page.waitForURL(${url})`,
      'WAIT_FOR_NAVIGATION': () => 'await page.waitForNavigation()',
      'WAIT_FOR_NETWORK_IDLE': () => 'await page.waitForLoadState(\'networkidle\')',
      'WAIT_FOR_DOM_LOADED': () => 'await page.waitForLoadState(\'domcontentloaded\')',
      'WAIT_FOR_LOAD': () => 'await page.waitForLoadState(\'load\')',
      'WAIT_FOR_RESPONSE': (url) => `await page.waitForResponse(${url})`,
      'WAIT_FOR_REQUEST': (url) => `await page.waitForRequest(${url})`,
      'WAIT_FOR_FUNCTION': (fn) => `await page.waitForFunction(${fn})`
    }
  },

  selenium: {
    patterns: {
      // Fixed waits
      'await driver\\.sleep\\((\\d+)\\)': 'WAIT($1)',

      // Explicit waits
      'await driver\\.wait\\(until\\.elementLocated\\(([^)]+)\\),\\s*(\\d+)\\)': 'WAIT_FOR_ELEMENT($1, $2)',
      'await driver\\.wait\\(until\\.elementIsVisible\\(([^)]+)\\),\\s*(\\d+)\\)': 'WAIT_FOR_VISIBLE($1, $2)',
      'await driver\\.wait\\(until\\.elementIsNotVisible\\(([^)]+)\\),\\s*(\\d+)\\)': 'WAIT_FOR_HIDDEN($1, $2)',
      'await driver\\.wait\\(until\\.elementIsEnabled\\(([^)]+)\\),\\s*(\\d+)\\)': 'WAIT_FOR_ENABLED($1, $2)',
      'await driver\\.wait\\(until\\.elementTextContains\\(([^,]+),\\s*([^)]+)\\),\\s*(\\d+)\\)': 'WAIT_FOR_TEXT($1, $2, $3)',
      'await driver\\.wait\\(until\\.urlContains\\(([^)]+)\\),\\s*(\\d+)\\)': 'WAIT_FOR_URL($1)',
      'await driver\\.wait\\(until\\.titleContains\\(([^)]+)\\),\\s*(\\d+)\\)': 'WAIT_FOR_TITLE($1)',
      'await driver\\.wait\\(until\\.stalenessOf\\(([^)]+)\\),\\s*(\\d+)\\)': 'WAIT_FOR_STALE($1)',
      'await driver\\.wait\\(until\\.alertIsPresent\\(\\),\\s*(\\d+)\\)': 'WAIT_FOR_ALERT()'
    },

    generators: {
      'WAIT': (ms) => `await driver.sleep(${ms})`,
      'WAIT_FOR_ALIAS': (_alias) => '// Selenium doesn\'t have alias waits - use explicit wait',
      'INTERCEPT_AS': () => '// Selenium doesn\'t have built-in network interception',
      'WAIT_FOR_ELEMENT': (sel, timeout) => `await driver.wait(until.elementLocated(${sel}), ${timeout || 10000})`,
      'WAIT_FOR_VISIBLE': (el, timeout) => `await driver.wait(until.elementIsVisible(${el}), ${timeout || 10000})`,
      'WAIT_FOR_HIDDEN': (el, timeout) => `await driver.wait(until.elementIsNotVisible(${el}), ${timeout || 10000})`,
      'WAIT_FOR_ATTACHED': (sel) => `await driver.wait(until.elementLocated(${sel}), 10000)`,
      'WAIT_FOR_DETACHED': (el) => `await driver.wait(until.stalenessOf(${el}), 10000)`,
      'WAIT_FOR_ENABLED': (el, timeout) => `await driver.wait(until.elementIsEnabled(${el}), ${timeout || 10000})`,
      'WAIT_FOR_TEXT': (el, text, timeout) => `await driver.wait(until.elementTextContains(${el}, ${text}), ${timeout || 10000})`,
      'WAIT_FOR_URL': (url) => `await driver.wait(until.urlContains(${url}), 10000)`,
      'WAIT_FOR_NAVIGATION': () => 'await driver.wait(until.urlContains(\'/\'), 10000)',
      'WAIT_FOR_NETWORK_IDLE': () => 'await driver.sleep(1000)', // Selenium doesn't have network idle
      'WAIT_FOR_DOM_LOADED': () => 'await driver.wait(async () => await driver.executeScript(\'return document.readyState\') === \'complete\', 10000)',
      'WAIT_FOR_LOAD': () => 'await driver.wait(async () => await driver.executeScript(\'return document.readyState\') === \'complete\', 10000)',
      'WAIT_FOR_RESPONSE': () => '// Use explicit waits in Selenium',
      'WAIT_FOR_REQUEST': () => '// Use explicit waits in Selenium',
      'WAIT_FOR_FUNCTION': (fn) => `await driver.wait(async () => await driver.executeScript(${fn}), 10000)`,
      'WAIT_FOR_TITLE': (title) => `await driver.wait(until.titleContains(${title}), 10000)`,
      'WAIT_FOR_STALE': (el) => `await driver.wait(until.stalenessOf(${el}), 10000)`,
      'WAIT_FOR_ALERT': () => 'await driver.wait(until.alertIsPresent(), 10000)'
    }
  }
};

/**
 * Direct conversion mappings between frameworks
 */
export const directMappings = {
  'cypress-playwright': {
    'cy\\.wait\\((\\d+)\\)': 'await page.waitForTimeout($1)',
    'cy\\.wait\\([\'"]@([^\'"]+)[\'"]\\)': 'await page.waitForResponse(response => response.url().includes("$1"))',
    'cy\\.intercept\\(([^)]+)\\)\\.as\\([\'"]([^\'"]+)[\'"]\\)': 'await page.route($1, route => route.continue())'
  },

  'cypress-selenium': {
    'cy\\.wait\\((\\d+)\\)': 'await driver.sleep($1)',
    'cy\\.wait\\([\'"]@([^\'"]+)[\'"]\\)': '// Selenium: use explicit wait for network response'
  },

  'playwright-cypress': {
    'await page\\.waitForTimeout\\((\\d+)\\)': 'cy.wait($1)',
    'await page\\.waitForSelector\\(([^)]+)\\)': 'cy.get($1)',
    'await page\\.waitForURL\\(([^)]+)\\)': 'cy.url().should("include", $1)',
    'await page\\.waitForLoadState\\([\'"]networkidle[\'"]\\)': 'cy.wait(1000)'
  },

  'playwright-selenium': {
    'await page\\.waitForTimeout\\((\\d+)\\)': 'await driver.sleep($1)',
    'await page\\.waitForSelector\\(([^)]+)\\)': 'await driver.wait(until.elementLocated(By.css($1)), 10000)',
    'await page\\.waitForURL\\(([^)]+)\\)': 'await driver.wait(until.urlContains($1), 10000)'
  },

  'selenium-cypress': {
    'await driver\\.sleep\\((\\d+)\\)': 'cy.wait($1)',
    'await driver\\.wait\\(until\\.elementLocated\\(([^)]+)\\),\\s*(\\d+)\\)': 'cy.get($1, { timeout: $2 })',
    'await driver\\.wait\\(until\\.elementIsVisible\\(([^)]+)\\),\\s*(\\d+)\\)': 'cy.get($1, { timeout: $2 }).should("be.visible")'
  },

  'selenium-playwright': {
    'await driver\\.sleep\\((\\d+)\\)': 'await page.waitForTimeout($1)',
    'await driver\\.wait\\(until\\.elementLocated\\(([^)]+)\\),\\s*(\\d+)\\)': 'await page.waitForSelector($1, { timeout: $2 })',
    'await driver\\.wait\\(until\\.elementIsVisible\\(([^)]+)\\),\\s*(\\d+)\\)': 'await page.waitForSelector($1, { state: "visible", timeout: $2 })',
    'await driver\\.wait\\(until\\.urlContains\\(([^)]+)\\),\\s*(\\d+)\\)': 'await page.waitForURL($1)'
  }
};
