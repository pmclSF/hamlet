const puppeteer = require('puppeteer');

describe('waits', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should wait for selector', async () => {
    await page.goto('/slow');
    await page.waitForSelector('#loaded');
    await page.click('#loaded');
  });
});
