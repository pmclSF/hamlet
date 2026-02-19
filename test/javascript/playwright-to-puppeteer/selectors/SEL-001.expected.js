const puppeteer = require('puppeteer');

describe('selectors', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should find elements', async () => {
    await page.goto('/form');
    await page.type('#username', 'test');
    await page.click('#submit');
  });
});
