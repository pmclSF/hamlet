/**
 * Selenium WebDriver (Python) framework definition.
 *
 * Provides detect, parse, and emit for Python Selenium WebDriver tests.
 * emit() converts JS Selenium, Java Selenium, Cypress, and Playwright
 * source code into Python Selenium + pytest code.
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  ImportStatement,
  RawCode,
  Comment,
  Modifier,
} from '../../../core/ir.js';
import { TodoFormatter } from '../../../core/TodoFormatter.js';

const formatter = new TodoFormatter('python');

function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong signals
  if (/from\s+selenium\s+import\s+webdriver/.test(source)) score += 40;
  if (/from\s+selenium\.webdriver\.common\.by\s+import\s+By/.test(source))
    score += 35;
  if (/webdriver\.Chrome\s*\(/.test(source)) score += 20;
  if (/driver\.find_element\s*\(\s*By\./.test(source)) score += 15;
  if (/By\.CSS_SELECTOR\b/.test(source)) score += 15;

  // Medium signals
  if (/\.send_keys\s*\(/.test(source)) score += 10;
  if (/WebDriverWait\s*\(/.test(source)) score += 10;
  if (/expected_conditions\b/.test(source)) score += 10;
  if (/driver\.get\s*\(/.test(source)) score += 5;
  if (/\.is_displayed\s*\(/.test(source)) score += 5;

  // Negative: JS Selenium
  if (/require\s*\(\s*['"]selenium-webdriver['"]/.test(source)) score -= 40;
  if (/from\s+['"]selenium-webdriver['"]/.test(source)) score -= 40;
  // Negative: Java Selenium
  if (/import\s+org\.openqa\.selenium/.test(source)) score -= 40;
  // Negative: Cypress/Playwright
  if (/\bcy\./.test(source)) score -= 30;
  if (/\bpage\.goto\s*\(/.test(source)) score -= 30;

  return Math.max(0, Math.min(100, score));
}

function parse(source) {
  const lines = source.split('\n');
  const imports = [];
  const allNodes = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    if (!trimmed) continue;

    if (trimmed.startsWith('#')) {
      allNodes.push(
        new Comment({
          text: line,
          sourceLocation: loc,
          originalSource: line,
        })
      );
      continue;
    }

    if (/^(?:import|from)\s/.test(trimmed)) {
      const sourceMatch = trimmed.match(
        /(?:from\s+(\S+)\s+import|import\s+(\S+))/
      );
      allNodes.push(
        new ImportStatement({
          kind: 'library',
          source: sourceMatch ? sourceMatch[1] || sourceMatch[2] : '',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      imports.push(allNodes[allNodes.length - 1]);
      continue;
    }

    if (/@pytest\.fixture\b/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: 'beforeEach',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/@pytest\.mark\./.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: 'tag',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/def\s+test_\w+\s*\(/.test(trimmed)) {
      allNodes.push(
        new TestCase({
          name: (trimmed.match(/def\s+(test_\w+)\s*\(/) || [])[1] || '',
          isAsync: /async\s+def/.test(trimmed),
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/^\s*assert\s+/.test(line)) {
      allNodes.push(
        new Assertion({
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/\bdriver\./.test(trimmed)) {
      allNodes.push(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    allNodes.push(
      new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
      })
    );
  }

  return new TestFile({
    language: 'python',
    imports,
    body: allNodes.filter((n) => !imports.includes(n)),
  });
}

/**
 * Emit Python Selenium + pytest code from IR + original source.
 */
function emit(_ir, source) {
  let result = source;

  const isJsSeleniumSource =
    /require\s*\(\s*['"]selenium-webdriver['"]/.test(source) ||
    /from\s+['"]selenium-webdriver['"]/.test(source);
  const isJavaSeleniumSource = /import\s+org\.openqa\.selenium/.test(source);
  const isCypressSource = /\bcy\./.test(source);
  const isPlaywrightSource =
    /from\s+['"]@playwright\/test['"]/.test(source) ||
    /\bpage\.goto\s*\(/.test(source);

  // Phase 1: Navigation
  if (isCypressSource) {
    result = result.replace(/cy\.visit\(([^)]+)\)/g, 'driver.get($1)');
    result = result.replace(/cy\.go\(['"]back['"]\)/g, 'driver.back()');
    result = result.replace(/cy\.go\(['"]forward['"]\)/g, 'driver.forward()');
    result = result.replace(/cy\.reload\(\)/g, 'driver.refresh()');
  }

  if (isPlaywrightSource) {
    result = result.replace(/await page\.goto\(([^)]+)\)/g, 'driver.get($1)');
    result = result.replace(/await page\.goBack\(\)/g, 'driver.back()');
    result = result.replace(/await page\.goForward\(\)/g, 'driver.forward()');
    result = result.replace(/await page\.reload\(\)/g, 'driver.refresh()');
  }

  if (isJsSeleniumSource) {
    // Navigation is similar, just remove await
    result = result.replace(/await\s+driver\.get\(/g, 'driver.get(');
    result = result.replace(
      /await\s+driver\.navigate\(\)\.back\(\)/g,
      'driver.back()'
    );
    result = result.replace(
      /await\s+driver\.navigate\(\)\.forward\(\)/g,
      'driver.forward()'
    );
    result = result.replace(
      /await\s+driver\.navigate\(\)\.refresh\(\)/g,
      'driver.refresh()'
    );
    // JS navigate() -> Python direct methods
    result = result.replace(/driver\.navigate\(\)\.back\(\)/g, 'driver.back()');
    result = result.replace(
      /driver\.navigate\(\)\.forward\(\)/g,
      'driver.forward()'
    );
    result = result.replace(
      /driver\.navigate\(\)\.refresh\(\)/g,
      'driver.refresh()'
    );
  }

  if (isJavaSeleniumSource) {
    result = result.replace(/driver\.navigate\(\)\.back\(\)/g, 'driver.back()');
    result = result.replace(
      /driver\.navigate\(\)\.forward\(\)/g,
      'driver.forward()'
    );
    result = result.replace(
      /driver\.navigate\(\)\.refresh\(\)/g,
      'driver.refresh()'
    );
  }

  // Phase 2: Selectors
  if (isCypressSource) {
    result = result.replace(
      /cy\.get\(([^)]+)\)/g,
      'driver.find_element(By.CSS_SELECTOR, $1)'
    );
  }

  if (isPlaywrightSource) {
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.fill\(([^)]+)\)/g,
      'driver.find_element(By.CSS_SELECTOR, $1).send_keys($2)'
    );
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.click\(\)/g,
      'driver.find_element(By.CSS_SELECTOR, $1).click()'
    );
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.textContent\(\)/g,
      'driver.find_element(By.CSS_SELECTOR, $1).text'
    );
    result = result.replace(
      /page\.locator\(([^)]+)\)/g,
      'driver.find_element(By.CSS_SELECTOR, $1)'
    );
  }

  if (isJsSeleniumSource) {
    // findElement -> find_element, By.css -> By.CSS_SELECTOR
    result = result.replace(/\.findElement\s*\(/g, '.find_element(');
    result = result.replace(/\.findElements\s*\(/g, '.find_elements(');
    result = result.replace(/By\.css\s*\(([^)]+)\)/g, 'By.CSS_SELECTOR, $1');
    result = result.replace(/By\.id\s*\(([^)]+)\)/g, 'By.ID, $1');
    result = result.replace(/By\.xpath\s*\(([^)]+)\)/g, 'By.XPATH, $1');
    result = result.replace(
      /By\.className\s*\(([^)]+)\)/g,
      'By.CLASS_NAME, $1'
    );
    result = result.replace(/By\.name\s*\(([^)]+)\)/g, 'By.NAME, $1');
  }

  if (isJavaSeleniumSource) {
    result = result.replace(/\.findElement\s*\(/g, '.find_element(');
    result = result.replace(/\.findElements\s*\(/g, '.find_elements(');
    result = result.replace(
      /By\.cssSelector\s*\(([^)]+)\)/g,
      'By.CSS_SELECTOR, $1'
    );
    result = result.replace(/By\.id\s*\(([^)]+)\)/g, 'By.ID, $1');
    result = result.replace(/By\.xpath\s*\(([^)]+)\)/g, 'By.XPATH, $1');
    result = result.replace(
      /By\.className\s*\(([^)]+)\)/g,
      'By.CLASS_NAME, $1'
    );
    result = result.replace(/By\.name\s*\(([^)]+)\)/g, 'By.NAME, $1');
  }

  // Phase 3: Interactions
  if (isCypressSource) {
    result = result.replace(/\.type\(([^)]+)\)/g, '.send_keys($1)');
    result = result.replace(/\.clear\(\)/g, '.clear()');
    result = result.replace(/\.check\(\)/g, '.click()');
  }

  if (isJsSeleniumSource || isJavaSeleniumSource) {
    result = result.replace(/\.sendKeys\s*\(/g, '.send_keys(');
    result = result.replace(/\.getText\s*\(\s*\)/g, '.text');
    result = result.replace(/\.isDisplayed\s*\(\s*\)/g, '.is_displayed()');
    result = result.replace(/\.getAttribute\s*\(/g, '.get_attribute(');
    result = result.replace(/\.getCurrentUrl\s*\(\s*\)/g, '.current_url');
    result = result.replace(/\.getTitle\s*\(\s*\)/g, '.title');
  }

  // Phase 4: Assertions
  if (isCypressSource) {
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.visible['"]\)/g,
      'assert driver.find_element(By.CSS_SELECTOR, $1).is_displayed()'
    );
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]have\.text['"],\s*([^)]+)\)/g,
      'assert driver.find_element(By.CSS_SELECTOR, $1).text == $2'
    );
    result = result.replace(
      /cy\.url\(\)\.should\(['"]include['"],\s*([^)]+)\)/g,
      'assert $1 in driver.current_url'
    );
  }

  if (isPlaywrightSource) {
    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toBeVisible\(\)/g,
      'assert driver.find_element(By.CSS_SELECTOR, $1).is_displayed()'
    );
    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toHaveText\(([^)]+)\)/g,
      'assert driver.find_element(By.CSS_SELECTOR, $1).text == $2'
    );
    result = result.replace(
      /await expect\(page\)\.toHaveURL\(([^)]+)\)/g,
      'assert driver.current_url == $1'
    );
    result = result.replace(
      /await expect\(page\)\.toHaveTitle\(([^)]+)\)/g,
      'assert driver.title == $1'
    );
  }

  if (isJsSeleniumSource) {
    result = result.replace(
      /expect\(([^)]+)\)\.toBe\(([^)]+)\)/g,
      'assert $1 == $2'
    );
    result = result.replace(
      /expect\(([^)]+)\)\.toContain\(([^)]+)\)/g,
      'assert $2 in $1'
    );
    result = result.replace(
      /expect\(([^)]+)\)\.toBeDefined\(\)/g,
      'assert $1 is not None'
    );
  }

  if (isJavaSeleniumSource) {
    result = result.replace(
      /Assertions\.assertEquals\s*\(([^,]+),\s*([^)]+)\)/g,
      'assert $2 == $1'
    );
    result = result.replace(
      /Assertions\.assertTrue\s*\(([^)]+)\)/g,
      'assert $1'
    );
    result = result.replace(
      /Assertions\.assertFalse\s*\(([^)]+)\)/g,
      'assert not $1'
    );
    result = result.replace(
      /Assertions\.assertNotNull\s*\(([^)]+)\)/g,
      'assert $1 is not None'
    );
  }

  // Phase 5: Convert structure to Python pytest
  if (isJsSeleniumSource || isCypressSource || isPlaywrightSource) {
    result = convertJsToPythonStructure(result);
  }

  if (isJavaSeleniumSource) {
    result = convertJavaToPythonStructure(result);
  }

  // Phase 6: Convert imports
  result = convertToPythonImports(
    result,
    isJsSeleniumSource,
    isJavaSeleniumSource,
    isCypressSource,
    isPlaywrightSource
  );

  // Phase 7: Cleanup
  result = result.replace(/\n{4,}/g, '\n\n\n');
  result = result.replace(/^\n+/, '');
  if (result.length > 0 && !result.endsWith('\n')) result += '\n';

  return result;
}

/**
 * Convert JS test structure to Python pytest functions.
 */
function convertJsToPythonStructure(source) {
  let result = source;

  // Remove describe wrapper
  result = result.replace(
    /describe\s*\([^,]+,\s*(?:async\s*)?\(\)\s*=>\s*\{\n?/g,
    ''
  );

  // Convert it() to def test_*():
  result = result.replace(
    /\s*it\s*\(\s*(['"])([^'"]+)\1\s*,\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
    (_, _q, name) => {
      const funcName =
        'test_' +
        name
          .toLowerCase()
          .replace(/[^a-z0-9\s]/g, '')
          .trim()
          .replace(/\s+/g, '_');
      return `\ndef ${funcName}(driver):`;
    }
  );

  // Convert beforeAll to fixture
  result = result.replace(
    /\s*beforeAll\s*\(\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
    '\n@pytest.fixture(scope="module", autouse=True)\ndef setup_driver():'
  );

  // Convert afterAll to fixture teardown
  result = result.replace(
    /\s*afterAll\s*\(\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
    '\n# teardown handled in fixture'
  );

  // Remove closing }); from JS blocks
  result = result.replace(/^\s*\}\s*\)\s*;?\s*$/gm, '');

  // Remove remaining lone closing braces at start of line
  result = result.replace(/^\s*\}\s*$/gm, '');

  // Remove await keywords
  result = result.replace(/\bawait\s+/g, '');

  // Remove semicolons
  result = result.replace(/;(\s*)$/gm, '$1');

  // Convert const/let/var to bare assignment
  result = result.replace(/\b(?:const|let|var)\s+/g, '');

  // Fix indentation: remove one level of JS indentation (2 spaces)
  const lines = result.split('\n');
  const processed = lines.map((line) => {
    if (/^    /.test(line)) {
      return line.replace(/^    /, '    ');
    }
    if (/^  /.test(line) && !/^def /.test(line.trim())) {
      return line.replace(/^  /, '');
    }
    return line;
  });
  result = processed.join('\n');

  return result;
}

/**
 * Convert Java test structure to Python pytest functions.
 */
function convertJavaToPythonStructure(source) {
  let result = source;

  // Remove class wrapper
  result = result.replace(/public\s+class\s+\w+\s*\{?\n?/g, '');

  // Convert @Test + void method to def test_*
  result = result.replace(
    /@Test\s*\n\s*(?:public\s+)?void\s+(\w+)\s*\(\s*\)\s*\{/g,
    (_, name) => {
      const pyName = name
        .replace(/([A-Z])/g, '_$1')
        .toLowerCase()
        .replace(/^_/, '');
      return `def ${pyName.startsWith('test') ? pyName : 'test_' + pyName}(driver):`;
    }
  );

  // Convert @BeforeAll
  result = result.replace(
    /@BeforeAll\s*\n\s*(?:public\s+)?static\s+void\s+\w+\s*\(\s*\)\s*\{/g,
    '@pytest.fixture(scope="module", autouse=True)\ndef setup_driver():'
  );

  // Convert @AfterAll
  result = result.replace(
    /@AfterAll\s*\n\s*(?:public\s+)?static\s+void\s+\w+\s*\(\s*\)\s*\{/g,
    '# teardown handled in fixture'
  );

  // Remove closing braces
  result = result.replace(/^\s*\}\s*$/gm, '');

  // Remove semicolons
  result = result.replace(/;(\s*)$/gm, '$1');

  // Remove Java type declarations
  result = result.replace(
    /\b(?:WebDriver|WebElement|String|int|boolean|void|static)\s+/g,
    ''
  );

  return result;
}

/**
 * Convert imports to Python Selenium + pytest.
 */
function convertToPythonImports(
  source,
  isJsSel,
  isJavaSel,
  isCypress,
  isPlaywright
) {
  let result = source;

  // Remove JS imports
  result = result.replace(
    /(?:const|let|var)\s+\{[^}]*\}\s*=\s*require\s*\([^)]+\)\s*;?\n?/g,
    ''
  );
  result = result.replace(
    /import\s+\{[^}]*\}\s+from\s+['"][^'"]+['"];?\n?/g,
    ''
  );
  result = result.replace(/import\s+\w+\s+from\s+['"][^'"]+['"];?\n?/g, '');
  result = result.replace(
    /\/\/\/\s*<reference\s+types=["'][^'"]+["']\s*\/>\n?/g,
    ''
  );

  // Remove Java imports
  result = result.replace(/^import\s+org\.\w+[^;\n]*;?\n?/gm, '');
  result = result.replace(/^import\s+static\s+[^;\n]*;?\n?/gm, '');

  // Add Python imports
  const pythonImports = [
    'import pytest',
    'from selenium import webdriver',
    'from selenium.webdriver.common.by import By',
  ].join('\n');

  // Add at top
  result = pythonImports + '\n\n\n' + result.replace(/^\n+/, '');

  return result;
}

export default {
  name: 'selenium-python',
  language: 'python',
  paradigm: 'function-e2e',
  detect,
  parse,
  emit,
  imports: {
    packages: ['selenium', 'pytest'],
  },
};
