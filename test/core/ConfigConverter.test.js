import { ConfigConverter } from '../../src/core/ConfigConverter.js';

describe('ConfigConverter', () => {
  let converter;

  beforeEach(() => {
    converter = new ConfigConverter();
  });

  describe('Jest → Vitest', () => {
    it('should convert testEnvironment', () => {
      const input = `module.exports = { testEnvironment: 'jsdom' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('vitest/config');
      expect(result).toContain("environment: 'jsdom'");
    });

    it('should convert testEnvironment node', () => {
      const input = `module.exports = { testEnvironment: 'node' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain("environment: 'node'");
    });

    it('should convert setupFiles', () => {
      const input = `module.exports = { setupFiles: './setup.js' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('setupFiles');
    });

    it('should convert testTimeout', () => {
      const input = `module.exports = { testTimeout: 30000 };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('testTimeout');
      expect(result).toContain('30000');
    });

    it('should convert clearMocks', () => {
      const input = `module.exports = { clearMocks: true };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('clearMocks');
      expect(result).toContain('true');
    });

    it('should add HAMLET-TODO for unrecognized keys', () => {
      const input = `module.exports = { testEnvironment: 'node', moduleNameMapper: './mappers' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('moduleNameMapper');
    });

    it('should handle export default syntax', () => {
      const input = `export default { testEnvironment: 'node' };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain("environment: 'node'");
    });

    it('should handle empty config', () => {
      const input = `module.exports = {};`;
      const result = converter.convert(input, 'jest', 'vitest');

      // Should still produce valid vitest config structure
      expect(result).toContain('defineConfig');
    });
  });

  describe('Vitest → Jest', () => {
    it('should convert environment to testEnvironment', () => {
      const input = `export default defineConfig({ test: { environment: 'jsdom' } });`;
      const result = converter.convert(input, 'vitest', 'jest');

      expect(result).toContain('module.exports');
      expect(result).toContain("testEnvironment: 'jsdom'");
    });

    it('should convert testTimeout', () => {
      const input = `export default defineConfig({ test: { testTimeout: 10000 } });`;
      const result = converter.convert(input, 'vitest', 'jest');

      expect(result).toContain('testTimeout: 10000');
    });

    it('should convert clearMocks and restoreMocks', () => {
      const input = `export default defineConfig({ test: { clearMocks: true, restoreMocks: true } });`;
      const result = converter.convert(input, 'vitest', 'jest');

      expect(result).toContain('clearMocks: true');
      expect(result).toContain('restoreMocks: true');
    });

    it('should add HAMLET-TODO for unsupported keys', () => {
      const input = `export default defineConfig({ test: { environment: 'node', deps: 'inline' } });`;
      const result = converter.convert(input, 'vitest', 'jest');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('deps');
    });
  });

  describe('Cypress → Playwright', () => {
    it('should convert baseUrl', () => {
      const input = `module.exports = { baseUrl: 'http://localhost:3000' };`;
      const result = converter.convert(input, 'cypress', 'playwright');

      expect(result).toContain('@playwright/test');
      expect(result).toContain('baseURL');
    });

    it('should convert viewportWidth and viewportHeight', () => {
      const input = `module.exports = { viewportWidth: 1280, viewportHeight: 720 };`;
      const result = converter.convert(input, 'cypress', 'playwright');

      expect(result).toContain('viewport');
      expect(result).toContain('1280');
    });

    it('should convert retries', () => {
      const input = `module.exports = { retries: 2 };`;
      const result = converter.convert(input, 'cypress', 'playwright');

      expect(result).toContain('retries');
      expect(result).toContain('2');
    });

    it('should convert defaultCommandTimeout', () => {
      const input = `module.exports = { defaultCommandTimeout: 10000 };`;
      const result = converter.convert(input, 'cypress', 'playwright');

      expect(result).toContain('timeout: 10000');
    });

    it('should add HAMLET-TODO for unrecognized Cypress keys', () => {
      const input = `module.exports = { baseUrl: 'http://localhost', chromeWebSecurity: false };`;
      const result = converter.convert(input, 'cypress', 'playwright');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('chromeWebSecurity');
    });
  });

  describe('Playwright → Cypress', () => {
    it('should convert baseURL to baseUrl', () => {
      const input = `export default defineConfig({ baseURL: 'http://localhost:3000' });`;
      const result = converter.convert(input, 'playwright', 'cypress');

      expect(result).toContain('defineConfig');
      expect(result).toContain('e2e');
      expect(result).toContain("baseUrl: 'http://localhost:3000'");
    });

    it('should convert timeout to defaultCommandTimeout', () => {
      const input = `export default defineConfig({ timeout: 30000 });`;
      const result = converter.convert(input, 'playwright', 'cypress');

      expect(result).toContain('defaultCommandTimeout: 30000');
    });

    it('should convert retries', () => {
      const input = `export default defineConfig({ retries: 3 });`;
      const result = converter.convert(input, 'playwright', 'cypress');

      expect(result).toContain('retries: 3');
    });

    it('should add HAMLET-TODO for unsupported keys', () => {
      const input = `export default defineConfig({ baseURL: '/', workers: 4 });`;
      const result = converter.convert(input, 'playwright', 'cypress');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('workers');
    });
  });

  describe('WebdriverIO → Playwright', () => {
    it('should convert baseUrl to use.baseURL', () => {
      const input = `exports.config = { baseUrl: 'http://localhost:3000' };`;
      const result = converter.convert(input, 'webdriverio', 'playwright');

      expect(result).toContain('@playwright/test');
      expect(result).toContain("use.baseURL: 'http://localhost:3000'");
    });

    it('should convert waitforTimeout to timeout', () => {
      const input = `exports.config = { waitforTimeout: 10000 };`;
      const result = converter.convert(input, 'webdriverio', 'playwright');

      expect(result).toContain('timeout: 10000');
    });

    it('should convert maxInstances to workers', () => {
      const input = `exports.config = { maxInstances: 5 };`;
      const result = converter.convert(input, 'webdriverio', 'playwright');

      expect(result).toContain('workers: 5');
    });

    it('should add HAMLET-TODO for unsupported keys', () => {
      const input = `exports.config = { baseUrl: '/', framework: 'mocha' };`;
      const result = converter.convert(input, 'webdriverio', 'playwright');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('framework');
    });
  });

  describe('Playwright → WebdriverIO', () => {
    it('should convert baseURL to baseUrl', () => {
      const input = `export default defineConfig({ baseURL: 'http://localhost:3000' });`;
      const result = converter.convert(input, 'playwright', 'webdriverio');

      expect(result).toContain('exports.config');
      expect(result).toContain("baseUrl: 'http://localhost:3000'");
    });

    it('should convert timeout to waitforTimeout', () => {
      const input = `export default defineConfig({ timeout: 30000 });`;
      const result = converter.convert(input, 'playwright', 'webdriverio');

      expect(result).toContain('waitforTimeout: 30000');
    });

    it('should convert workers to maxInstances', () => {
      const input = `export default defineConfig({ workers: 4 });`;
      const result = converter.convert(input, 'playwright', 'webdriverio');

      expect(result).toContain('maxInstances: 4');
    });
  });

  describe('WebdriverIO → Cypress', () => {
    it('should convert baseUrl', () => {
      const input = `exports.config = { baseUrl: 'http://localhost:3000' };`;
      const result = converter.convert(input, 'webdriverio', 'cypress');

      expect(result).toContain('defineConfig');
      expect(result).toContain('e2e');
      expect(result).toContain("baseUrl: 'http://localhost:3000'");
    });

    it('should convert waitforTimeout to defaultCommandTimeout', () => {
      const input = `exports.config = { waitforTimeout: 10000 };`;
      const result = converter.convert(input, 'webdriverio', 'cypress');

      expect(result).toContain('defaultCommandTimeout: 10000');
    });

    it('should convert specs to specPattern', () => {
      const input = `exports.config = { specs: './test/specs/**/*.js' };`;
      const result = converter.convert(input, 'webdriverio', 'cypress');

      expect(result).toContain('specPattern');
    });
  });

  describe('Cypress → WebdriverIO', () => {
    it('should convert baseUrl', () => {
      const input = `module.exports = { baseUrl: 'http://localhost:3000' };`;
      const result = converter.convert(input, 'cypress', 'webdriverio');

      expect(result).toContain('exports.config');
      expect(result).toContain("baseUrl: 'http://localhost:3000'");
    });

    it('should convert defaultCommandTimeout to waitforTimeout', () => {
      const input = `module.exports = { defaultCommandTimeout: 10000 };`;
      const result = converter.convert(input, 'cypress', 'webdriverio');

      expect(result).toContain('waitforTimeout: 10000');
    });

    it('should convert specPattern to specs', () => {
      const input = `module.exports = { specPattern: 'cypress/e2e/**/*.cy.js' };`;
      const result = converter.convert(input, 'cypress', 'webdriverio');

      expect(result).toContain('specs');
    });
  });

  describe('Mocha → Jest', () => {
    it('should convert YAML config', () => {
      const input = `timeout: 5000\nspec: ./test/**/*.test.js\nbail: true`;
      const result = converter.convert(input, 'mocha', 'jest');

      expect(result).toContain('module.exports');
      expect(result).toContain('testTimeout: 5000');
      expect(result).toContain('bail: true');
    });

    it('should convert timeout from JSON format', () => {
      const input = `{ "timeout": 5000, "spec": "./test/**/*.js" }`;
      const result = converter.convert(input, 'mocha', 'jest');

      expect(result).toContain('testTimeout: 5000');
    });

    it('should convert require to setupFiles', () => {
      const input = `timeout: 10000\nrequire: ./setup.js`;
      const result = converter.convert(input, 'mocha', 'jest');

      expect(result).toContain('setupFiles');
    });

    it('should add HAMLET-TODO for unsupported keys', () => {
      const input = `timeout: 5000\nreporter: spec`;
      const result = converter.convert(input, 'mocha', 'jest');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('reporter');
    });
  });

  describe('Jasmine → Jest', () => {
    it('should convert JSON config', () => {
      const input = `{ "spec_dir": "spec", "spec_files": ["**/*[sS]pec.js"], "helpers": ["helpers/**/*.js"] }`;
      const result = converter.convert(input, 'jasmine', 'jest');

      expect(result).toContain('module.exports');
      expect(result).toContain("roots: 'spec'");
    });

    it('should convert spec_files to testMatch', () => {
      const input = `{ "spec_files": ["**/*Spec.js"] }`;
      const result = converter.convert(input, 'jasmine', 'jest');

      expect(result).toContain('testMatch');
    });

    it('should convert random to randomize', () => {
      const input = `{ "random": true }`;
      const result = converter.convert(input, 'jasmine', 'jest');

      expect(result).toContain('randomize: true');
    });

    it('should add HAMLET-TODO for unsupported keys', () => {
      const input = `{ "spec_dir": "spec", "stopSpecOnExpectationFailure": true }`;
      const result = converter.convert(input, 'jasmine', 'jest');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('stopSpecOnExpectationFailure');
    });
  });

  describe('pytest → unittest', () => {
    it('should convert pytest.ini with testpaths', () => {
      const input = `[pytest]\ntestpaths = tests\npython_files = test_*.py`;
      const result = converter.convert(input, 'pytest', 'unittest');

      expect(result).toContain('unittest configuration');
      expect(result).toContain('python -m unittest discover -s tests');
      expect(result).toContain('test_*.py');
    });

    it('should handle missing testpaths gracefully', () => {
      const input = `[pytest]\npython_files = test_*.py`;
      const result = converter.convert(input, 'pytest', 'unittest');

      expect(result).toContain('python -m unittest discover');
    });

    it('should add HAMLET-TODO for unsupported pytest keys', () => {
      const input = `[pytest]\ntestpaths = tests\naddopts = -v --tb=short`;
      const result = converter.convert(input, 'pytest', 'unittest');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('addopts');
    });
  });

  describe('TestNG → JUnit5', () => {
    it('should convert testng.xml with classes', () => {
      const input = `<?xml version="1.0"?>
<suite name="My Suite">
  <test name="Test1">
    <classes>
      <class name="com.example.LoginTest"/>
      <class name="com.example.SignupTest"/>
    </classes>
  </test>
</suite>`;
      const result = converter.convert(input, 'testng', 'junit5');

      expect(result).toContain('@Suite');
      expect(result).toContain('@SelectClasses');
      expect(result).toContain('com.example.LoginTest.class');
      expect(result).toContain('com.example.SignupTest.class');
      expect(result).toContain('My Suite');
    });

    it('should handle suite with no classes', () => {
      const input = `<suite name="Empty Suite"><test name="T1"><packages><package name="com.example"/></packages></test></suite>`;
      const result = converter.convert(input, 'testng', 'junit5');

      expect(result).toContain('@Suite');
      expect(result).toContain('Empty Suite');
    });

    it('should sanitize class name from suite name', () => {
      const input = `<suite name="My Test-Suite 2"><test name="T"><classes><class name="Foo"/></classes></test></suite>`;
      const result = converter.convert(input, 'testng', 'junit5');

      expect(result).toContain('class MyTestSuite2Test');
    });
  });

  describe('JUnit4 → JUnit5', () => {
    it('should update Maven dependencies', () => {
      const input = `<dependency>
    <groupId>junit</groupId>
    <artifactId>junit</artifactId>
    <version>4.13.2</version>
</dependency>`;
      const result = converter.convert(input, 'junit4', 'junit5');

      expect(result).toContain('org.junit.jupiter');
      expect(result).toContain('junit-jupiter');
      expect(result).toContain('5.10.0');
      expect(result).not.toContain('4.13.2');
    });

    it('should update Gradle dependencies', () => {
      const input = `testImplementation 'junit:junit:4.13.2'`;
      const result = converter.convert(input, 'junit4', 'junit5');

      expect(result).toContain('org.junit.jupiter:junit-jupiter:5.10.0');
      expect(result).not.toContain('junit:junit:4');
    });

    it('should handle Gradle testCompile', () => {
      const input = `testCompile 'junit:junit:4.12'`;
      const result = converter.convert(input, 'junit4', 'junit5');

      expect(result).toContain('testImplementation');
      expect(result).toContain('junit-jupiter');
    });

    it('should add HAMLET-TODO for unrecognized build format', () => {
      const input = `some_unknown_format { junit 4.13 }`;
      const result = converter.convert(input, 'junit4', 'junit5');

      expect(result).toContain('HAMLET-TODO');
    });
  });

  describe('edge cases', () => {
    it('should handle config with JS logic (conditional) by adding HAMLET-TODO', () => {
      const input = `const env = process.env.NODE_ENV;\nmodule.exports = env === 'ci' ? { retries: 3 } : { retries: 0 };`;
      const result = converter.convert(input, 'jest', 'vitest');

      expect(result).toContain('HAMLET-TODO');
    });

    it('should handle nested config (projects array) with HAMLET-TODO', () => {
      const input = `module.exports = { projects: [{ displayName: 'unit' }, { displayName: 'e2e' }] };`;
      const result = converter.convert(input, 'jest', 'vitest');

      // projects is unsupported → should have HAMLET-TODO
      expect(result).toContain('HAMLET-TODO');
    });

    it('should handle unsupported conversion direction', () => {
      const input = `module.exports = { baseUrl: '/' };`;
      const result = converter.convert(input, 'selenium', 'playwright');

      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('Manual action required');
    });

    it('should parse WDIO exports.config format', () => {
      const input = `exports.config = { baseUrl: 'http://localhost:3000', waitforTimeout: 10000 };`;
      const result = converter.convert(input, 'webdriverio', 'playwright');

      expect(result).toContain('@playwright/test');
      expect(result).toContain('use.baseURL');
      expect(result).toContain('timeout: 10000');
    });
  });
});
