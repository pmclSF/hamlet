const puppeteer = require('puppeteer');

describe('mixed setup', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
    await page.goto('/setup');
    await page.setViewport({ width: 1280, height: 720 });
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should have setup applied', async () => {
    await page.click('#ready');
  });
});
