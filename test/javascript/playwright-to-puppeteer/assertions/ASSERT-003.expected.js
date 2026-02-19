const puppeteer = require('puppeteer');

describe('text assertions', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should check text and value', async () => {
    await page.goto('/form');
    expect(await page.$eval('#msg', el => el.textContent)).toBe('Hello');
    expect(await page.$eval('#input', el => el.value)).toBe('test');
  });
});
