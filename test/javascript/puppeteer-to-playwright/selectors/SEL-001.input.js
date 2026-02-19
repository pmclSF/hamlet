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
    const element = await page.$('#username');
    const elements = await page.$$('.items');
  });
});
