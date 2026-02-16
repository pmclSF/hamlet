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
    it('should export all ten frameworks', () => {
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
    it('should return all 14 conversion directions', () => {
      const conversions = ConverterFactory.getSupportedConversions();
      expect(conversions).toHaveLength(14);
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
    });

    it('should return false for invalid conversions', () => {
      expect(ConverterFactory.isSupported('invalid', 'playwright')).toBe(false);
      expect(ConverterFactory.isSupported('cypress', 'invalid')).toBe(false);
      expect(ConverterFactory.isSupported('cypress', 'cypress')).toBe(false);
      expect(ConverterFactory.isSupported('jest', 'selenium')).toBe(false);
    });
  });
});
