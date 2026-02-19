const puppeteer = require('puppeteer');

describe('click actions', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should interact', async () => {
    await page.goto('/app');
    await page.click('#btn');
    await page.click('#dbl', { clickCount: 2 });
    await page.hover('#menu');
  });
});
