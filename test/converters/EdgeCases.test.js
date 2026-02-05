import { ConverterFactory } from '../../src/core/ConverterFactory.js';

describe('Edge Cases - Advanced Patterns', () => {
  describe('Nested Selectors', () => {
    test('Cypress nested selectors to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Nested', () => {
  it('handles nested selectors', () => {
    cy.get('.parent').find('.child').click();
    cy.get('.container').within(() => {
      cy.get('.item').should('be.visible');
    });
    cy.get('table').find('tr').eq(2).find('td').first().click();
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('page.locator');
      expect(result).toContain('.click()');
    });

    test('Cypress nested selectors to Selenium', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'selenium');
      const input = `
describe('Nested', () => {
  it('handles nested selectors', () => {
    cy.get('.parent').find('.child').click();
    cy.get('table tr:nth-child(3) td:first-child').click();
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('driver.findElement');
      expect(result).toContain('.click()');
    });
  });

  describe('Complex Assertions', () => {
    test('Cypress chained assertions to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Assertions', () => {
  it('handles chained assertions', () => {
    cy.get('.element').should('be.visible').and('have.text', 'Hello');
    cy.get('.input').should('have.value', 'test').and('be.enabled');
    cy.get('.count').should('have.length.greaterThan', 3);
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('expect');
      expect(result).toContain('toBeVisible');
    });

    test('Playwright complex assertions to Cypress', async () => {
      const converter = await ConverterFactory.createConverter('playwright', 'cypress');
      const input = `
import { test, expect } from '@playwright/test';

test.describe('Assertions', () => {
  test('handles complex assertions', async ({ page }) => {
    await expect(page.locator('.element')).toBeVisible();
    await expect(page.locator('.element')).toHaveText('Hello');
    await expect(page.locator('.items')).toHaveCount(5);
    await expect(page).toHaveURL(/dashboard/);
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain("should('be.visible')");
      expect(result).toContain("should('have.text'");
      expect(result).toContain("should('have.length'");
    });
  });

  describe('Wait Conditions', () => {
    test('Cypress wait conditions to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Waits', () => {
  it('handles various waits', () => {
    cy.wait(1000);
    cy.wait('@apiRequest');
    cy.get('.loading').should('not.exist');
    cy.get('.content', { timeout: 10000 }).should('be.visible');
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('waitForTimeout(1000)');
    });

    test('Playwright waits to Selenium', async () => {
      const converter = await ConverterFactory.createConverter('playwright', 'selenium');
      const input = `
import { test, expect } from '@playwright/test';

test.describe('Waits', () => {
  test('handles waits', async ({ page }) => {
    await page.waitForTimeout(2000);
    await page.waitForSelector('.loaded');
    await page.waitForURL('/dashboard');
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('driver.sleep(2000)');
    });

    test('Selenium waits to Cypress', async () => {
      const converter = await ConverterFactory.createConverter('selenium', 'cypress');
      const input = `
const { Builder, By, until } = require('selenium-webdriver');

describe('Waits', () => {
  it('handles waits', async () => {
    await driver.sleep(3000);
    await driver.wait(until.elementLocated(By.css('.ready')), 10000);
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('cy.wait(3000)');
    });
  });

  describe('Form Interactions - Advanced', () => {
    test('Cypress form with multiple inputs to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Form', () => {
  it('fills complex form', () => {
    cy.get('#name').clear().type('John Doe');
    cy.get('#email').type('john@example.com', { delay: 50 });
    cy.get('input[type="file"]').selectFile('file.pdf');
    cy.get('#terms').check({ force: true });
    cy.get('#newsletter').uncheck();
    cy.get('select#country').select('USA');
    cy.get('select#city').select(['NYC', 'LA']);
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('.clear()');
      expect(result).toContain('.fill(');
      expect(result).toContain('.check()');
      expect(result).toContain('.uncheck()');
      expect(result).toContain('.selectOption(');
    });

    test('Selenium form interactions to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('selenium', 'playwright');
      const input = `
const { Builder, By } = require('selenium-webdriver');

describe('Form', () => {
  it('handles form', async () => {
    await driver.findElement(By.css('#username')).sendKeys('admin');
    await driver.findElement(By.css('#password')).sendKeys('secret');
    await driver.findElement(By.css('#submit')).click();
    const checkbox = await driver.findElement(By.css('#remember'));
      if (!(await checkbox.isSelected())) await checkbox.click();
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain(".fill('admin')");
      expect(result).toContain(".fill('secret')");
      expect(result).toContain('.click()');
      expect(result).toContain('.check()');
    });
  });

  describe('Navigation - Advanced', () => {
    test('Cypress navigation with options to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Navigation', () => {
  it('navigates with options', () => {
    cy.visit('/login', { timeout: 30000 });
    cy.visit('https://example.com', { failOnStatusCode: false });
    cy.go(-2);
    cy.reload(true);
    cy.url().should('include', '/dashboard');
    cy.title().should('eq', 'Dashboard');
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('page.goto');
      expect(result).toContain('page.goBack()');
      expect(result).toContain('page.reload()');
    });

    test('Playwright navigation to Cypress', async () => {
      const converter = await ConverterFactory.createConverter('playwright', 'cypress');
      const input = `
import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test('navigates', async ({ page }) => {
    await page.goto('/home');
    await page.goBack();
    await page.goForward();
    await page.reload();
    await expect(page).toHaveURL(/home/);
    await expect(page).toHaveTitle('Home Page');
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain("cy.visit('/home')");
      expect(result).toContain("cy.go('back')");
      expect(result).toContain("cy.go('forward')");
      expect(result).toContain('cy.reload()');
    });
  });

  describe('Element State Assertions', () => {
    test('Cypress element state to Selenium', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'selenium');
      const input = `
describe('Element State', () => {
  it('checks element states', () => {
    cy.get('#btn').should('be.disabled');
    cy.get('#input').should('be.enabled');
    cy.get('#checkbox').should('be.checked');
    cy.get('#radio').should('not.be.checked');
    cy.get('.focus-input').should('have.focus');
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('isEnabled()');
      expect(result).toContain('isSelected()');
    });

    test('Selenium element state to Cypress', async () => {
      const converter = await ConverterFactory.createConverter('selenium', 'cypress');
      const input = `
const { Builder, By } = require('selenium-webdriver');
const { expect } = require('@jest/globals');

describe('Element State', () => {
  it('checks states', async () => {
    expect(await (await driver.findElement(By.css('#btn'))).isEnabled()).toBe(false);
    expect(await (await driver.findElement(By.css('#input'))).isEnabled()).toBe(true);
    expect(await (await driver.findElement(By.css('#check'))).isSelected()).toBe(true);
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain("should('be.disabled')");
      expect(result).toContain("should('be.enabled')");
      expect(result).toContain("should('be.checked')");
    });
  });

  describe('Viewport and Window', () => {
    test('Cypress viewport to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Viewport', () => {
  it('sets viewport', () => {
    cy.viewport(1280, 720);
    cy.viewport('iphone-x');
    cy.scrollTo('bottom');
    cy.scrollTo(0, 500);
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('setViewportSize');
    });

    test('Playwright viewport to Selenium', async () => {
      const converter = await ConverterFactory.createConverter('playwright', 'selenium');
      const input = `
import { test, expect } from '@playwright/test';

test.describe('Viewport', () => {
  test('handles viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('window().setRect');
      expect(result).toContain('width: 1920');
      expect(result).toContain('height: 1080');
    });
  });

  describe('Text Content', () => {
    test('Cypress text assertions to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Text', () => {
  it('checks text content', () => {
    cy.get('.title').should('have.text', 'Welcome');
    cy.get('.desc').should('contain', 'Hello');
    cy.get('.empty').should('be.empty');
    cy.get('.message').invoke('text').should('match', /success/i);
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('toHaveText');
      expect(result).toContain('toContainText');
    });

    test('Playwright text to Selenium', async () => {
      const converter = await ConverterFactory.createConverter('playwright', 'selenium');
      const input = `
import { test, expect } from '@playwright/test';

test.describe('Text', () => {
  test('checks text', async ({ page }) => {
    await expect(page.locator('.title')).toHaveText('Welcome');
    await expect(page.locator('.desc')).toContainText('Hello');
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('getText()');
      expect(result).toContain("toBe('Welcome')");
      expect(result).toContain("toContain('Hello')");
    });
  });

  describe('Attributes and CSS', () => {
    test('Cypress attribute checks to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Attributes', () => {
  it('checks attributes', () => {
    cy.get('input').should('have.attr', 'placeholder', 'Enter name');
    cy.get('.box').should('have.css', 'background-color', 'rgb(255, 0, 0)');
    cy.get('a').should('have.attr', 'href').and('include', '/home');
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('expect');
    });

    test('Playwright attribute to Cypress', async () => {
      const converter = await ConverterFactory.createConverter('playwright', 'cypress');
      const input = `
import { test, expect } from '@playwright/test';

test.describe('Attributes', () => {
  test('checks attrs', async ({ page }) => {
    await expect(page.locator('input')).toHaveAttribute('placeholder', 'Enter name');
    await expect(page.locator('.box')).toHaveClass('active');
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain("should('have.attr'");
      expect(result).toContain("should('have.class'");
    });
  });

  describe('Special Characters in Selectors', () => {
    test('handles selectors with special characters', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Special Selectors', () => {
  it('handles special chars', () => {
    cy.get('[data-test="user-name"]').click();
    cy.get('[data-cy=submit-btn]').click();
    cy.get('#user\\\\.name').type('test');
    cy.get('.class-with-dash').should('exist');
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('page.locator');
      expect(result).toContain('data-test="user-name"');
    });

    test('handles XPath selectors', async () => {
      const converter = await ConverterFactory.createConverter('selenium', 'playwright');
      const input = `
const { Builder, By } = require('selenium-webdriver');

describe('XPath', () => {
  it('handles xpath', async () => {
    await driver.findElement(By.xpath("//button[text()='Submit']")).click();
    await driver.findElement(By.xpath("//div[@class='container']//span")).click();
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('locator');
    });
  });

  describe('Multiple Elements', () => {
    test('Cypress multiple elements to Playwright', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const input = `
describe('Multiple', () => {
  it('handles multiple elements', () => {
    cy.get('.items').should('have.length', 5);
    cy.get('.items').first().click();
    cy.get('.items').last().click();
    cy.get('.items').eq(2).click();
    cy.get('.items').each(($el, index) => {
      cy.wrap($el).should('contain', 'Item');
    });
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain('toHaveCount(5)');
      expect(result).toContain('.first()');
      expect(result).toContain('.last()');
      expect(result).toContain('.nth(2)');
    });

    test('Playwright multiple elements to Cypress', async () => {
      const converter = await ConverterFactory.createConverter('playwright', 'cypress');
      const input = `
import { test, expect } from '@playwright/test';

test.describe('Multiple', () => {
  test('handles multiple', async ({ page }) => {
    await expect(page.locator('.items')).toHaveCount(5);
    await page.locator('.items').first().click();
    await page.locator('.items').last().click();
    await page.locator('.items').nth(2).click();
  });
});`;
      const result = await converter.convert(input);
      expect(result).toContain("should('have.length', 5)");
      expect(result).toContain('.first()');
      expect(result).toContain('.last()');
      expect(result).toContain('.eq(2)');
    });
  });
});
