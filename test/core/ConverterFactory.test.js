import { ConverterFactory, FRAMEWORKS } from '../../src/core/ConverterFactory.js';
import { CypressToPlaywright } from '../../src/converters/CypressToPlaywright.js';
import { PlaywrightToCypress } from '../../src/converters/PlaywrightToCypress.js';
import { CypressToSelenium } from '../../src/converters/CypressToSelenium.js';
import { SeleniumToCypress } from '../../src/converters/SeleniumToCypress.js';
import { PlaywrightToSelenium } from '../../src/converters/PlaywrightToSelenium.js';
import { SeleniumToPlaywright } from '../../src/converters/SeleniumToPlaywright.js';

describe('ConverterFactory', () => {
  describe('FRAMEWORKS constant', () => {
    it('should export all three frameworks', () => {
      expect(FRAMEWORKS.CYPRESS).toBe('cypress');
      expect(FRAMEWORKS.PLAYWRIGHT).toBe('playwright');
      expect(FRAMEWORKS.SELENIUM).toBe('selenium');
    });
  });

  describe('createConverter', () => {
    it('should create CypressToPlaywright converter', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      expect(converter).toBeInstanceOf(CypressToPlaywright);
    });

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

    it('should be case insensitive', async () => {
      const converter = await ConverterFactory.createConverter('CYPRESS', 'PLAYWRIGHT');
      expect(converter).toBeInstanceOf(CypressToPlaywright);
    });

    it('should pass options to converter', async () => {
      const options = { batchSize: 10 };
      const converter = await ConverterFactory.createConverter('cypress', 'playwright', options);
      expect(converter.options.batchSize).toBe(10);
    });
  });

  describe('getSupportedConversions', () => {
    it('should return all 6 conversion directions', () => {
      const conversions = ConverterFactory.getSupportedConversions();
      expect(conversions).toHaveLength(6);
      expect(conversions).toContain('cypress-playwright');
      expect(conversions).toContain('playwright-cypress');
      expect(conversions).toContain('cypress-selenium');
      expect(conversions).toContain('selenium-cypress');
      expect(conversions).toContain('playwright-selenium');
      expect(conversions).toContain('selenium-playwright');
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
    });

    it('should return false for invalid conversions', () => {
      expect(ConverterFactory.isSupported('invalid', 'playwright')).toBe(false);
      expect(ConverterFactory.isSupported('cypress', 'invalid')).toBe(false);
      expect(ConverterFactory.isSupported('cypress', 'cypress')).toBe(false);
    });
  });

  describe('createConverterSync', () => {
    it('should create converter from pre-loaded classes', () => {
      const converterClasses = {
        'cypress-playwright': CypressToPlaywright
      };
      const converter = ConverterFactory.createConverterSync('cypress', 'playwright', {}, converterClasses);
      expect(converter).toBeInstanceOf(CypressToPlaywright);
    });

    it('should pass options to converter', () => {
      const converterClasses = {
        'cypress-playwright': CypressToPlaywright
      };
      const converter = ConverterFactory.createConverterSync('cypress', 'playwright', { batchSize: 20 }, converterClasses);
      expect(converter.options.batchSize).toBe(20);
    });

    it('should throw when converter class not found', () => {
      expect(() => {
        ConverterFactory.createConverterSync('cypress', 'playwright', {}, {});
      }).toThrow('Converter not found for cypress to playwright');
    });

    it('should be case insensitive for framework names', () => {
      const converterClasses = {
        'cypress-selenium': CypressToSelenium
      };
      const converter = ConverterFactory.createConverterSync('CYPRESS', 'SELENIUM', {}, converterClasses);
      expect(converter).toBeInstanceOf(CypressToSelenium);
    });
  });

  describe('getFrameworks', () => {
    it('should return all supported frameworks', () => {
      const frameworks = ConverterFactory.getFrameworks();
      expect(frameworks).toContain('cypress');
      expect(frameworks).toContain('playwright');
      expect(frameworks).toContain('selenium');
      expect(frameworks).toHaveLength(3);
    });
  });

  describe('getConversionMatrix', () => {
    it('should return matrix with all framework combinations', () => {
      const matrix = ConverterFactory.getConversionMatrix();
      expect(matrix.cypress).toBeDefined();
      expect(matrix.playwright).toBeDefined();
      expect(matrix.selenium).toBeDefined();
    });

    it('should mark same-framework conversions as false', () => {
      const matrix = ConverterFactory.getConversionMatrix();
      expect(matrix.cypress.cypress).toBe(false);
      expect(matrix.playwright.playwright).toBe(false);
      expect(matrix.selenium.selenium).toBe(false);
    });

    it('should mark valid cross-framework conversions as true', () => {
      const matrix = ConverterFactory.getConversionMatrix();
      expect(matrix.cypress.playwright).toBe(true);
      expect(matrix.cypress.selenium).toBe(true);
      expect(matrix.playwright.cypress).toBe(true);
      expect(matrix.playwright.selenium).toBe(true);
      expect(matrix.selenium.cypress).toBe(true);
      expect(matrix.selenium.playwright).toBe(true);
    });
  });
});
