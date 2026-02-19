const puppeteer = require('puppeteer');

describe('element assertions', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should check element exists', async () => {
    await page.goto('/page');
    expect(await page.$('#visible')).toBeTruthy();
    expect(await page.$('#hidden')).toBeFalsy();
  });
});
