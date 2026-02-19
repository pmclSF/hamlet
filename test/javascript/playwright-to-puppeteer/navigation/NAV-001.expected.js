const puppeteer = require('puppeteer');

describe('navigation', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should navigate', async () => {
    await page.goto('/dashboard');
  });
});
