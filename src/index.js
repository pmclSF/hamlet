import fs from 'fs/promises';
import path from 'path';
import chalk from 'chalk';

/**
 * Detect type of Cypress test
 * @param {string} content - Test content
 * @returns {string[]} - Array of detected test types
 */
function detectTestType(content) {
    const patterns = {
      api: /cy\.request|cy\.intercept|\.then\s*\(\s*{\s*status/i,
      component: /cy\.mount|mount\(/i,
      accessibility: /cy\.injectAxe|cy\.checkA11y|aria-|role=/i,
      visual: /cy\.screenshot|matchImageSnapshot/i,
      performance: /cy\.lighthouse|performance\.|timing/i,
      mobile: /viewport|mobile|touch|swipe/i,
    };
  
    return Object.entries(patterns)
      .filter(([_, pattern]) => pattern.test(content))
      .map(([type]) => type);
  }
  
  /**
   * Generate required imports based on test type
   * @param {string[]} types - Array of test types
   * @returns {string[]} - Array of import statements
   */
  function generateImports(types) {
    const imports = new Set([`import { test, expect } from '@playwright/test';`]);
  
    const typeImports = {
      api: `import { request } from '@playwright/test';`,
      component: `import { mount } from '@playwright/experimental-ct-react';`,
      accessibility: `import { injectAxe, checkA11y } from 'axe-playwright';`,
      visual: `import { expect } from '@playwright/test';`,
    };
  
    types.forEach(type => {
      if (typeImports[type]) {
        imports.add(typeImports[type]);
      }
    });
  
    return Array.from(imports);
  }

/**
 * Convert Cypress test to Playwright format
 * @param {string} cypressContent - Content of Cypress test file
 * @returns {string} - Converted Playwright test content
 */
export async function convertCypressToPlaywright(cypressContent) {
  let playwrightContent = cypressContent;

  // Detect test type
  const testType = detectTestType(cypressContent);
  
  // Get required imports based on test type
  const imports = generateImports(testType);

  // Basic conversion patterns
  const conversions = {
    // Test Structure
    'describe\\(': 'test.describe(',
    'it\\(': 'test(',
    'cy\\.': 'await page.',
    'before\\(': 'test.beforeAll(',
    'after\\(': 'test.afterAll(',
    'beforeEach\\(': 'test.beforeEach(',
    'afterEach\\(': 'test.afterEach(',
    
    // Basic Commands
    'visit\\(': 'goto(',
    'get\\(': 'locator(',
    'find\\(': 'locator(',
    'type\\(': 'fill(',
    'click\\(': 'click(',
    'dblclick\\(': 'dblclick(',
    'rightclick\\(': 'click({ button: "right" })',
    'check\\(': 'check(',
    'uncheck\\(': 'uncheck(',
    'select\\(': 'selectOption(',
    'scrollTo\\(': 'scroll(',
    'scrollIntoView\\(': 'scrollIntoViewIfNeeded(',
    'trigger\\(': 'dispatchEvent(',
    'focus\\(': 'focus(',
    'blur\\(': 'blur(',
    'clear\\(': 'clear(',
    
    // Assertions
    'should\\(\'be.visible\'\\)': 'toBeVisible()',
    'should\\(\'not.be.visible\'\\)': 'toBeHidden()',
    'should\\(\'exist\'\\)': 'toBeVisible()',
    'should\\(\'not.exist\'\\)': 'toBeHidden()',
    'should\\(\'have.text\',\\s*([^)]+)\\)': 'toHaveText($1)',
    'should\\(\'have.value\',\\s*([^)]+)\\)': 'toHaveValue($1)',
    'should\\(\'be.checked\'\\)': 'toBeChecked()',
    'should\\(\'be.disabled\'\\)': 'toBeDisabled()',
    'should\\(\'be.enabled\'\\)': 'toBeEnabled()',
    'should\\(\'have.class\',\\s*([^)]+)\\)': 'toHaveClass($1)',
    'should\\(\'have.attr\',\\s*([^)]+)\\)': 'toHaveAttribute($1)',
    'should\\(\'have.length\'\\)': 'toHaveCount(',
    'should\\(\'be.empty\'\\)': 'toBeEmpty()',
    'should\\(\'be.focused\'\\)': 'toBeFocused()',
    
    // API Testing
    'request\\(': 'await request.fetch(',
    'intercept\\(': 'await page.route(',
    'wait\\(@([^)]+)\\)': 'waitForResponse(response => response.url().includes($1))',
    
    // Component Testing
    'mount\\(': 'await mount(',
    '\\.shadow\\(\\)': '.shadowRoot()',
    
    // Accessibility Testing
    'injectAxe\\(': 'await injectAxe(page)',
    'checkA11y\\(': 'await checkA11y(page)',
    
    // Visual Testing
    'matchImageSnapshot\\(': 'screenshot({ name: ',
    
    // File Handling
    'readFile\\(': 'await fs.readFile(',
    'writeFile\\(': 'await fs.writeFile(',
    'fixture\\(': 'await fs.readFile(path.join(\'fixtures\', ',
    
    // Iframe Handling
    'iframe\\(\\)': 'frameLocator()',
    
    // Multiple Windows/Tabs
    'window\\(\\)': 'context.newPage()',
    
    // Local Storage
    'clearLocalStorage\\(': 'evaluate(() => localStorage.clear())',
    'clearCookies\\(': 'context.clearCookies()',
    
    // Mouse Events
    'hover\\(': 'hover(',
    'mousedown\\(': 'mouseDown(',
    'mouseup\\(': 'mouseUp(',
    'mousemove\\(': 'moveBy(',
    
    // Keyboard Events
    'type\\(': 'type(',
    'press\\(': 'press(',
    
    // Viewport/Responsive
    'viewport\\(': 'setViewportSize(',
    
    // Network
    'server\\(': '// Use page.route() instead of cy.server()',
    
    // State Management
    'window\\.store': 'await page.evaluate(() => window.store',
    
    // Database
    'task\\(': 'await request.fetch(\'/api/db\', ',
    
    // Custom Commands (placeholder - will be handled separately)
    'Cypress\\.Commands\\.add\\(': '// Convert to Playwright helper function: '
  };
// Apply conversions
for (const [cypressPattern, playwrightPattern] of Object.entries(conversions)) {
    playwrightContent = playwrightContent.replace(
      new RegExp(cypressPattern, 'g'),
      playwrightPattern
    );
  }

  // Detect test type and add necessary imports and setup
  const testTypes = {
    api: /cy\.request|cy\.intercept|\.then\s*\(\s*{\s*status/i,
    component: /cy\.mount|mount\(/i,
    accessibility: /cy\.injectAxe|cy\.checkA11y|aria-|role=/i,
    visual: /cy\.screenshot|matchImageSnapshot/i,
    performance: /cy\.lighthouse|performance\.|timing/i,
    mobile: /viewport|mobile|touch|swipe/i,
    network: /intercept|wait|route|stub/i,
    auth: /login|auth|session|token/i,
    database: /db|database|query|model/i,
    file: /upload|download|readFile|writeFile/i,
    iframe: /iframe|frame/i,
    storage: /localStorage|sessionStorage|cookie/i,
    state: /store|redux|vuex|state/i,
    custom: /Cypress\.Commands\.add/i
  };

  // Detect test types
  const detectedTypes = Object.entries(testTypes)
    .filter(([_, pattern]) => pattern.test(cypressContent))
    .map(([type]) => type);

  // Generate imports based on test types
  let requiredImports = [`import { test, expect } from '@playwright/test';`];
  
  if (detectedTypes.includes('api')) {
    requiredImports.push(`import { request } from '@playwright/test';`);
  }
  
  if (detectedTypes.includes('component')) {
    requiredImports.push(`import { mount } from '@playwright/experimental-ct-react';`);
  }
  
  if (detectedTypes.includes('accessibility')) {
    requiredImports.push(`import { injectAxe, checkA11y } from 'axe-playwright';`);
  }
  
  if (detectedTypes.includes('file')) {
    requiredImports.push(`import fs from 'fs/promises';`);
    requiredImports.push(`import path from 'path';`);
  }

  // Add test type specific setup
  let setup = '';
  if (detectedTypes.length > 0) {
    setup = `
// Test type: ${detectedTypes.join(', ')}
test.describe.configure({ mode: 'parallel' });
`;
  }

  // Clean up remaining issues
  playwrightContent = playwrightContent
    // Make test callbacks async and include page parameter
    .replace(/test\((.*?),\s*\((.*?)\)\s*=>/g, 'test($1, async ({ page' + 
      (detectedTypes.includes('api') ? ', request' : '') + 
      ' }) =>')
    // Clean up any remaining text after the last closing brace
    .replace(/}[^}]*$/, '});')
    // Fix any remaining vistest to goto
    .replace(/vistest\(/g, 'goto(')
    // Remove any XML-style tags and their content
    .replace(/<\/?userStyle[^>]*>.*?<\/userStyle>/g, '')
    // Remove any other XML-style tags
    .replace(/<[^>]+>/g, '')
    // Remove any stray characters and whitespace at the end
    .replace(/[%\$#@\s]+$/, '')
    // Add final newline
    .trim() + '\n';

  // Combine imports, setup, and converted content
  return requiredImports.join('\n') + '\n\n' + setup + playwrightContent;
}

/**
 * Convert custom Cypress commands to Playwright helper functions
 * @param {string} content - Cypress test content
 * @returns {Object} - Extracted commands and helper functions
 */
function convertCustomCommands(content) {
    const customCommands = {};
    const commandRegex = /Cypress\.Commands\.add\(['"](.*?)['"],\s*function\s*\((.*?)\)\s*{([\s\S]*?)}\)/g;
    let match;
  
    while ((match = commandRegex.exec(content)) !== null) {
      const [_, commandName, params, commandBody] = match;
      customCommands[commandName] = {
        params: params.split(',').map(p => p.trim()),
        body: commandBody.trim()
      };
    }
  
    // Convert commands to Playwright helper functions
    const helperFunctions = Object.entries(customCommands).map(([name, command]) => {
      const functionBody = command.body
        .replace(/cy\./g, 'await page.')
        .replace(/\.should\(/g, '.expect(')
        .replace(/\.and\(/g, '.expect(');
  
      return `
  async function ${name}(page${command.params.length ? ', ' + command.params.join(', ') : ''}) {
    ${functionBody}
  }`;
    });
  
    return {
      commands: customCommands,
      helpers: helperFunctions.join('\n\n')
    };
  }
  
  /**
   * Convert Cypress configuration to Playwright configuration
   * @param {string} configPath - Path to cypress.json
   * @returns {string} - Playwright config content
   */
  async function convertConfig(configPath) {
    try {
      const cypressConfig = JSON.parse(await fs.readFile(configPath, 'utf8'));
      
      const playwrightConfig = {
        testDir: './tests',
        timeout: cypressConfig.defaultCommandTimeout || 4000,
        expect: {
          timeout: cypressConfig.defaultCommandTimeout || 4000,
        },
        use: {
          baseURL: cypressConfig.baseUrl,
          viewport: cypressConfig.viewportWidth && cypressConfig.viewportHeight 
            ? { width: cypressConfig.viewportWidth, height: cypressConfig.viewportHeight }
            : undefined,
          video: cypressConfig.video ? 'on' : 'off',
          screenshot: cypressConfig.screenshotOnFailure ? 'only-on-failure' : 'off',
        },
        projects: [
          {
            name: 'chromium',
            use: { browserName: 'chromium' },
          },
          {
            name: 'firefox',
            use: { browserName: 'firefox' },
          },
          {
            name: 'webkit',
            use: { browserName: 'webkit' },
          },
        ],
      };
  
      return `
  import { defineConfig, devices } from '@playwright/test';
  
  export default ${JSON.stringify(playwrightConfig, null, 2)};
  `;
    } catch (error) {
      console.error(chalk.yellow('Warning: Could not convert Cypress config:'), error.message);
      return null;
    }
  }
  
  /**
   * Process plugins and support files
   * @param {string} cypressDir - Cypress directory path
   * @returns {Object} - Converted support files and plugins
   */
  async function processSupport(cypressDir) {
    const results = {
      support: [],
      plugins: [],
    };
  
    try {
      // Convert support files
      const supportPath = path.join(cypressDir, 'support');
      if (await fs.access(supportPath).then(() => true).catch(() => false)) {
        const files = await fs.readdir(supportPath);
        for (const file of files) {
          const content = await fs.readFile(path.join(supportPath, file), 'utf8');
          const converted = await convertCypressToPlaywright(content);
          results.support.push({
            name: file.replace(/\.js$/, '.ts'),
            content: converted
          });
        }
      }
  
      // Convert plugins
      const pluginsPath = path.join(cypressDir, 'plugins');
      if (await fs.access(pluginsPath).then(() => true).catch(() => false)) {
        const files = await fs.readdir(pluginsPath);
        for (const file of files) {
          // Handle plugin conversion specifically
          // This might need manual review as plugins often need different approaches in Playwright
          results.plugins.push({
            name: file,
            needsReview: true
          });
        }
      }
    } catch (error) {
      console.error(chalk.yellow('Warning: Error processing support files:'), error.message);
    }
  
    return results;
  }
  
  /**
   * Enhanced file conversion with error recovery
   * @param {string} sourcePath - Path to source Cypress file
   * @param {string} outputPath - Path for output Playwright file
   */
  export async function convertFile(sourcePath, outputPath) {
    try {
      const content = await fs.readFile(sourcePath, 'utf8');
      
      // Extract and convert custom commands first
      const { helpers } = convertCustomCommands(content);
      
      // Convert main test content
      const converted = await convertCypressToPlaywright(content);
      
      // Combine helpers and converted content
      const finalContent = helpers ? `${helpers}\n\n${converted}` : converted;
      
      // Ensure output directory exists
      await fs.mkdir(path.dirname(outputPath), { recursive: true });
      
      // Write converted file
      await fs.writeFile(outputPath, finalContent);
      
      console.log(chalk.green(`✓ Successfully converted ${path.basename(sourcePath)}`));
      
      // Check for potential issues
      const potentialIssues = analyzeConversion(finalContent);
      if (potentialIssues.length > 0) {
        console.log(chalk.yellow('\nPotential issues to review:'));
        potentialIssues.forEach(issue => {
          console.log(chalk.yellow(`- ${issue}`));
        });
      }
  
    } catch (error) {
      console.error(chalk.red(`Error converting ${sourcePath}:`), error.message);
      throw error;
    }
  }
  
  /**
   * Analyze converted content for potential issues
   * @param {string} content - Converted test content
   * @returns {string[]} - Array of potential issues
   */
  function analyzeConversion(content) {
    const issues = [];
    
    if (content.includes('cy.')) {
      issues.push('Some Cypress commands may not have been converted');
    }
    
    if (content.includes('should(')) {
      issues.push('Some Cypress assertions may need manual review');
    }
    
    if (content.includes('Cypress.Commands.add')) {
      issues.push('Custom commands found - verify helper function conversion');
    }
    
    return issues;
  }