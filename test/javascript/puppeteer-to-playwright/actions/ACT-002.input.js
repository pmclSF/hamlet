const puppeteer = require('puppeteer');

describe('click action', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should click', async () => {
    await page.goto('/app');
    await page.click('#submit');
    await page.hover('#menu');
  });
});
