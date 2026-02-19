const puppeteer = require('puppeteer');

describe('basic suite', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should work', async () => {
    await page.goto('/test');
  });
});
