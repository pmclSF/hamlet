/**
 * Interaction patterns for all three frameworks
 * Handles clicks, typing, form interactions, and user actions
 */

export const interactionPatterns = {
  cypress: {
    patterns: {
      // Clicks
      '\\.click\\(\\)': '.CLICK()',
      '\\.click\\(\\{\\s*force:\\s*true\\s*\\}\\)': '.CLICK_FORCE()',
      '\\.dblclick\\(\\)': '.DOUBLE_CLICK()',
      '\\.rightclick\\(\\)': '.RIGHT_CLICK()',

      // Input
      '\\.type\\(([^)]+)\\)': '.TYPE($1)',
      '\\.type\\(([^,]+),\\s*\\{[^}]*delay:\\s*(\\d+)[^}]*\\}\\)':
        '.TYPE_SLOW($1, $2)',
      '\\.clear\\(\\)': '.CLEAR()',
      '\\.focus\\(\\)': '.FOCUS()',
      '\\.blur\\(\\)': '.BLUR()',

      // Form elements
      '\\.check\\(\\)': '.CHECK()',
      '\\.uncheck\\(\\)': '.UNCHECK()',
      '\\.select\\(([^)]+)\\)': '.SELECT_OPTION($1)',

      // Mouse actions
      '\\.trigger\\([\'"]mouseover[\'"]\\)': '.HOVER()',
      '\\.trigger\\([\'"]mouseenter[\'"]\\)': '.HOVER()',
      '\\.trigger\\([\'"]mouseleave[\'"]\\)': '.MOUSE_LEAVE()',
      '\\.trigger\\([\'"]mousedown[\'"]\\)': '.MOUSE_DOWN()',
      '\\.trigger\\([\'"]mouseup[\'"]\\)': '.MOUSE_UP()',

      // Scrolling
      '\\.scrollTo\\(([^)]+)\\)': '.SCROLL_TO($1)',
      '\\.scrollIntoView\\(\\)': '.SCROLL_INTO_VIEW()',

      // Drag and drop
      '\\.drag\\(([^)]+)\\)': '.DRAG($1)',
      '\\.trigger\\([\'"]dragstart[\'"]\\)': '.DRAG_START()',
      '\\.trigger\\([\'"]drop[\'"]\\)': '.DROP()',

      // File upload
      '\\.selectFile\\(([^)]+)\\)': '.UPLOAD_FILE($1)',
      '\\.attachFile\\(([^)]+)\\)': '.UPLOAD_FILE($1)',
    },

    generators: {
      CLICK: () => '.click()',
      CLICK_FORCE: () => '.click({ force: true })',
      DOUBLE_CLICK: () => '.dblclick()',
      RIGHT_CLICK: () => '.rightclick()',
      TYPE: (text) => `.type(${text})`,
      TYPE_SLOW: (text, delay) => `.type(${text}, { delay: ${delay} })`,
      CLEAR: () => '.clear()',
      FOCUS: () => '.focus()',
      BLUR: () => '.blur()',
      CHECK: () => '.check()',
      UNCHECK: () => '.uncheck()',
      SELECT_OPTION: (value) => `.select(${value})`,
      HOVER: () => '.trigger("mouseover")',
      MOUSE_LEAVE: () => '.trigger("mouseleave")',
      MOUSE_DOWN: () => '.trigger("mousedown")',
      MOUSE_UP: () => '.trigger("mouseup")',
      SCROLL_TO: (pos) => `.scrollTo(${pos})`,
      SCROLL_INTO_VIEW: () => '.scrollIntoView()',
      DRAG: (target) => `.drag(${target})`,
      DRAG_START: () => '.trigger("dragstart")',
      DROP: () => '.trigger("drop")',
      UPLOAD_FILE: (file) => `.selectFile(${file})`,
    },
  },

  playwright: {
    patterns: {
      // Clicks
      '\\.click\\(\\)': '.CLICK()',
      '\\.click\\(\\{\\s*force:\\s*true\\s*\\}\\)': '.CLICK_FORCE()',
      '\\.dblclick\\(\\)': '.DOUBLE_CLICK()',
      '\\.click\\(\\{\\s*button:\\s*[\'"]right[\'"]\\s*\\}\\)':
        '.RIGHT_CLICK()',

      // Input
      '\\.fill\\(([^)]+)\\)': '.TYPE($1)',
      '\\.type\\(([^)]+)\\)': '.TYPE_CHAR($1)',
      '\\.clear\\(\\)': '.CLEAR()',
      '\\.focus\\(\\)': '.FOCUS()',
      '\\.blur\\(\\)': '.BLUR()',

      // Form elements
      '\\.check\\(\\)': '.CHECK()',
      '\\.uncheck\\(\\)': '.UNCHECK()',
      '\\.selectOption\\(([^)]+)\\)': '.SELECT_OPTION($1)',

      // Mouse actions
      '\\.hover\\(\\)': '.HOVER()',

      // Scrolling
      '\\.scrollIntoViewIfNeeded\\(\\)': '.SCROLL_INTO_VIEW()',

      // Drag and drop
      '\\.dragTo\\(([^)]+)\\)': '.DRAG($1)',

      // File upload
      '\\.setInputFiles\\(([^)]+)\\)': '.UPLOAD_FILE($1)',
    },

    generators: {
      CLICK: () => '.click()',
      CLICK_FORCE: () => '.click({ force: true })',
      DOUBLE_CLICK: () => '.dblclick()',
      RIGHT_CLICK: () => '.click({ button: "right" })',
      TYPE: (text) => `.fill(${text})`,
      TYPE_CHAR: (text) => `.type(${text})`,
      TYPE_SLOW: (text, delay) => `.type(${text}, { delay: ${delay} })`,
      CLEAR: () => '.clear()',
      FOCUS: () => '.focus()',
      BLUR: () => '.blur()',
      CHECK: () => '.check()',
      UNCHECK: () => '.uncheck()',
      SELECT_OPTION: (value) => `.selectOption(${value})`,
      HOVER: () => '.hover()',
      MOUSE_LEAVE: () => '.dispatchEvent("mouseleave")',
      MOUSE_DOWN: () => '.dispatchEvent("mousedown")',
      MOUSE_UP: () => '.dispatchEvent("mouseup")',
      SCROLL_TO: (pos) => `.evaluate((el, p) => el.scrollTo(p), ${pos})`,
      SCROLL_INTO_VIEW: () => '.scrollIntoViewIfNeeded()',
      DRAG: (target) => `.dragTo(${target})`,
      DRAG_START: () => '.dispatchEvent("dragstart")',
      DROP: () => '.dispatchEvent("drop")',
      UPLOAD_FILE: (file) => `.setInputFiles(${file})`,
    },
  },

  selenium: {
    patterns: {
      // Clicks
      '\\.click\\(\\)': '.CLICK()',

      // Input
      '\\.sendKeys\\(([^)]+)\\)': '.TYPE($1)',
      '\\.clear\\(\\)': '.CLEAR()',

      // Form (Select class)
      'new Select\\(([^)]+)\\)\\.selectByVisibleText\\(([^)]+)\\)':
        'SELECT_OPTION_IN($1, $2)',
      'new Select\\(([^)]+)\\)\\.selectByValue\\(([^)]+)\\)':
        'SELECT_BY_VALUE_IN($1, $2)',
    },

    generators: {
      CLICK: () => '.click()',
      CLICK_FORCE: () => '.click()', // Selenium doesn't have force option
      DOUBLE_CLICK: (context) =>
        `await driver.actions().doubleClick(${context || 'element'}).perform()`,
      RIGHT_CLICK: (context) =>
        `await driver.actions().contextClick(${context || 'element'}).perform()`,
      TYPE: (text) => `.sendKeys(${text})`,
      TYPE_CHAR: (text) => `.sendKeys(${text})`,
      TYPE_SLOW: (text) => `.sendKeys(${text})`, // Selenium doesn't have built-in delay
      CLEAR: () => '.clear()',
      FOCUS: () => '.click()', // Selenium focuses by clicking
      BLUR: () =>
        '; await driver.executeScript("arguments[0].blur()", element)',
      CHECK: () => '.click()', // Selenium uses click for checkbox
      UNCHECK: () => '.click()',
      SELECT_OPTION: (value) =>
        `; await new Select(element).selectByVisibleText(${value})`,
      HOVER: (context) =>
        `await driver.actions().move({ origin: ${context || 'element'} }).perform()`,
      MOUSE_LEAVE: () =>
        'await driver.actions().move({ x: 0, y: 0 }).perform()',
      MOUSE_DOWN: (context) =>
        `await driver.actions().press(${context || 'element'}).perform()`,
      MOUSE_UP: () => 'await driver.actions().release().perform()',
      SCROLL_TO: (pos) =>
        `await driver.executeScript("window.scrollTo(${pos})")`,
      SCROLL_INTO_VIEW: () =>
        'await driver.executeScript("arguments[0].scrollIntoView(true)", element)',
      DRAG: (target) =>
        `await driver.actions().dragAndDrop(element, ${target}).perform()`,
      DRAG_START: () =>
        'await driver.executeScript("arguments[0].dispatchEvent(new DragEvent(\'dragstart\'))", element)',
      DROP: () =>
        'await driver.executeScript("arguments[0].dispatchEvent(new DragEvent(\'drop\'))", element)',
      UPLOAD_FILE: (file) => `.sendKeys(${file})`,
    },
  },
};

/**
 * Direct conversion mappings between frameworks
 */
export const directMappings = {
  'cypress-playwright': {
    '\\.click\\(\\)': '.click()',
    '\\.dblclick\\(\\)': '.dblclick()',
    '\\.rightclick\\(\\)': '.click({ button: "right" })',
    '\\.type\\(([^)]+)\\)': '.fill($1)',
    '\\.clear\\(\\)': '.clear()',
    '\\.focus\\(\\)': '.focus()',
    '\\.blur\\(\\)': '.blur()',
    '\\.check\\(\\)': '.check()',
    '\\.uncheck\\(\\)': '.uncheck()',
    '\\.select\\(([^)]+)\\)': '.selectOption($1)',
    '\\.trigger\\([\'"]mouseover[\'"]\\)': '.hover()',
    '\\.trigger\\([\'"]mouseenter[\'"]\\)': '.hover()',
    '\\.scrollIntoView\\(\\)': '.scrollIntoViewIfNeeded()',
    '\\.selectFile\\(([^)]+)\\)': '.setInputFiles($1)',
    '\\.attachFile\\(([^)]+)\\)': '.setInputFiles($1)',
  },

  'cypress-selenium': {
    '\\.click\\(\\)': '.click()',
    '\\.type\\(([^)]+)\\)': '.sendKeys($1)',
    '\\.clear\\(\\)': '.clear()',
    '\\.check\\(\\)': '.click()',
    '\\.uncheck\\(\\)': '.click()',
    '\\.scrollIntoView\\(\\)':
      '; await driver.executeScript("arguments[0].scrollIntoView(true)", element)',
  },

  'playwright-cypress': {
    '\\.click\\(\\)': '.click()',
    '\\.dblclick\\(\\)': '.dblclick()',
    '\\.click\\(\\{\\s*button:\\s*[\'"]right[\'"]\\s*\\}\\)': '.rightclick()',
    '\\.fill\\(([^)]+)\\)': '.type($1)',
    '\\.clear\\(\\)': '.clear()',
    '\\.focus\\(\\)': '.focus()',
    '\\.blur\\(\\)': '.blur()',
    '\\.check\\(\\)': '.check()',
    '\\.uncheck\\(\\)': '.uncheck()',
    '\\.selectOption\\(([^)]+)\\)': '.select($1)',
    '\\.hover\\(\\)': '.trigger("mouseover")',
    '\\.scrollIntoViewIfNeeded\\(\\)': '.scrollIntoView()',
    '\\.setInputFiles\\(([^)]+)\\)': '.selectFile($1)',
  },

  'playwright-selenium': {
    '\\.click\\(\\)': '.click()',
    '\\.fill\\(([^)]+)\\)': '.sendKeys($1)',
    '\\.clear\\(\\)': '.clear()',
    '\\.check\\(\\)': '.click()',
    '\\.uncheck\\(\\)': '.click()',
    '\\.setInputFiles\\(([^)]+)\\)': '.sendKeys($1)',
  },

  'selenium-cypress': {
    '\\.click\\(\\)': '.click()',
    '\\.sendKeys\\(([^)]+)\\)': '.type($1)',
    '\\.clear\\(\\)': '.clear()',
  },

  'selenium-playwright': {
    '\\.click\\(\\)': '.click()',
    '\\.sendKeys\\(([^)]+)\\)': '.fill($1)',
    '\\.clear\\(\\)': '.clear()',
  },
};
