// Puppeteer test with Jest for a login page and visual regression
// Inspired by real-world Puppeteer tests for web applications

import puppeteer from 'puppeteer';

describe('Login Page', () => {
  let browser;
  let page;

  beforeAll(async () => {
    browser = await puppeteer.launch({ headless: 'new' });
    page = await browser.newPage();
    await page.setViewport({ width: 1280, height: 720 });
  });

  afterAll(async () => {
    await browser.close();
  });

  beforeEach(async () => {
    await page.goto('http://localhost:3000/login', { waitUntil: 'networkidle0' });
  });

  it('should render the login form with email and password fields', async () => {
    await page.waitForSelector('[data-testid="login-form"]');

    const emailField = await page.$('input[type="email"]');
    const passwordField = await page.$('input[type="password"]');
    const submitButton = await page.$('button[type="submit"]');

    expect(emailField).not.toBeNull();
    expect(passwordField).not.toBeNull();
    expect(submitButton).not.toBeNull();
  });

  it('should display a validation error for empty submission', async () => {
    await page.click('button[type="submit"]');
    await page.waitForSelector('.error-message');

    const errorText = await page.evaluate(() => {
      return document.querySelector('.error-message').textContent;
    });

    expect(errorText).toContain('Email is required');
  });

  it('should show an error for invalid credentials', async () => {
    await page.type('input[type="email"]', 'wrong@example.com');
    await page.type('input[type="password"]', 'badpassword');
    await page.click('button[type="submit"]');

    await page.waitForSelector('[data-testid="auth-error"]');

    const authError = await page.evaluate(() => {
      return document.querySelector('[data-testid="auth-error"]').textContent;
    });

    expect(authError).toBe('Invalid email or password');
  });

  it('should redirect to the dashboard after successful login', async () => {
    await page.type('input[type="email"]', 'admin@example.com');
    await page.type('input[type="password"]', 'correct-password');
    await page.click('button[type="submit"]');

    await page.waitForNavigation({ waitUntil: 'networkidle0' });

    const url = page.url();
    expect(url).toContain('/dashboard');
  });

  it('should toggle password visibility when the eye icon is clicked', async () => {
    const passwordInput = await page.$('input[type="password"]');
    expect(passwordInput).not.toBeNull();

    await page.click('[data-testid="toggle-password"]');

    const inputType = await page.evaluate(() => {
      return document.querySelector('[data-testid="password-field"]').type;
    });

    expect(inputType).toBe('text');
  });

  it('should capture a screenshot of the login page for visual review', async () => {
    await page.screenshot({ path: 'screenshots/login-page.png', fullPage: true });

    const title = await page.evaluate(() => document.title);
    expect(title).toBe('Sign In - MyApp');
  });
});
