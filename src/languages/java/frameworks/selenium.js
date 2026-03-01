/**
 * Selenium WebDriver (Java) framework definition.
 *
 * Provides detect, parse, and emit for Java Selenium WebDriver tests.
 * emit() converts JS Selenium, Cypress, and Playwright source code
 * into Java Selenium + JUnit 5 code.
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

const formatter = new TodoFormatter('java');

function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong signals
  if (/import\s+org\.openqa\.selenium/.test(source)) score += 40;
  if (/WebDriver\s+\w+\s*=/.test(source)) score += 20;
  if (/new\s+ChromeDriver\s*\(/.test(source)) score += 15;
  if (/driver\.findElement\s*\(\s*By\.cssSelector/.test(source)) score += 15;
  if (/By\.cssSelector\s*\(/.test(source)) score += 15;

  // Medium signals
  if (/WebDriverWait\b/.test(source)) score += 10;
  if (/ExpectedConditions\b/.test(source)) score += 10;
  if (/\.sendKeys\s*\(/.test(source)) score += 10;
  if (/\.getText\s*\(\s*\)/.test(source)) score += 10;
  if (/import\s+org\.openqa\.selenium\.chrome/.test(source)) score += 10;

  // Negative: JS Selenium
  if (/require\s*\(\s*['"]selenium-webdriver['"]/.test(source)) score -= 40;
  if (/from\s+['"]selenium-webdriver['"]/.test(source)) score -= 40;
  // Negative: Python Selenium
  if (/from\s+selenium\s+import/.test(source)) score -= 40;
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

    if (
      trimmed.startsWith('//') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*')
    ) {
      allNodes.push(
        new Comment({
          text: line,
          sourceLocation: loc,
          originalSource: line,
        })
      );
      continue;
    }

    if (/^import\s/.test(trimmed)) {
      const sourceMatch = trimmed.match(/import\s+(?:static\s+)?([^\s;]+)/);
      allNodes.push(
        new ImportStatement({
          kind: 'library',
          source: sourceMatch ? sourceMatch[1] : '',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      imports.push(allNodes[allNodes.length - 1]);
      continue;
    }

    if (/\bclass\s+\w+/.test(trimmed)) {
      allNodes.push(
        new TestSuite({
          name: (trimmed.match(/class\s+(\w+)/) || [])[1] || '',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/@Test\b/.test(trimmed)) {
      allNodes.push(
        new Modifier({
          modifierType: 'test',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/@BeforeAll\b/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: 'beforeAll',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/@AfterAll\b/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: 'afterAll',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/@BeforeEach\b/.test(trimmed)) {
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
    if (/@AfterEach\b/.test(trimmed)) {
      allNodes.push(
        new Hook({
          hookType: 'afterEach',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/(?:public\s+|protected\s+|private\s+)?void\s+\w+\s*\(/.test(trimmed)) {
      allNodes.push(
        new TestCase({
          name: (trimmed.match(/void\s+(\w+)\s*\(/) || [])[1] || '',
          isAsync: false,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/\bAssertions\.\w+\s*\(/.test(trimmed)) {
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
    language: 'java',
    imports,
    body: allNodes.filter((n) => !imports.includes(n)),
  });
}

/**
 * Emit Java Selenium + JUnit 5 code from IR + original source.
 *
 * Handles JS Selenium → Java Selenium, Cypress → Java Selenium,
 * and Playwright → Java Selenium conversions.
 */
function emit(_ir, source) {
  let result = source;

  const isJsSeleniumSource =
    /require\s*\(\s*['"]selenium-webdriver['"]/.test(source) ||
    /from\s+['"]selenium-webdriver['"]/.test(source);
  const isCypressSource = /\bcy\./.test(source);
  const isPlaywrightSource =
    /from\s+['"]@playwright\/test['"]/.test(source) ||
    /\bpage\.goto\s*\(/.test(source);

  // Phase 1: Convert navigation
  if (isCypressSource) {
    result = result.replace(/cy\.visit\(([^)]+)\)/g, 'driver.get($1)');
    result = result.replace(
      /cy\.go\(['"]back['"]\)/g,
      'driver.navigate().back()'
    );
    result = result.replace(
      /cy\.go\(['"]forward['"]\)/g,
      'driver.navigate().forward()'
    );
    result = result.replace(/cy\.reload\(\)/g, 'driver.navigate().refresh()');
  }

  if (isPlaywrightSource) {
    result = result.replace(/await page\.goto\(([^)]+)\)/g, 'driver.get($1)');
    result = result.replace(
      /await page\.goBack\(\)/g,
      'driver.navigate().back()'
    );
    result = result.replace(
      /await page\.goForward\(\)/g,
      'driver.navigate().forward()'
    );
    result = result.replace(
      /await page\.reload\(\)/g,
      'driver.navigate().refresh()'
    );
  }

  if (isJsSeleniumSource) {
    result = result.replace(/await driver\.get\(/g, 'driver.get(');
    result = result.replace(/await driver\.navigate\(\)/g, 'driver.navigate()');
  }

  // Phase 2: Convert selectors
  if (isCypressSource) {
    result = result.replace(
      /cy\.get\(([^)]+)\)/g,
      'driver.findElement(By.cssSelector($1))'
    );
    result = result.replace(
      /cy\.contains\(([^)]+)\)/g,
      'driver.findElement(By.xpath("//*[contains(text()," + $1 + ")]"))'
    );
  }

  if (isPlaywrightSource) {
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.fill\(([^)]+)\)/g,
      'driver.findElement(By.cssSelector($1)).sendKeys($2)'
    );
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.click\(\)/g,
      'driver.findElement(By.cssSelector($1)).click()'
    );
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.textContent\(\)/g,
      'driver.findElement(By.cssSelector($1)).getText()'
    );
    result = result.replace(
      /page\.locator\(([^)]+)\)/g,
      'driver.findElement(By.cssSelector($1))'
    );
  }

  if (isJsSeleniumSource) {
    // By.css -> By.cssSelector
    result = result.replace(/By\.css\(/g, 'By.cssSelector(');
    // Remove awaits (Java Selenium is synchronous)
    result = result.replace(/await\s+/g, '');
  }

  // Phase 3: Convert interactions
  if (isCypressSource) {
    result = result.replace(/\.type\(([^)]+)\)/g, '.sendKeys($1)');
    result = result.replace(/\.clear\(\)/g, '.clear()');
    result = result.replace(/\.check\(\)/g, '.click()');
    result = result.replace(/\.uncheck\(\)/g, '.click()');
  }

  // Phase 4: Convert assertions
  if (isCypressSource) {
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.visible['"]\)/g,
      'Assertions.assertTrue(driver.findElement(By.cssSelector($1)).isDisplayed())'
    );
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]have\.text['"],\s*([^)]+)\)/g,
      'Assertions.assertEquals($2, driver.findElement(By.cssSelector($1)).getText())'
    );
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]contain['"],\s*([^)]+)\)/g,
      'Assertions.assertTrue(driver.findElement(By.cssSelector($1)).getText().contains($2))'
    );
    result = result.replace(
      /cy\.url\(\)\.should\(['"]include['"],\s*([^)]+)\)/g,
      'Assertions.assertTrue(driver.getCurrentUrl().contains($1))'
    );
  }

  if (isPlaywrightSource) {
    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toBeVisible\(\)/g,
      'Assertions.assertTrue(driver.findElement(By.cssSelector($1)).isDisplayed())'
    );
    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toHaveText\(([^)]+)\)/g,
      'Assertions.assertEquals($2, driver.findElement(By.cssSelector($1)).getText())'
    );
    result = result.replace(
      /await expect\(page\)\.toHaveURL\(([^)]+)\)/g,
      'Assertions.assertEquals($1, driver.getCurrentUrl())'
    );
    result = result.replace(
      /await expect\(page\)\.toHaveTitle\(([^)]+)\)/g,
      'Assertions.assertEquals($1, driver.getTitle())'
    );
  }

  if (isJsSeleniumSource) {
    result = result.replace(
      /expect\(([^)]+)\)\.toBe\(([^)]+)\)/g,
      'Assertions.assertEquals($2, $1)'
    );
    result = result.replace(
      /expect\(([^)]+)\)\.toContain\(([^)]+)\)/g,
      'Assertions.assertTrue($1.contains($2))'
    );
    result = result.replace(
      /expect\(([^)]+)\)\.toBeDefined\(\)/g,
      'Assertions.assertNotNull($1)'
    );
  }

  // Phase 5: Convert test structure to Java class + JUnit 5
  if (isJsSeleniumSource || isCypressSource || isPlaywrightSource) {
    result = convertToJavaStructure(result);
  }

  // Phase 6: Convert imports
  result = convertToJavaImports(
    result,
    isJsSeleniumSource,
    isCypressSource,
    isPlaywrightSource
  );

  // Phase 7: Cleanup
  result = result.replace(/\n{3,}/g, '\n\n');
  if (!result.endsWith('\n')) result += '\n';

  return result;
}

/**
 * Convert JS test structure to Java class + JUnit 5 structure.
 */
function convertToJavaStructure(source) {
  let result = source;

  // Extract test name from describe block
  const describeMatch = result.match(/describe\s*\(\s*(['"])([^'"]+)\1/);
  const className = describeMatch
    ? describeMatch[2].replace(/[^a-zA-Z0-9]/g, '') + 'Test'
    : 'ConvertedTest';

  // Convert describe/it structure
  // Remove describe wrapper
  result = result.replace(
    /describe\s*\([^,]+,\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
    `public class ${className} {`
  );

  // Convert beforeAll
  result = result.replace(
    /beforeAll\s*\(\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
    '@BeforeAll\n    static void setUp() {'
  );

  // Convert afterAll
  result = result.replace(
    /afterAll\s*\(\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
    '@AfterAll\n    static void tearDown() {'
  );

  // Convert beforeEach
  result = result.replace(
    /beforeEach\s*\(\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
    '@BeforeEach\n    void beforeEach() {'
  );

  // Convert afterEach
  result = result.replace(
    /afterEach\s*\(\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
    '@AfterEach\n    void afterEach() {'
  );

  // Convert it() to @Test void
  result = result.replace(
    /it\s*\(\s*(['"])([^'"]+)\1\s*,\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
    (_, _q, name) => {
      const methodName =
        'test' +
        name
          .replace(/should\s+/i, '')
          .replace(/[^a-zA-Z0-9\s]/g, '')
          .split(/\s+/)
          .map((w, i) =>
            i === 0
              ? w.toLowerCase()
              : w.charAt(0).toUpperCase() + w.slice(1).toLowerCase()
          )
          .join('');
      return `@Test\n    void ${methodName}() {`;
    }
  );

  // Remove arrow function closing }) -> just }
  result = result.replace(/\}\s*\)\s*;/g, '}');

  // Convert const/let/var to Java types
  result = result.replace(
    /(?:const|let|var)\s+driver\s*=/g,
    'static WebDriver driver ='
  );
  result = result.replace(
    /(?:const|let|var)\s+(\w+)\s*=\s*(?:await\s+)?driver\.findElement/g,
    'WebElement $1 = driver.findElement'
  );
  result = result.replace(
    /(?:const|let|var)\s+(\w+)\s*=\s*(?:await\s+)?new\s+Builder/g,
    'static WebDriver $1 = new Builder'
  );

  // Remove remaining await keywords
  result = result.replace(/\bawait\s+/g, '');

  // Remove semicolons at end of blocks (JS artifact)
  // Add semicolons to statements that need them
  const lines = result.split('\n');
  const processed = lines.map((line) => {
    const trimmed = line.trim();
    // Skip empty lines, comments, annotations, braces
    if (
      !trimmed ||
      trimmed.startsWith('//') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*') ||
      trimmed.startsWith('@') ||
      trimmed === '{' ||
      trimmed === '}' ||
      trimmed.endsWith('{') ||
      trimmed.endsWith('}')
    ) {
      return line;
    }
    // Add semicolons to statement lines that don't have them
    if (
      !trimmed.endsWith(';') &&
      !trimmed.endsWith('{') &&
      !trimmed.endsWith('}') &&
      !trimmed.startsWith('public class') &&
      !trimmed.startsWith('import ')
    ) {
      return line + ';';
    }
    return line;
  });
  result = processed.join('\n');

  return result;
}

/**
 * Convert imports to Java Selenium + JUnit 5.
 */
function convertToJavaImports(source, isJsSel, isCypress, isPlaywright) {
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

  // Add Java imports at the top
  const javaImports = [
    'import org.junit.jupiter.api.*;',
    'import static org.junit.jupiter.api.Assertions.*;',
    'import org.openqa.selenium.*;',
    'import org.openqa.selenium.chrome.ChromeDriver;',
  ].join('\n');

  // Insert imports before the first class declaration or at top
  const classMatch = result.match(/^(public\s+class\s)/m);
  if (classMatch) {
    result = result.replace(
      classMatch[0],
      javaImports + '\n\n' + classMatch[0]
    );
  } else {
    result = javaImports + '\n\n' + result;
  }

  // Deduplicate: remove duplicate import blocks
  result = result.replace(
    /(import\s+org\.openqa\.selenium\.\*;\n?){2,}/g,
    'import org.openqa.selenium.*;\n'
  );

  return result;
}

export default {
  name: 'selenium-java',
  language: 'java',
  paradigm: 'xunit-e2e',
  detect,
  parse,
  emit,
  imports: {
    packages: [
      'org.openqa.selenium',
      'org.openqa.selenium.chrome',
      'org.junit.jupiter.api',
    ],
  },
};
