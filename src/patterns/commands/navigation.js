/**
 * Navigation patterns for all three frameworks
 * Handles page navigation, history, and URL operations
 */

export const navigationPatterns = {
  cypress: {
    // Patterns to convert FROM Cypress
    patterns: {
      // Basic navigation
      'cy\\.visit\\(([^)]+)\\)': 'NAVIGATE($1)',
      'cy\\.go\\([\'"]back[\'"]\\)': 'GO_BACK()',
      'cy\\.go\\([\'"]forward[\'"]\\)': 'GO_FORWARD()',
      'cy\\.go\\((-?\\d+)\\)': 'GO($1)',
      'cy\\.reload\\(\\)': 'RELOAD()',
      'cy\\.reload\\(true\\)': 'RELOAD_HARD()',

      // URL operations
      'cy\\.url\\(\\)': 'GET_URL()',
      'cy\\.location\\(\\)': 'GET_LOCATION()',
      'cy\\.location\\([\'"]pathname[\'"]\\)': 'GET_PATHNAME()',
      'cy\\.location\\([\'"]search[\'"]\\)': 'GET_SEARCH()',
      'cy\\.location\\([\'"]hash[\'"]\\)': 'GET_HASH()',
      'cy\\.hash\\(\\)': 'GET_HASH()',

      // Title
      'cy\\.title\\(\\)': 'GET_TITLE()'
    },

    // Generate Cypress code from abstract operations
    generators: {
      'NAVIGATE': (url) => `cy.visit(${url})`,
      'GO_BACK': () => `cy.go('back')`,
      'GO_FORWARD': () => `cy.go('forward')`,
      'GO': (delta) => `cy.go(${delta})`,
      'RELOAD': () => `cy.reload()`,
      'RELOAD_HARD': () => `cy.reload(true)`,
      'GET_URL': () => `cy.url()`,
      'GET_LOCATION': () => `cy.location()`,
      'GET_PATHNAME': () => `cy.location('pathname')`,
      'GET_SEARCH': () => `cy.location('search')`,
      'GET_HASH': () => `cy.hash()`,
      'GET_TITLE': () => `cy.title()`
    }
  },

  playwright: {
    // Patterns to convert FROM Playwright
    patterns: {
      'await page\\.goto\\(([^)]+)\\)': 'NAVIGATE($1)',
      'await page\\.goBack\\(\\)': 'GO_BACK()',
      'await page\\.goForward\\(\\)': 'GO_FORWARD()',
      'await page\\.reload\\(\\)': 'RELOAD()',
      'await page\\.reload\\(\\{[^}]*hardReload:\\s*true[^}]*\\}\\)': 'RELOAD_HARD()',

      'page\\.url\\(\\)': 'GET_URL()',
      'await page\\.title\\(\\)': 'GET_TITLE()'
    },

    // Generate Playwright code from abstract operations
    generators: {
      'NAVIGATE': (url) => `await page.goto(${url})`,
      'GO_BACK': () => `await page.goBack()`,
      'GO_FORWARD': () => `await page.goForward()`,
      'GO': (delta) => delta < 0 ? `await page.goBack()` : `await page.goForward()`,
      'RELOAD': () => `await page.reload()`,
      'RELOAD_HARD': () => `await page.reload({ hardReload: true })`,
      'GET_URL': () => `page.url()`,
      'GET_LOCATION': () => `new URL(page.url())`,
      'GET_PATHNAME': () => `new URL(page.url()).pathname`,
      'GET_SEARCH': () => `new URL(page.url()).search`,
      'GET_HASH': () => `new URL(page.url()).hash`,
      'GET_TITLE': () => `await page.title()`
    }
  },

  selenium: {
    // Patterns to convert FROM Selenium
    patterns: {
      'await driver\\.get\\(([^)]+)\\)': 'NAVIGATE($1)',
      'await driver\\.navigate\\(\\)\\.to\\(([^)]+)\\)': 'NAVIGATE($1)',
      'await driver\\.navigate\\(\\)\\.back\\(\\)': 'GO_BACK()',
      'await driver\\.navigate\\(\\)\\.forward\\(\\)': 'GO_FORWARD()',
      'await driver\\.navigate\\(\\)\\.refresh\\(\\)': 'RELOAD()',

      'await driver\\.getCurrentUrl\\(\\)': 'GET_URL()',
      'await driver\\.getTitle\\(\\)': 'GET_TITLE()'
    },

    // Generate Selenium code from abstract operations
    generators: {
      'NAVIGATE': (url) => `await driver.get(${url})`,
      'GO_BACK': () => `await driver.navigate().back()`,
      'GO_FORWARD': () => `await driver.navigate().forward()`,
      'GO': (delta) => delta < 0 ? `await driver.navigate().back()` : `await driver.navigate().forward()`,
      'RELOAD': () => `await driver.navigate().refresh()`,
      'RELOAD_HARD': () => `await driver.navigate().refresh()`,
      'GET_URL': () => `await driver.getCurrentUrl()`,
      'GET_LOCATION': () => `new URL(await driver.getCurrentUrl())`,
      'GET_PATHNAME': () => `new URL(await driver.getCurrentUrl()).pathname`,
      'GET_SEARCH': () => `new URL(await driver.getCurrentUrl()).search`,
      'GET_HASH': () => `new URL(await driver.getCurrentUrl()).hash`,
      'GET_TITLE': () => `await driver.getTitle()`
    }
  }
};

/**
 * Direct conversion mappings between frameworks
 */
export const directMappings = {
  'cypress-playwright': {
    'cy\\.visit\\(([^)]+)\\)': 'await page.goto($1)',
    'cy\\.go\\([\'"]back[\'"]\\)': 'await page.goBack()',
    'cy\\.go\\([\'"]forward[\'"]\\)': 'await page.goForward()',
    'cy\\.reload\\(\\)': 'await page.reload()',
    'cy\\.url\\(\\)': 'page.url()',
    'cy\\.title\\(\\)': 'await page.title()'
  },

  'cypress-selenium': {
    'cy\\.visit\\(([^)]+)\\)': 'await driver.get($1)',
    'cy\\.go\\([\'"]back[\'"]\\)': 'await driver.navigate().back()',
    'cy\\.go\\([\'"]forward[\'"]\\)': 'await driver.navigate().forward()',
    'cy\\.reload\\(\\)': 'await driver.navigate().refresh()',
    'cy\\.url\\(\\)': 'await driver.getCurrentUrl()',
    'cy\\.title\\(\\)': 'await driver.getTitle()'
  },

  'playwright-cypress': {
    'await page\\.goto\\(([^)]+)\\)': 'cy.visit($1)',
    'await page\\.goBack\\(\\)': "cy.go('back')",
    'await page\\.goForward\\(\\)': "cy.go('forward')",
    'await page\\.reload\\(\\)': 'cy.reload()',
    'page\\.url\\(\\)': 'cy.url()',
    'await page\\.title\\(\\)': 'cy.title()'
  },

  'playwright-selenium': {
    'await page\\.goto\\(([^)]+)\\)': 'await driver.get($1)',
    'await page\\.goBack\\(\\)': 'await driver.navigate().back()',
    'await page\\.goForward\\(\\)': 'await driver.navigate().forward()',
    'await page\\.reload\\(\\)': 'await driver.navigate().refresh()',
    'page\\.url\\(\\)': 'await driver.getCurrentUrl()',
    'await page\\.title\\(\\)': 'await driver.getTitle()'
  },

  'selenium-cypress': {
    'await driver\\.get\\(([^)]+)\\)': 'cy.visit($1)',
    'await driver\\.navigate\\(\\)\\.back\\(\\)': "cy.go('back')",
    'await driver\\.navigate\\(\\)\\.forward\\(\\)': "cy.go('forward')",
    'await driver\\.navigate\\(\\)\\.refresh\\(\\)': 'cy.reload()',
    'await driver\\.getCurrentUrl\\(\\)': 'cy.url()',
    'await driver\\.getTitle\\(\\)': 'cy.title()'
  },

  'selenium-playwright': {
    'await driver\\.get\\(([^)]+)\\)': 'await page.goto($1)',
    'await driver\\.navigate\\(\\)\\.back\\(\\)': 'await page.goBack()',
    'await driver\\.navigate\\(\\)\\.forward\\(\\)': 'await page.goForward()',
    'await driver\\.navigate\\(\\)\\.refresh\\(\\)': 'await page.reload()',
    'await driver\\.getCurrentUrl\\(\\)': 'page.url()',
    'await driver\\.getTitle\\(\\)': 'await page.title()'
  }
};
