/**
 * Selector patterns for all three frameworks
 * Handles element selection, finding, and traversal
 */

export const selectorPatterns = {
  cypress: {
    patterns: {
      // Basic selection
      'cy\\.get\\(([^)]+)\\)': 'SELECT($1)',
      'cy\\.find\\(([^)]+)\\)': 'FIND($1)',
      'cy\\.contains\\(([^)]+)\\)': 'SELECT_BY_TEXT($1)',
      'cy\\.contains\\(([^,]+),\\s*([^)]+)\\)': 'SELECT_BY_TEXT_IN($1, $2)',

      // Traversal
      '\\.first\\(\\)': '.FIRST()',
      '\\.last\\(\\)': '.LAST()',
      '\\.eq\\((\\d+)\\)': '.NTH($1)',
      '\\.parent\\(\\)': '.PARENT()',
      '\\.parents\\(([^)]+)\\)': '.PARENTS($1)',
      '\\.children\\(\\)': '.CHILDREN()',
      '\\.children\\(([^)]+)\\)': '.CHILDREN($1)',
      '\\.siblings\\(\\)': '.SIBLINGS()',
      '\\.siblings\\(([^)]+)\\)': '.SIBLINGS($1)',
      '\\.closest\\(([^)]+)\\)': '.CLOSEST($1)',
      '\\.next\\(\\)': '.NEXT()',
      '\\.prev\\(\\)': '.PREV()',
      '\\.filter\\(([^)]+)\\)': '.FILTER($1)',
      '\\.not\\(([^)]+)\\)': '.NOT($1)',

      // Special selectors
      'cy\\.focused\\(\\)': 'SELECT_FOCUSED()',
      'cy\\.root\\(\\)': 'SELECT_ROOT()',
      'cy\\.document\\(\\)': 'SELECT_DOCUMENT()',
      'cy\\.window\\(\\)': 'SELECT_WINDOW()',
    },

    generators: {
      SELECT: (sel) => `cy.get(${sel})`,
      FIND: (sel) => `.find(${sel})`,
      SELECT_BY_TEXT: (text) => `cy.contains(${text})`,
      SELECT_BY_TEXT_IN: (sel, text) => `cy.contains(${sel}, ${text})`,
      FIRST: () => '.first()',
      LAST: () => '.last()',
      NTH: (n) => `.eq(${n})`,
      PARENT: () => '.parent()',
      PARENTS: (sel) => `.parents(${sel})`,
      CHILDREN: (sel) => (sel ? `.children(${sel})` : '.children()'),
      SIBLINGS: (sel) => (sel ? `.siblings(${sel})` : '.siblings()'),
      CLOSEST: (sel) => `.closest(${sel})`,
      NEXT: () => '.next()',
      PREV: () => '.prev()',
      FILTER: (sel) => `.filter(${sel})`,
      NOT: (sel) => `.not(${sel})`,
      SELECT_FOCUSED: () => 'cy.focused()',
      SELECT_ROOT: () => 'cy.root()',
      SELECT_DOCUMENT: () => 'cy.document()',
      SELECT_WINDOW: () => 'cy.window()',
    },
  },

  playwright: {
    patterns: {
      'page\\.locator\\(([^)]+)\\)': 'SELECT($1)',
      'page\\.getByText\\(([^)]+)\\)': 'SELECT_BY_TEXT($1)',
      'page\\.getByRole\\(([^)]+)\\)': 'SELECT_BY_ROLE($1)',
      'page\\.getByLabel\\(([^)]+)\\)': 'SELECT_BY_LABEL($1)',
      'page\\.getByPlaceholder\\(([^)]+)\\)': 'SELECT_BY_PLACEHOLDER($1)',
      'page\\.getByTestId\\(([^)]+)\\)': 'SELECT_BY_TESTID($1)',
      'page\\.getByTitle\\(([^)]+)\\)': 'SELECT_BY_TITLE($1)',
      'page\\.getByAltText\\(([^)]+)\\)': 'SELECT_BY_ALT($1)',

      '\\.locator\\(([^)]+)\\)': '.FIND($1)',
      '\\.first\\(\\)': '.FIRST()',
      '\\.last\\(\\)': '.LAST()',
      '\\.nth\\((\\d+)\\)': '.NTH($1)',
      '\\.filter\\(([^)]+)\\)': '.FILTER($1)',
    },

    generators: {
      SELECT: (sel) => `page.locator(${sel})`,
      FIND: (sel) => `.locator(${sel})`,
      SELECT_BY_TEXT: (text) => `page.getByText(${text})`,
      SELECT_BY_TEXT_IN: (sel, text) =>
        `page.locator(${sel}).filter({ hasText: ${text} })`,
      SELECT_BY_ROLE: (role) => `page.getByRole(${role})`,
      SELECT_BY_LABEL: (label) => `page.getByLabel(${label})`,
      SELECT_BY_PLACEHOLDER: (placeholder) =>
        `page.getByPlaceholder(${placeholder})`,
      SELECT_BY_TESTID: (testId) => `page.getByTestId(${testId})`,
      SELECT_BY_TITLE: (title) => `page.getByTitle(${title})`,
      SELECT_BY_ALT: (alt) => `page.getByAltText(${alt})`,
      FIRST: () => '.first()',
      LAST: () => '.last()',
      NTH: (n) => `.nth(${n})`,
      PARENT: () => '.locator("..")',
      PARENTS: (sel) => `.locator(${sel})`,
      CHILDREN: (sel) => (sel ? `.locator(${sel})` : '.locator("> *")'),
      SIBLINGS: () => '.locator("~ *")',
      CLOSEST: (sel) => `.locator(${sel})`,
      NEXT: () => '.locator("+ *")',
      PREV: () => '.locator(":prev")',
      FILTER: (opts) => `.filter(${opts})`,
      NOT: (sel) => `.filter({ hasNot: page.locator(${sel}) })`,
      SELECT_FOCUSED: () => 'page.locator(":focus")',
      SELECT_ROOT: () => 'page.locator("html")',
      SELECT_DOCUMENT: () => 'page',
      SELECT_WINDOW: () => 'page',
    },
  },

  selenium: {
    patterns: {
      'await driver\\.findElement\\(By\\.css\\(([^)]+)\\)\\)': 'SELECT($1)',
      'await driver\\.findElement\\(By\\.xpath\\(([^)]+)\\)\\)':
        'SELECT_XPATH($1)',
      'await driver\\.findElement\\(By\\.id\\(([^)]+)\\)\\)':
        'SELECT_BY_ID($1)',
      'await driver\\.findElement\\(By\\.name\\(([^)]+)\\)\\)':
        'SELECT_BY_NAME($1)',
      'await driver\\.findElement\\(By\\.className\\(([^)]+)\\)\\)':
        'SELECT_BY_CLASS($1)',
      'await driver\\.findElement\\(By\\.tagName\\(([^)]+)\\)\\)':
        'SELECT_BY_TAG($1)',
      'await driver\\.findElement\\(By\\.linkText\\(([^)]+)\\)\\)':
        'SELECT_BY_LINK_TEXT($1)',
      'await driver\\.findElement\\(By\\.partialLinkText\\(([^)]+)\\)\\)':
        'SELECT_BY_PARTIAL_LINK_TEXT($1)',
      'await driver\\.findElements\\(([^)]+)\\)': 'SELECT_ALL($1)',
      '\\.findElement\\(([^)]+)\\)': '.FIND($1)',
    },

    generators: {
      SELECT: (sel) => `await driver.findElement(By.css(${sel}))`,
      SELECT_XPATH: (xpath) => `await driver.findElement(By.xpath(${xpath}))`,
      SELECT_BY_ID: (id) => `await driver.findElement(By.id(${id}))`,
      SELECT_BY_NAME: (name) => `await driver.findElement(By.name(${name}))`,
      SELECT_BY_CLASS: (className) =>
        `await driver.findElement(By.className(${className}))`,
      SELECT_BY_TAG: (tag) => `await driver.findElement(By.tagName(${tag}))`,
      SELECT_BY_LINK_TEXT: (text) =>
        `await driver.findElement(By.linkText(${text}))`,
      SELECT_BY_PARTIAL_LINK_TEXT: (text) =>
        `await driver.findElement(By.partialLinkText(${text}))`,
      SELECT_ALL: (sel) => `await driver.findElements(${sel})`,
      FIND: (sel) => `.findElement(${sel})`,
      SELECT_BY_TEXT: (text) =>
        `await driver.findElement(By.xpath(\`//*[contains(text(),${text})]\`))`,
      SELECT_BY_TEXT_IN: (sel, text) =>
        `await driver.findElement(By.xpath(\`${sel}[contains(text(),${text})]\`))`,
      FIRST: () => '[0]',
      LAST: () => '.slice(-1)[0]',
      NTH: (n) => `[${n}]`,
      PARENT: () => '.findElement(By.xpath(".."))',
      PARENTS: (sel) => `.findElement(By.xpath("ancestor::${sel}"))`,
      CHILDREN: (sel) =>
        sel
          ? `.findElements(By.css(${sel}))`
          : '.findElements(By.xpath("child::*"))',
      SIBLINGS: () =>
        '.findElements(By.xpath("following-sibling::* | preceding-sibling::*"))',
      CLOSEST: (sel) => `.findElement(By.xpath("ancestor-or-self::${sel}"))`,
      NEXT: () => '.findElement(By.xpath("following-sibling::*[1]"))',
      PREV: () => '.findElement(By.xpath("preceding-sibling::*[1]"))',
      FILTER: () => '', // Requires manual filtering in Selenium
      NOT: () => '', // Requires manual filtering in Selenium
      SELECT_FOCUSED: () => 'await driver.switchTo().activeElement()',
      SELECT_ROOT: () => 'await driver.findElement(By.css("html"))',
      SELECT_DOCUMENT: () => 'driver',
      SELECT_WINDOW: () => 'driver',
    },
  },
};

/**
 * Direct conversion mappings between frameworks
 */
export const directMappings = {
  'cypress-playwright': {
    'cy\\.get\\(([^)]+)\\)': 'page.locator($1)',
    'cy\\.contains\\(([^)]+)\\)': 'page.getByText($1)',
    '\\.find\\(([^)]+)\\)': '.locator($1)',
    '\\.first\\(\\)': '.first()',
    '\\.last\\(\\)': '.last()',
    '\\.eq\\((\\d+)\\)': '.nth($1)',
    'cy\\.focused\\(\\)': 'page.locator(":focus")',
  },

  'cypress-selenium': {
    'cy\\.get\\(([^)]+)\\)': 'await driver.findElement(By.css($1))',
    'cy\\.contains\\(([^)]+)\\)':
      'await driver.findElement(By.xpath(`//*[contains(text(),$1)]`))',
    '\\.find\\(([^)]+)\\)': '.findElement(By.css($1))',
    '\\.first\\(\\)': '[0]',
    '\\.last\\(\\)': '.slice(-1)[0]',
    '\\.eq\\((\\d+)\\)': '[$1]',
    'cy\\.focused\\(\\)': 'await driver.switchTo().activeElement()',
  },

  'playwright-cypress': {
    'page\\.locator\\(([^)]+)\\)': 'cy.get($1)',
    'page\\.getByText\\(([^)]+)\\)': 'cy.contains($1)',
    'page\\.getByTestId\\(([^)]+)\\)': 'cy.get(`[data-testid=${$1}]`)',
    '\\.locator\\(([^)]+)\\)': '.find($1)',
    '\\.first\\(\\)': '.first()',
    '\\.last\\(\\)': '.last()',
    '\\.nth\\((\\d+)\\)': '.eq($1)',
  },

  'playwright-selenium': {
    'page\\.locator\\(([^)]+)\\)': 'await driver.findElement(By.css($1))',
    'page\\.getByText\\(([^)]+)\\)':
      'await driver.findElement(By.xpath(`//*[contains(text(),$1)]`))',
    '\\.locator\\(([^)]+)\\)': '.findElement(By.css($1))',
    '\\.first\\(\\)': '[0]',
    '\\.last\\(\\)': '.slice(-1)[0]',
    '\\.nth\\((\\d+)\\)': '[$1]',
  },

  'selenium-cypress': {
    'await driver\\.findElement\\(By\\.css\\(([^)]+)\\)\\)': 'cy.get($1)',
    'await driver\\.findElement\\(By\\.id\\(([^)]+)\\)\\)': 'cy.get(`#${$1}`)',
    'await driver\\.findElement\\(By\\.xpath\\(([^)]+)\\)\\)': 'cy.xpath($1)',
    '\\.findElement\\(By\\.css\\(([^)]+)\\)\\)': '.find($1)',
    'await driver\\.switchTo\\(\\)\\.activeElement\\(\\)': 'cy.focused()',
  },

  'selenium-playwright': {
    'await driver\\.findElement\\(By\\.css\\(([^)]+)\\)\\)': 'page.locator($1)',
    'await driver\\.findElement\\(By\\.id\\(([^)]+)\\)\\)':
      'page.locator(`#${$1}`)',
    'await driver\\.findElement\\(By\\.xpath\\(([^)]+)\\)\\)':
      'page.locator($1)',
    '\\.findElement\\(By\\.css\\(([^)]+)\\)\\)': '.locator($1)',
    'await driver\\.switchTo\\(\\)\\.activeElement\\(\\)':
      'page.locator(":focus")',
  },
};
