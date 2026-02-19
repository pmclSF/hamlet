const puppeteer = require('puppeteer');

describe('url assertions', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should check url', async () => {
    await page.goto('http://localhost/dashboard');
    expect(page.url()).toBe('http://localhost/dashboard');
  });
});
