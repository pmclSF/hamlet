import { ConverterFactory, FRAMEWORKS } from '../../src/core/ConverterFactory.js';
import { PipelineConverter } from '../../src/core/PipelineConverter.js';
import { PlaywrightToCypress } from '../../src/converters/PlaywrightToCypress.js';
import { CypressToSelenium } from '../../src/converters/CypressToSelenium.js';
import { SeleniumToCypress } from '../../src/converters/SeleniumToCypress.js';
import { PlaywrightToSelenium } from '../../src/converters/PlaywrightToSelenium.js';
import { SeleniumToPlaywright } from '../../src/converters/SeleniumToPlaywright.js';

describe('ConverterFactory', () => {
  beforeEach(() => {
    // Reset initialization state so each test starts fresh
    ConverterFactory.initialized = false;
    ConverterFactory.converters = new Map();
  });

  describe('FRAMEWORKS constant', () => {
    it('should export all sixteen frameworks', () => {
      expect(FRAMEWORKS.CYPRESS).toBe('cypress');
      expect(FRAMEWORKS.PLAYWRIGHT).toBe('playwright');
      expect(FRAMEWORKS.SELENIUM).toBe('selenium');
      expect(FRAMEWORKS.JEST).toBe('jest');
      expect(FRAMEWORKS.VITEST).toBe('vitest');
      expect(FRAMEWORKS.MOCHA).toBe('mocha');
      expect(FRAMEWORKS.JASMINE).toBe('jasmine');
      expect(FRAMEWORKS.JUNIT4).toBe('junit4');
      expect(FRAMEWORKS.JUNIT5).toBe('junit5');
      expect(FRAMEWORKS.TESTNG).toBe('testng');
      expect(FRAMEWORKS.PYTEST).toBe('pytest');
      expect(FRAMEWORKS.UNITTEST).toBe('unittest');
      expect(FRAMEWORKS.NOSE2).toBe('nose2');
      expect(FRAMEWORKS.WEBDRIVERIO).toBe('webdriverio');
      expect(FRAMEWORKS.PUPPETEER).toBe('puppeteer');
      expect(FRAMEWORKS.TESTCAFE).toBe('testcafe');
    });
  });

  describe('createConverter', () => {
    describe('pipeline-backed directions', () => {
      it('should create PipelineConverter for cypress→playwright', async () => {
        const converter = await ConverterFactory.createConverter('cypress', 'playwright');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('cypress');
        expect(converter.targetFramework).toBe('playwright');
      });

      it('should create PipelineConverter for jest→vitest', async () => {
        const converter = await ConverterFactory.createConverter('jest', 'vitest');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('jest');
        expect(converter.targetFramework).toBe('vitest');
      });

      it('should convert Cypress to Playwright through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('cypress', 'playwright');
        const result = await converter.convert(`cy.get('.btn').click();`);
        expect(result).toContain("await page.locator('.btn').click()");
      });

      it('should convert Jest to Vitest through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('jest', 'vitest');
        const result = await converter.convert(`jest.fn();`);
        expect(result).toContain('vi.fn()');
      });

      it('should create PipelineConverter for mocha→jest', async () => {
        const converter = await ConverterFactory.createConverter('mocha', 'jest');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('mocha');
        expect(converter.targetFramework).toBe('jest');
      });

      it('should create PipelineConverter for jasmine→jest', async () => {
        const converter = await ConverterFactory.createConverter('jasmine', 'jest');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('jasmine');
        expect(converter.targetFramework).toBe('jest');
      });

      it('should create PipelineConverter for jest→mocha', async () => {
        const converter = await ConverterFactory.createConverter('jest', 'mocha');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('jest');
        expect(converter.targetFramework).toBe('mocha');
      });

      it('should create PipelineConverter for jest→jasmine', async () => {
        const converter = await ConverterFactory.createConverter('jest', 'jasmine');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('jest');
        expect(converter.targetFramework).toBe('jasmine');
      });

      it('should convert Mocha+Chai to Jest through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('mocha', 'jest');
        const result = await converter.convert(
          "const { expect } = require('chai');\ndescribe('test', () => {\n  it('works', () => {\n    expect(1).to.equal(1);\n  });\n});"
        );
        expect(result).toContain('expect(1).toBe(1)');
        expect(result).not.toContain("require('chai')");
      });

      it('should convert Jasmine to Jest through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('jasmine', 'jest');
        const result = await converter.convert(
          "describe('test', () => {\n  it('works', () => {\n    const spy = jasmine.createSpy();\n    spy();\n    expect(spy).toHaveBeenCalled();\n  });\n});"
        );
        expect(result).toContain('jest.fn()');
        expect(result).not.toContain('jasmine.createSpy');
      });

      it('should create PipelineConverter for junit4→junit5', async () => {
        const converter = await ConverterFactory.createConverter('junit4', 'junit5');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('junit4');
        expect(converter.targetFramework).toBe('junit5');
      });

      it('should create PipelineConverter for junit5→testng', async () => {
        const converter = await ConverterFactory.createConverter('junit5', 'testng');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('junit5');
        expect(converter.targetFramework).toBe('testng');
      });

      it('should create PipelineConverter for testng→junit5', async () => {
        const converter = await ConverterFactory.createConverter('testng', 'junit5');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('testng');
        expect(converter.targetFramework).toBe('junit5');
      });

      it('should convert JUnit 4 to JUnit 5 through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('junit4', 'junit5');
        const result = await converter.convert(
          'import org.junit.Test;\nimport org.junit.Assert;\n\npublic class MyTest {\n    @Test\n    public void testBasic() {\n        Assert.assertEquals(1, 1);\n    }\n}'
        );
        expect(result).toContain('org.junit.jupiter.api.Test');
        expect(result).toContain('Assertions.assertEquals');
      });

      it('should convert JUnit 5 to TestNG through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('junit5', 'testng');
        const result = await converter.convert(
          'import org.junit.jupiter.api.Test;\nimport org.junit.jupiter.api.Assertions;\n\npublic class MyTest {\n    @Test\n    public void testBasic() {\n        Assertions.assertTrue(true);\n    }\n}'
        );
        expect(result).toContain('org.testng.annotations.Test');
        expect(result).toContain('Assert.assertTrue');
      });

      it('should convert TestNG to JUnit 5 through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('testng', 'junit5');
        const result = await converter.convert(
          'import org.testng.annotations.Test;\nimport org.testng.Assert;\n\npublic class MyTest {\n    @Test\n    public void testBasic() {\n        Assert.assertTrue(true);\n    }\n}'
        );
        expect(result).toContain('org.junit.jupiter.api.Test');
        expect(result).toContain('Assertions.assertTrue');
      });

      it('should create PipelineConverter for pytest→unittest', async () => {
        const converter = await ConverterFactory.createConverter('pytest', 'unittest');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('pytest');
        expect(converter.targetFramework).toBe('unittest');
      });

      it('should create PipelineConverter for unittest→pytest', async () => {
        const converter = await ConverterFactory.createConverter('unittest', 'pytest');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('unittest');
        expect(converter.targetFramework).toBe('pytest');
      });

      it('should create PipelineConverter for nose2→pytest', async () => {
        const converter = await ConverterFactory.createConverter('nose2', 'pytest');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('nose2');
        expect(converter.targetFramework).toBe('pytest');
      });

      it('should convert pytest to unittest through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('pytest', 'unittest');
        const result = await converter.convert(
          'def test_basic():\n    assert 1 == 1\n'
        );
        expect(result).toContain('unittest.TestCase');
        expect(result).toContain('self.assertEqual(1, 1)');
      });

      it('should convert unittest to pytest through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('unittest', 'pytest');
        const result = await converter.convert(
          'import unittest\n\nclass TestBasic(unittest.TestCase):\n    def test_basic(self):\n        self.assertEqual(1, 1)\n'
        );
        expect(result).toContain('assert 1 == 1');
        expect(result).not.toContain('class TestBasic');
      });

      it('should convert nose2 to pytest through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('nose2', 'pytest');
        const result = await converter.convert(
          'from nose.tools import assert_equal\n\ndef test_basic():\n    assert_equal(1, 1)\n'
        );
        expect(result).toContain('assert 1 == 1');
        expect(result).not.toContain('assert_equal');
      });

      it('should create PipelineConverter for webdriverio→playwright', async () => {
        const converter = await ConverterFactory.createConverter('webdriverio', 'playwright');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('webdriverio');
        expect(converter.targetFramework).toBe('playwright');
      });

      it('should create PipelineConverter for webdriverio→cypress', async () => {
        const converter = await ConverterFactory.createConverter('webdriverio', 'cypress');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('webdriverio');
        expect(converter.targetFramework).toBe('cypress');
      });

      it('should create PipelineConverter for playwright→webdriverio', async () => {
        const converter = await ConverterFactory.createConverter('playwright', 'webdriverio');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('playwright');
        expect(converter.targetFramework).toBe('webdriverio');
      });

      it('should create PipelineConverter for cypress→webdriverio', async () => {
        const converter = await ConverterFactory.createConverter('cypress', 'webdriverio');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('cypress');
        expect(converter.targetFramework).toBe('webdriverio');
      });

      it('should create PipelineConverter for puppeteer→playwright', async () => {
        const converter = await ConverterFactory.createConverter('puppeteer', 'playwright');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('puppeteer');
        expect(converter.targetFramework).toBe('playwright');
      });

      it('should create PipelineConverter for playwright→puppeteer', async () => {
        const converter = await ConverterFactory.createConverter('playwright', 'puppeteer');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('playwright');
        expect(converter.targetFramework).toBe('puppeteer');
      });

      it('should create PipelineConverter for testcafe→playwright', async () => {
        const converter = await ConverterFactory.createConverter('testcafe', 'playwright');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('testcafe');
        expect(converter.targetFramework).toBe('playwright');
      });

      it('should create PipelineConverter for testcafe→cypress', async () => {
        const converter = await ConverterFactory.createConverter('testcafe', 'cypress');
        expect(converter).toBeInstanceOf(PipelineConverter);
        expect(converter.sourceFramework).toBe('testcafe');
        expect(converter.targetFramework).toBe('cypress');
      });

      it('should convert WebdriverIO to Playwright through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('webdriverio', 'playwright');
        const result = await converter.convert(
          "describe('test', () => {\n  it('works', async () => {\n    await browser.url('/page');\n    await $('#btn').click();\n  });\n});"
        );
        expect(result).toContain("await page.goto('/page')");
        expect(result).toContain("await page.locator('#btn').click()");
      });

      it('should convert WebdriverIO to Cypress through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('webdriverio', 'cypress');
        const result = await converter.convert(
          "describe('test', () => {\n  it('works', async () => {\n    await browser.url('/page');\n    await $('#btn').click();\n  });\n});"
        );
        expect(result).toContain("cy.visit('/page')");
        expect(result).toContain("cy.get('#btn').click()");
      });

      it('should convert Playwright to WebdriverIO through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('playwright', 'webdriverio');
        const result = await converter.convert(
          "import { test, expect } from '@playwright/test';\n\ntest.describe('test', () => {\n  test('works', async ({ page }) => {\n    await page.goto('/page');\n    await page.locator('#btn').click();\n  });\n});"
        );
        expect(result).toContain("await browser.url('/page')");
        expect(result).toContain("await $('#btn').click()");
      });

      it('should convert Cypress to WebdriverIO through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('cypress', 'webdriverio');
        const result = await converter.convert(
          "describe('test', () => {\n  it('works', () => {\n    cy.visit('/page');\n    cy.get('#btn').click();\n  });\n});"
        );
        expect(result).toContain("await browser.url('/page')");
        expect(result).toContain("await $('#btn').click()");
      });

      it('should convert Puppeteer to Playwright through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('puppeteer', 'playwright');
        const result = await converter.convert(
          "const puppeteer = require('puppeteer');\n\ndescribe('test', () => {\n  let browser, page;\n\n  beforeAll(async () => {\n    browser = await puppeteer.launch();\n    page = await browser.newPage();\n  });\n\n  afterAll(async () => {\n    await browser.close();\n  });\n\n  it('works', async () => {\n    await page.goto('/page');\n    await page.click('#btn');\n  });\n});"
        );
        expect(result).toContain("await page.goto('/page')");
        expect(result).toContain("page.locator('#btn').click()");
        expect(result).not.toContain('puppeteer.launch');
      });

      it('should convert Playwright to Puppeteer through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('playwright', 'puppeteer');
        const result = await converter.convert(
          "import { test, expect } from '@playwright/test';\n\ntest.describe('test', () => {\n  test('works', async ({ page }) => {\n    await page.goto('/page');\n    await page.locator('#btn').click();\n  });\n});"
        );
        expect(result).toContain("await page.goto('/page')");
        expect(result).toContain("await page.click('#btn')");
        expect(result).toContain("puppeteer");
      });

      it('should convert TestCafe to Playwright through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('testcafe', 'playwright');
        const result = await converter.convert(
          "import { Selector } from 'testcafe';\n\nfixture`Test`.page`http://localhost`;\n\ntest('works', async t => {\n  await t.click('#btn');\n  await t.expect(Selector('#msg').visible).ok();\n});"
        );
        expect(result).toContain("page.locator('#btn').click()");
        expect(result).toContain('toBeVisible');
        expect(result).not.toContain('testcafe');
      });

      it('should convert TestCafe to Cypress through pipeline', async () => {
        const converter = await ConverterFactory.createConverter('testcafe', 'cypress');
        const result = await converter.convert(
          "import { Selector } from 'testcafe';\n\nfixture`Test`.page`http://localhost`;\n\ntest('works', async t => {\n  await t.click('#btn');\n  await t.expect(Selector('#msg').visible).ok();\n});"
        );
        expect(result).toContain("cy.get('#btn').click()");
        expect(result).toContain("should('be.visible')");
        expect(result).not.toContain('testcafe');
      });
    });

    describe('legacy converter directions', () => {
      it('should create PlaywrightToCypress converter', async () => {
        const converter = await ConverterFactory.createConverter('playwright', 'cypress');
        expect(converter).toBeInstanceOf(PlaywrightToCypress);
      });

      it('should create CypressToSelenium converter', async () => {
        const converter = await ConverterFactory.createConverter('cypress', 'selenium');
        expect(converter).toBeInstanceOf(CypressToSelenium);
      });

      it('should create SeleniumToCypress converter', async () => {
        const converter = await ConverterFactory.createConverter('selenium', 'cypress');
        expect(converter).toBeInstanceOf(SeleniumToCypress);
      });

      it('should create PlaywrightToSelenium converter', async () => {
        const converter = await ConverterFactory.createConverter('playwright', 'selenium');
        expect(converter).toBeInstanceOf(PlaywrightToSelenium);
      });

      it('should create SeleniumToPlaywright converter', async () => {
        const converter = await ConverterFactory.createConverter('selenium', 'playwright');
        expect(converter).toBeInstanceOf(SeleniumToPlaywright);
      });
    });

    describe('error handling', () => {
      it('should throw error for invalid source framework', async () => {
        await expect(
          ConverterFactory.createConverter('invalid', 'playwright')
        ).rejects.toThrow('Invalid source framework');
      });

      it('should throw error for invalid target framework', async () => {
        await expect(
          ConverterFactory.createConverter('cypress', 'invalid')
        ).rejects.toThrow('Invalid target framework');
      });

      it('should throw error for same source and target', async () => {
        await expect(
          ConverterFactory.createConverter('cypress', 'cypress')
        ).rejects.toThrow('Source and target frameworks must be different');
      });

      it('should throw error for unsupported direction', async () => {
        await expect(
          ConverterFactory.createConverter('jest', 'selenium')
        ).rejects.toThrow('Unsupported conversion');
      });
    });

    it('should be case insensitive', async () => {
      const converter = await ConverterFactory.createConverter('CYPRESS', 'PLAYWRIGHT');
      expect(converter).toBeInstanceOf(PipelineConverter);
    });

    it('should pass options to pipeline converter', async () => {
      const options = { batchSize: 10 };
      const converter = await ConverterFactory.createConverter('cypress', 'playwright', options);
      expect(converter.options.batchSize).toBe(10);
    });
  });

  describe('getSupportedConversions', () => {
    it('should return all 25 conversion directions', () => {
      const conversions = ConverterFactory.getSupportedConversions();
      expect(conversions).toHaveLength(25);
      expect(conversions).toContain('cypress-playwright');
      expect(conversions).toContain('playwright-cypress');
      expect(conversions).toContain('cypress-selenium');
      expect(conversions).toContain('selenium-cypress');
      expect(conversions).toContain('playwright-selenium');
      expect(conversions).toContain('selenium-playwright');
      expect(conversions).toContain('jest-vitest');
      expect(conversions).toContain('mocha-jest');
      expect(conversions).toContain('jasmine-jest');
      expect(conversions).toContain('jest-mocha');
      expect(conversions).toContain('jest-jasmine');
      expect(conversions).toContain('junit4-junit5');
      expect(conversions).toContain('junit5-testng');
      expect(conversions).toContain('testng-junit5');
      expect(conversions).toContain('pytest-unittest');
      expect(conversions).toContain('unittest-pytest');
      expect(conversions).toContain('nose2-pytest');
      expect(conversions).toContain('webdriverio-playwright');
      expect(conversions).toContain('webdriverio-cypress');
      expect(conversions).toContain('playwright-webdriverio');
      expect(conversions).toContain('cypress-webdriverio');
      expect(conversions).toContain('puppeteer-playwright');
      expect(conversions).toContain('playwright-puppeteer');
      expect(conversions).toContain('testcafe-playwright');
      expect(conversions).toContain('testcafe-cypress');
    });
  });

  describe('isSupported', () => {
    it('should return true for valid conversions', () => {
      expect(ConverterFactory.isSupported('cypress', 'playwright')).toBe(true);
      expect(ConverterFactory.isSupported('playwright', 'cypress')).toBe(true);
      expect(ConverterFactory.isSupported('cypress', 'selenium')).toBe(true);
      expect(ConverterFactory.isSupported('selenium', 'cypress')).toBe(true);
      expect(ConverterFactory.isSupported('playwright', 'selenium')).toBe(true);
      expect(ConverterFactory.isSupported('selenium', 'playwright')).toBe(true);
      expect(ConverterFactory.isSupported('jest', 'vitest')).toBe(true);
      expect(ConverterFactory.isSupported('mocha', 'jest')).toBe(true);
      expect(ConverterFactory.isSupported('jasmine', 'jest')).toBe(true);
      expect(ConverterFactory.isSupported('jest', 'mocha')).toBe(true);
      expect(ConverterFactory.isSupported('jest', 'jasmine')).toBe(true);
      expect(ConverterFactory.isSupported('junit4', 'junit5')).toBe(true);
      expect(ConverterFactory.isSupported('junit5', 'testng')).toBe(true);
      expect(ConverterFactory.isSupported('testng', 'junit5')).toBe(true);
      expect(ConverterFactory.isSupported('pytest', 'unittest')).toBe(true);
      expect(ConverterFactory.isSupported('unittest', 'pytest')).toBe(true);
      expect(ConverterFactory.isSupported('nose2', 'pytest')).toBe(true);
      expect(ConverterFactory.isSupported('webdriverio', 'playwright')).toBe(true);
      expect(ConverterFactory.isSupported('webdriverio', 'cypress')).toBe(true);
      expect(ConverterFactory.isSupported('playwright', 'webdriverio')).toBe(true);
      expect(ConverterFactory.isSupported('cypress', 'webdriverio')).toBe(true);
      expect(ConverterFactory.isSupported('puppeteer', 'playwright')).toBe(true);
      expect(ConverterFactory.isSupported('playwright', 'puppeteer')).toBe(true);
      expect(ConverterFactory.isSupported('testcafe', 'playwright')).toBe(true);
      expect(ConverterFactory.isSupported('testcafe', 'cypress')).toBe(true);
    });

    it('should return false for invalid conversions', () => {
      expect(ConverterFactory.isSupported('invalid', 'playwright')).toBe(false);
      expect(ConverterFactory.isSupported('cypress', 'invalid')).toBe(false);
      expect(ConverterFactory.isSupported('cypress', 'cypress')).toBe(false);
      expect(ConverterFactory.isSupported('jest', 'selenium')).toBe(false);
    });
  });
});
