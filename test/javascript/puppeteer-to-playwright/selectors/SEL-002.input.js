const puppeteer = require('puppeteer');

describe('eval selectors', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should evaluate', async () => {
    await page.goto('/page');
    const text = await page.$eval('#msg', el => el.textContent);
    const texts = await page.$$eval('.items', els => els.map(el => el.textContent));
  });
});
