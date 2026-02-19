const puppeteer = require('puppeteer');

describe('type action', () => {
  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });

  it('should type text', async () => {
    await page.goto('/form');
    await page.type('#email', 'user@test.com');
    await page.type('#password', 'secret');
  });
});
