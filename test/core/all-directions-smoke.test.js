/**
 * All-directions smoke test.
 *
 * For every conversion direction returned by ConverterFactory.getSupportedConversions(),
 * creates a converter and converts a minimal fixture, asserting non-empty output.
 */
import { ConverterFactory } from '../../src/core/ConverterFactory.js';

/**
 * Minimal valid inputs per source framework.
 * Each must be enough for the converter to produce non-empty output.
 */
const FIXTURES = {
  cypress: `describe('test', () => {
  it('works', () => {
    cy.visit('/');
    cy.get('#btn').click();
  });
});
`,
  playwright: `import { test, expect } from '@playwright/test';
test.describe('test', () => {
  test('works', async ({ page }) => {
    await page.goto('/');
    await page.locator('#btn').click();
  });
});
`,
  selenium: `const { Builder, By } = require('selenium-webdriver');
describe('test', () => {
  it('works', async () => {
    const driver = await new Builder().forBrowser('chrome').build();
    await driver.get('http://localhost');
    await driver.findElement(By.css('#btn')).click();
  });
});
`,
  jest: `describe('test', () => {
  it('works', () => {
    expect(1 + 1).toBe(2);
  });
});
`,
  vitest: `import { describe, it, expect } from 'vitest';
describe('test', () => {
  it('works', () => {
    expect(1 + 1).toBe(2);
  });
});
`,
  mocha: `const { expect } = require('chai');
describe('test', () => {
  it('works', () => {
    expect(1 + 1).to.equal(2);
  });
});
`,
  jasmine: `describe('test', () => {
  it('works', () => {
    const spy = jasmine.createSpy();
    spy();
    expect(spy).toHaveBeenCalled();
  });
});
`,
  webdriverio: `describe('test', () => {
  it('works', async () => {
    await browser.url('/');
    await $('#btn').click();
  });
});
`,
  puppeteer: `const puppeteer = require('puppeteer');
describe('test', () => {
  it('works', async () => {
    const browser = await puppeteer.launch();
    const page = await browser.newPage();
    await page.goto('/');
    await page.click('#btn');
    await browser.close();
  });
});
`,
  testcafe: `import { Selector } from 'testcafe';
fixture\`Test\`.page\`http://localhost\`;
test('works', async t => {
  await t.click('#btn');
});
`,
  junit4: `import org.junit.Test;
import static org.junit.Assert.*;

public class MyTest {
    @Test
    public void testBasic() {
        assertEquals(1, 1);
    }
}
`,
  junit5: `import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

public class MyTest {
    @Test
    public void testBasic() {
        Assertions.assertEquals(1, 1);
    }
}
`,
  testng: `import org.testng.annotations.Test;
import org.testng.Assert;

public class MyTest {
    @Test
    public void testBasic() {
        Assert.assertEquals(1, 1);
    }
}
`,
  pytest: `import pytest

def test_basic():
    assert 1 == 1
`,
  unittest: `import unittest

class TestBasic(unittest.TestCase):
    def test_basic(self):
        self.assertEqual(1, 1)
`,
  nose2: `from nose.tools import assert_equal

def test_basic():
    assert_equal(1, 1)
`,
};

describe('All 25 conversion directions smoke test', () => {
  const directions = ConverterFactory.getSupportedConversions();

  it('should report exactly 25 supported directions', () => {
    expect(directions).toHaveLength(25);
  });

  describe.each(directions)('%s', (direction) => {
    const [from, to] = direction.split('-');

    it('should create a converter without error', async () => {
      const converter = await ConverterFactory.createConverter(from, to);
      expect(converter).toBeDefined();
      expect(converter.convert).toBeInstanceOf(Function);
    });

    it('should produce non-empty output from minimal fixture', async () => {
      const fixture = FIXTURES[from];
      expect(fixture).toBeDefined();

      const converter = await ConverterFactory.createConverter(from, to);
      const output = await converter.convert(fixture);

      expect(typeof output).toBe('string');
      expect(output.trim().length).toBeGreaterThan(0);
    });
  });
});
