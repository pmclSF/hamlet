const puppeteer = require('puppeteer');

describe('hooks', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  beforeEach(async () => {
    await page.goto('/setup');
  });

  it('should use hooks', async () => {
    await page.click('#btn');
  });
});
