/**
 * Assertion patterns for all three frameworks
 * Handles visibility, text, attribute, and state assertions
 */

export const assertionPatterns = {
  cypress: {
    patterns: {
      // Visibility
      '\\.should\\([\'"]be\\.visible[\'"]\\)': '.ASSERT_VISIBLE()',
      '\\.should\\([\'"]not\\.be\\.visible[\'"]\\)': '.ASSERT_HIDDEN()',
      '\\.should\\([\'"]exist[\'"]\\)': '.ASSERT_EXISTS()',
      '\\.should\\([\'"]not\\.exist[\'"]\\)': '.ASSERT_NOT_EXISTS()',

      // Text
      '\\.should\\([\'"]have\\.text[\'"],\\s*([^)]+)\\)': '.ASSERT_TEXT($1)',
      '\\.should\\([\'"]contain[\'"],\\s*([^)]+)\\)': '.ASSERT_CONTAINS($1)',
      '\\.should\\([\'"]contain\\.text[\'"],\\s*([^)]+)\\)': '.ASSERT_CONTAINS($1)',
      '\\.should\\([\'"]not\\.contain[\'"],\\s*([^)]+)\\)': '.ASSERT_NOT_CONTAINS($1)',
      '\\.should\\([\'"]include\\.text[\'"],\\s*([^)]+)\\)': '.ASSERT_CONTAINS($1)',
      '\\.should\\([\'"]be\\.empty[\'"]\\)': '.ASSERT_EMPTY()',
      '\\.should\\([\'"]not\\.be\\.empty[\'"]\\)': '.ASSERT_NOT_EMPTY()',

      // Value
      '\\.should\\([\'"]have\\.value[\'"],\\s*([^)]+)\\)': '.ASSERT_VALUE($1)',

      // Attributes
      '\\.should\\([\'"]have\\.attr[\'"],\\s*([^,]+),\\s*([^)]+)\\)': '.ASSERT_ATTRIBUTE($1, $2)',
      '\\.should\\([\'"]have\\.attr[\'"],\\s*([^)]+)\\)': '.ASSERT_HAS_ATTRIBUTE($1)',
      '\\.should\\([\'"]have\\.class[\'"],\\s*([^)]+)\\)': '.ASSERT_CLASS($1)',
      '\\.should\\([\'"]not\\.have\\.class[\'"],\\s*([^)]+)\\)': '.ASSERT_NOT_CLASS($1)',
      '\\.should\\([\'"]have\\.css[\'"],\\s*([^,]+),\\s*([^)]+)\\)': '.ASSERT_CSS($1, $2)',
      '\\.should\\([\'"]have\\.id[\'"],\\s*([^)]+)\\)': '.ASSERT_ID($1)',

      // State
      '\\.should\\([\'"]be\\.checked[\'"]\\)': '.ASSERT_CHECKED()',
      '\\.should\\([\'"]not\\.be\\.checked[\'"]\\)': '.ASSERT_NOT_CHECKED()',
      '\\.should\\([\'"]be\\.disabled[\'"]\\)': '.ASSERT_DISABLED()',
      '\\.should\\([\'"]be\\.enabled[\'"]\\)': '.ASSERT_ENABLED()',
      '\\.should\\([\'"]not\\.be\\.disabled[\'"]\\)': '.ASSERT_ENABLED()',
      '\\.should\\([\'"]be\\.focused[\'"]\\)': '.ASSERT_FOCUSED()',
      '\\.should\\([\'"]be\\.selected[\'"]\\)': '.ASSERT_SELECTED()',

      // Count
      '\\.should\\([\'"]have\\.length[\'"],\\s*([^)]+)\\)': '.ASSERT_COUNT($1)',
      '\\.should\\([\'"]have\\.length\\.gt[\'"],\\s*([^)]+)\\)': '.ASSERT_COUNT_GT($1)',
      '\\.should\\([\'"]have\\.length\\.gte[\'"],\\s*([^)]+)\\)': '.ASSERT_COUNT_GTE($1)',
      '\\.should\\([\'"]have\\.length\\.lt[\'"],\\s*([^)]+)\\)': '.ASSERT_COUNT_LT($1)',
      '\\.should\\([\'"]have\\.length\\.lte[\'"],\\s*([^)]+)\\)': '.ASSERT_COUNT_LTE($1)',

      // URL assertions
      'cy\\.url\\(\\)\\.should\\([\'"]include[\'"],\\s*([^)]+)\\)': 'ASSERT_URL_CONTAINS($1)',
      'cy\\.url\\(\\)\\.should\\([\'"]eq[\'"],\\s*([^)]+)\\)': 'ASSERT_URL_EQUALS($1)',
      'cy\\.title\\(\\)\\.should\\([\'"]eq[\'"],\\s*([^)]+)\\)': 'ASSERT_TITLE($1)'
    },

    generators: {
      'ASSERT_VISIBLE': () => '.should("be.visible")',
      'ASSERT_HIDDEN': () => '.should("not.be.visible")',
      'ASSERT_EXISTS': () => '.should("exist")',
      'ASSERT_NOT_EXISTS': () => '.should("not.exist")',
      'ASSERT_TEXT': (text) => `.should("have.text", ${text})`,
      'ASSERT_CONTAINS': (text) => `.should("contain", ${text})`,
      'ASSERT_NOT_CONTAINS': (text) => `.should("not.contain", ${text})`,
      'ASSERT_EMPTY': () => '.should("be.empty")',
      'ASSERT_NOT_EMPTY': () => '.should("not.be.empty")',
      'ASSERT_VALUE': (value) => `.should("have.value", ${value})`,
      'ASSERT_ATTRIBUTE': (attr, value) => `.should("have.attr", ${attr}, ${value})`,
      'ASSERT_HAS_ATTRIBUTE': (attr) => `.should("have.attr", ${attr})`,
      'ASSERT_CLASS': (cls) => `.should("have.class", ${cls})`,
      'ASSERT_NOT_CLASS': (cls) => `.should("not.have.class", ${cls})`,
      'ASSERT_CSS': (prop, value) => `.should("have.css", ${prop}, ${value})`,
      'ASSERT_ID': (id) => `.should("have.id", ${id})`,
      'ASSERT_CHECKED': () => '.should("be.checked")',
      'ASSERT_NOT_CHECKED': () => '.should("not.be.checked")',
      'ASSERT_DISABLED': () => '.should("be.disabled")',
      'ASSERT_ENABLED': () => '.should("be.enabled")',
      'ASSERT_FOCUSED': () => '.should("be.focused")',
      'ASSERT_SELECTED': () => '.should("be.selected")',
      'ASSERT_COUNT': (n) => `.should("have.length", ${n})`,
      'ASSERT_COUNT_GT': (n) => `.should("have.length.gt", ${n})`,
      'ASSERT_COUNT_GTE': (n) => `.should("have.length.gte", ${n})`,
      'ASSERT_COUNT_LT': (n) => `.should("have.length.lt", ${n})`,
      'ASSERT_COUNT_LTE': (n) => `.should("have.length.lte", ${n})`,
      'ASSERT_URL_CONTAINS': (text) => `cy.url().should("include", ${text})`,
      'ASSERT_URL_EQUALS': (url) => `cy.url().should("eq", ${url})`,
      'ASSERT_TITLE': (title) => `cy.title().should("eq", ${title})`
    }
  },

  playwright: {
    patterns: {
      // Visibility
      'await expect\\(([^)]+)\\)\\.toBeVisible\\(\\)': '$1.ASSERT_VISIBLE()',
      'await expect\\(([^)]+)\\)\\.toBeHidden\\(\\)': '$1.ASSERT_HIDDEN()',
      'await expect\\(([^)]+)\\)\\.not\\.toBeVisible\\(\\)': '$1.ASSERT_HIDDEN()',
      'await expect\\(([^)]+)\\)\\.toBeAttached\\(\\)': '$1.ASSERT_EXISTS()',
      'await expect\\(([^)]+)\\)\\.not\\.toBeAttached\\(\\)': '$1.ASSERT_NOT_EXISTS()',

      // Text
      'await expect\\(([^)]+)\\)\\.toHaveText\\(([^)]+)\\)': '$1.ASSERT_TEXT($2)',
      'await expect\\(([^)]+)\\)\\.toContainText\\(([^)]+)\\)': '$1.ASSERT_CONTAINS($2)',
      'await expect\\(([^)]+)\\)\\.not\\.toContainText\\(([^)]+)\\)': '$1.ASSERT_NOT_CONTAINS($2)',
      'await expect\\(([^)]+)\\)\\.toBeEmpty\\(\\)': '$1.ASSERT_EMPTY()',

      // Value
      'await expect\\(([^)]+)\\)\\.toHaveValue\\(([^)]+)\\)': '$1.ASSERT_VALUE($2)',

      // Attributes
      'await expect\\(([^)]+)\\)\\.toHaveAttribute\\(([^,]+),\\s*([^)]+)\\)': '$1.ASSERT_ATTRIBUTE($2, $3)',
      'await expect\\(([^)]+)\\)\\.toHaveAttribute\\(([^)]+)\\)': '$1.ASSERT_HAS_ATTRIBUTE($2)',
      'await expect\\(([^)]+)\\)\\.toHaveClass\\(([^)]+)\\)': '$1.ASSERT_CLASS($2)',
      'await expect\\(([^)]+)\\)\\.toHaveCSS\\(([^,]+),\\s*([^)]+)\\)': '$1.ASSERT_CSS($2, $3)',
      'await expect\\(([^)]+)\\)\\.toHaveId\\(([^)]+)\\)': '$1.ASSERT_ID($2)',

      // State
      'await expect\\(([^)]+)\\)\\.toBeChecked\\(\\)': '$1.ASSERT_CHECKED()',
      'await expect\\(([^)]+)\\)\\.not\\.toBeChecked\\(\\)': '$1.ASSERT_NOT_CHECKED()',
      'await expect\\(([^)]+)\\)\\.toBeDisabled\\(\\)': '$1.ASSERT_DISABLED()',
      'await expect\\(([^)]+)\\)\\.toBeEnabled\\(\\)': '$1.ASSERT_ENABLED()',
      'await expect\\(([^)]+)\\)\\.toBeFocused\\(\\)': '$1.ASSERT_FOCUSED()',

      // Count
      'await expect\\(([^)]+)\\)\\.toHaveCount\\(([^)]+)\\)': '$1.ASSERT_COUNT($2)',

      // Page
      'await expect\\(page\\)\\.toHaveURL\\(([^)]+)\\)': 'ASSERT_URL_EQUALS($1)',
      'await expect\\(page\\)\\.toHaveTitle\\(([^)]+)\\)': 'ASSERT_TITLE($1)'
    },

    generators: {
      'ASSERT_VISIBLE': (el) => `await expect(${el}).toBeVisible()`,
      'ASSERT_HIDDEN': (el) => `await expect(${el}).toBeHidden()`,
      'ASSERT_EXISTS': (el) => `await expect(${el}).toBeAttached()`,
      'ASSERT_NOT_EXISTS': (el) => `await expect(${el}).not.toBeAttached()`,
      'ASSERT_TEXT': (el, text) => `await expect(${el}).toHaveText(${text})`,
      'ASSERT_CONTAINS': (el, text) => `await expect(${el}).toContainText(${text})`,
      'ASSERT_NOT_CONTAINS': (el, text) => `await expect(${el}).not.toContainText(${text})`,
      'ASSERT_EMPTY': (el) => `await expect(${el}).toBeEmpty()`,
      'ASSERT_NOT_EMPTY': (el) => `await expect(${el}).not.toBeEmpty()`,
      'ASSERT_VALUE': (el, value) => `await expect(${el}).toHaveValue(${value})`,
      'ASSERT_ATTRIBUTE': (el, attr, value) => `await expect(${el}).toHaveAttribute(${attr}, ${value})`,
      'ASSERT_HAS_ATTRIBUTE': (el, attr) => `await expect(${el}).toHaveAttribute(${attr})`,
      'ASSERT_CLASS': (el, cls) => `await expect(${el}).toHaveClass(${cls})`,
      'ASSERT_NOT_CLASS': (el, cls) => `await expect(${el}).not.toHaveClass(${cls})`,
      'ASSERT_CSS': (el, prop, value) => `await expect(${el}).toHaveCSS(${prop}, ${value})`,
      'ASSERT_ID': (el, id) => `await expect(${el}).toHaveId(${id})`,
      'ASSERT_CHECKED': (el) => `await expect(${el}).toBeChecked()`,
      'ASSERT_NOT_CHECKED': (el) => `await expect(${el}).not.toBeChecked()`,
      'ASSERT_DISABLED': (el) => `await expect(${el}).toBeDisabled()`,
      'ASSERT_ENABLED': (el) => `await expect(${el}).toBeEnabled()`,
      'ASSERT_FOCUSED': (el) => `await expect(${el}).toBeFocused()`,
      'ASSERT_SELECTED': (el) => `await expect(${el}).toBeChecked()`,
      'ASSERT_COUNT': (el, n) => `await expect(${el}).toHaveCount(${n})`,
      'ASSERT_COUNT_GT': (el, n) => `expect(await ${el}.count()).toBeGreaterThan(${n})`,
      'ASSERT_COUNT_GTE': (el, n) => `expect(await ${el}.count()).toBeGreaterThanOrEqual(${n})`,
      'ASSERT_COUNT_LT': (el, n) => `expect(await ${el}.count()).toBeLessThan(${n})`,
      'ASSERT_COUNT_LTE': (el, n) => `expect(await ${el}.count()).toBeLessThanOrEqual(${n})`,
      'ASSERT_URL_CONTAINS': (text) => `expect(page.url()).toContain(${text})`,
      'ASSERT_URL_EQUALS': (url) => `await expect(page).toHaveURL(${url})`,
      'ASSERT_TITLE': (title) => `await expect(page).toHaveTitle(${title})`
    }
  },

  selenium: {
    patterns: {
      // These are typically Jest/Mocha expect patterns used with Selenium
      'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(true\\)': '$1.ASSERT_VISIBLE()',
      'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(false\\)': '$1.ASSERT_HIDDEN()',
      'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toBe\\(([^)]+)\\)': '$1.ASSERT_TEXT($2)',
      'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toContain\\(([^)]+)\\)': '$1.ASSERT_CONTAINS($2)',
      'expect\\(await ([^.]+)\\.getAttribute\\([\'"]value[\'"]\\)\\)\\.toBe\\(([^)]+)\\)': '$1.ASSERT_VALUE($2)',
      'expect\\(await ([^.]+)\\.getAttribute\\(([^)]+)\\)\\)\\.toBe\\(([^)]+)\\)': '$1.ASSERT_ATTRIBUTE($2, $3)',
      'expect\\(await ([^.]+)\\.isSelected\\(\\)\\)\\.toBe\\(true\\)': '$1.ASSERT_CHECKED()',
      'expect\\(await ([^.]+)\\.isSelected\\(\\)\\)\\.toBe\\(false\\)': '$1.ASSERT_NOT_CHECKED()',
      'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(false\\)': '$1.ASSERT_DISABLED()',
      'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(true\\)': '$1.ASSERT_ENABLED()'
    },

    generators: {
      'ASSERT_VISIBLE': (el) => `expect(await ${el}.isDisplayed()).toBe(true)`,
      'ASSERT_HIDDEN': (el) => `expect(await ${el}.isDisplayed()).toBe(false)`,
      'ASSERT_EXISTS': (el) => `expect(${el}).toBeTruthy()`,
      'ASSERT_NOT_EXISTS': (el) => `expect(${el}).toBeFalsy()`,
      'ASSERT_TEXT': (el, text) => `expect(await ${el}.getText()).toBe(${text})`,
      'ASSERT_CONTAINS': (el, text) => `expect(await ${el}.getText()).toContain(${text})`,
      'ASSERT_NOT_CONTAINS': (el, text) => `expect(await ${el}.getText()).not.toContain(${text})`,
      'ASSERT_EMPTY': (el) => `expect(await ${el}.getText()).toBe("")`,
      'ASSERT_NOT_EMPTY': (el) => `expect(await ${el}.getText()).not.toBe("")`,
      'ASSERT_VALUE': (el, value) => `expect(await ${el}.getAttribute("value")).toBe(${value})`,
      'ASSERT_ATTRIBUTE': (el, attr, value) => `expect(await ${el}.getAttribute(${attr})).toBe(${value})`,
      'ASSERT_HAS_ATTRIBUTE': (el, attr) => `expect(await ${el}.getAttribute(${attr})).not.toBeNull()`,
      'ASSERT_CLASS': (el, cls) => `expect(await ${el}.getAttribute("class")).toContain(${cls})`,
      'ASSERT_NOT_CLASS': (el, cls) => `expect(await ${el}.getAttribute("class")).not.toContain(${cls})`,
      'ASSERT_CSS': (el, prop, value) => `expect(await ${el}.getCssValue(${prop})).toBe(${value})`,
      'ASSERT_ID': (el, id) => `expect(await ${el}.getAttribute("id")).toBe(${id})`,
      'ASSERT_CHECKED': (el) => `expect(await ${el}.isSelected()).toBe(true)`,
      'ASSERT_NOT_CHECKED': (el) => `expect(await ${el}.isSelected()).toBe(false)`,
      'ASSERT_DISABLED': (el) => `expect(await ${el}.isEnabled()).toBe(false)`,
      'ASSERT_ENABLED': (el) => `expect(await ${el}.isEnabled()).toBe(true)`,
      'ASSERT_FOCUSED': (el) => `expect(await driver.switchTo().activeElement()).toEqual(${el})`,
      'ASSERT_SELECTED': (el) => `expect(await ${el}.isSelected()).toBe(true)`,
      'ASSERT_COUNT': (els, n) => `expect(${els}.length).toBe(${n})`,
      'ASSERT_COUNT_GT': (els, n) => `expect(${els}.length).toBeGreaterThan(${n})`,
      'ASSERT_COUNT_GTE': (els, n) => `expect(${els}.length).toBeGreaterThanOrEqual(${n})`,
      'ASSERT_COUNT_LT': (els, n) => `expect(${els}.length).toBeLessThan(${n})`,
      'ASSERT_COUNT_LTE': (els, n) => `expect(${els}.length).toBeLessThanOrEqual(${n})`,
      'ASSERT_URL_CONTAINS': (text) => `expect(await driver.getCurrentUrl()).toContain(${text})`,
      'ASSERT_URL_EQUALS': (url) => `expect(await driver.getCurrentUrl()).toBe(${url})`,
      'ASSERT_TITLE': (title) => `expect(await driver.getTitle()).toBe(${title})`
    }
  }
};

/**
 * Direct conversion mappings between frameworks
 */
export const directMappings = {
  'cypress-playwright': {
    '\\.should\\([\'"]be\\.visible[\'"]\\)': '; await expect(element).toBeVisible()',
    '\\.should\\([\'"]not\\.be\\.visible[\'"]\\)': '; await expect(element).toBeHidden()',
    '\\.should\\([\'"]exist[\'"]\\)': '; await expect(element).toBeAttached()',
    '\\.should\\([\'"]not\\.exist[\'"]\\)': '; await expect(element).not.toBeAttached()',
    '\\.should\\([\'"]have\\.text[\'"],\\s*([^)]+)\\)': '; await expect(element).toHaveText($1)',
    '\\.should\\([\'"]contain[\'"],\\s*([^)]+)\\)': '; await expect(element).toContainText($1)',
    '\\.should\\([\'"]have\\.value[\'"],\\s*([^)]+)\\)': '; await expect(element).toHaveValue($1)',
    '\\.should\\([\'"]have\\.attr[\'"],\\s*([^,]+),\\s*([^)]+)\\)': '; await expect(element).toHaveAttribute($1, $2)',
    '\\.should\\([\'"]have\\.class[\'"],\\s*([^)]+)\\)': '; await expect(element).toHaveClass($1)',
    '\\.should\\([\'"]be\\.checked[\'"]\\)': '; await expect(element).toBeChecked()',
    '\\.should\\([\'"]be\\.disabled[\'"]\\)': '; await expect(element).toBeDisabled()',
    '\\.should\\([\'"]be\\.enabled[\'"]\\)': '; await expect(element).toBeEnabled()',
    '\\.should\\([\'"]have\\.length[\'"],\\s*([^)]+)\\)': '; await expect(element).toHaveCount($1)'
  },

  'cypress-selenium': {
    '\\.should\\([\'"]be\\.visible[\'"]\\)': '; expect(await element.isDisplayed()).toBe(true)',
    '\\.should\\([\'"]not\\.be\\.visible[\'"]\\)': '; expect(await element.isDisplayed()).toBe(false)',
    '\\.should\\([\'"]have\\.text[\'"],\\s*([^)]+)\\)': '; expect(await element.getText()).toBe($1)',
    '\\.should\\([\'"]contain[\'"],\\s*([^)]+)\\)': '; expect(await element.getText()).toContain($1)',
    '\\.should\\([\'"]have\\.value[\'"],\\s*([^)]+)\\)': '; expect(await element.getAttribute("value")).toBe($1)',
    '\\.should\\([\'"]be\\.checked[\'"]\\)': '; expect(await element.isSelected()).toBe(true)',
    '\\.should\\([\'"]be\\.disabled[\'"]\\)': '; expect(await element.isEnabled()).toBe(false)',
    '\\.should\\([\'"]be\\.enabled[\'"]\\)': '; expect(await element.isEnabled()).toBe(true)'
  },

  'playwright-cypress': {
    'await expect\\(([^)]+)\\)\\.toBeVisible\\(\\)': '$1.should("be.visible")',
    'await expect\\(([^)]+)\\)\\.toBeHidden\\(\\)': '$1.should("not.be.visible")',
    'await expect\\(([^)]+)\\)\\.toHaveText\\(([^)]+)\\)': '$1.should("have.text", $2)',
    'await expect\\(([^)]+)\\)\\.toContainText\\(([^)]+)\\)': '$1.should("contain", $2)',
    'await expect\\(([^)]+)\\)\\.toHaveValue\\(([^)]+)\\)': '$1.should("have.value", $2)',
    'await expect\\(([^)]+)\\)\\.toHaveAttribute\\(([^,]+),\\s*([^)]+)\\)': '$1.should("have.attr", $2, $3)',
    'await expect\\(([^)]+)\\)\\.toHaveClass\\(([^)]+)\\)': '$1.should("have.class", $2)',
    'await expect\\(([^)]+)\\)\\.toBeChecked\\(\\)': '$1.should("be.checked")',
    'await expect\\(([^)]+)\\)\\.toBeDisabled\\(\\)': '$1.should("be.disabled")',
    'await expect\\(([^)]+)\\)\\.toBeEnabled\\(\\)': '$1.should("be.enabled")',
    'await expect\\(([^)]+)\\)\\.toHaveCount\\(([^)]+)\\)': '$1.should("have.length", $2)'
  },

  'playwright-selenium': {
    'await expect\\(([^)]+)\\)\\.toBeVisible\\(\\)': 'expect(await $1.isDisplayed()).toBe(true)',
    'await expect\\(([^)]+)\\)\\.toBeHidden\\(\\)': 'expect(await $1.isDisplayed()).toBe(false)',
    'await expect\\(([^)]+)\\)\\.toHaveText\\(([^)]+)\\)': 'expect(await $1.getText()).toBe($2)',
    'await expect\\(([^)]+)\\)\\.toContainText\\(([^)]+)\\)': 'expect(await $1.getText()).toContain($2)',
    'await expect\\(([^)]+)\\)\\.toHaveValue\\(([^)]+)\\)': 'expect(await $1.getAttribute("value")).toBe($2)',
    'await expect\\(([^)]+)\\)\\.toBeChecked\\(\\)': 'expect(await $1.isSelected()).toBe(true)',
    'await expect\\(([^)]+)\\)\\.toBeDisabled\\(\\)': 'expect(await $1.isEnabled()).toBe(false)',
    'await expect\\(([^)]+)\\)\\.toBeEnabled\\(\\)': 'expect(await $1.isEnabled()).toBe(true)'
  },

  'selenium-cypress': {
    'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(true\\)': '$1.should("be.visible")',
    'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(false\\)': '$1.should("not.be.visible")',
    'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toBe\\(([^)]+)\\)': '$1.should("have.text", $2)',
    'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toContain\\(([^)]+)\\)': '$1.should("contain", $2)',
    'expect\\(await ([^.]+)\\.getAttribute\\([\'"]value[\'"]\\)\\)\\.toBe\\(([^)]+)\\)': '$1.should("have.value", $2)',
    'expect\\(await ([^.]+)\\.isSelected\\(\\)\\)\\.toBe\\(true\\)': '$1.should("be.checked")',
    'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(false\\)': '$1.should("be.disabled")',
    'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(true\\)': '$1.should("be.enabled")'
  },

  'selenium-playwright': {
    'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(true\\)': 'await expect($1).toBeVisible()',
    'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(false\\)': 'await expect($1).toBeHidden()',
    'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toBe\\(([^)]+)\\)': 'await expect($1).toHaveText($2)',
    'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toContain\\(([^)]+)\\)': 'await expect($1).toContainText($2)',
    'expect\\(await ([^.]+)\\.getAttribute\\([\'"]value[\'"]\\)\\)\\.toBe\\(([^)]+)\\)': 'await expect($1).toHaveValue($2)',
    'expect\\(await ([^.]+)\\.isSelected\\(\\)\\)\\.toBe\\(true\\)': 'await expect($1).toBeChecked()',
    'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(false\\)': 'await expect($1).toBeDisabled()',
    'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(true\\)': 'await expect($1).toBeEnabled()'
  }
};
