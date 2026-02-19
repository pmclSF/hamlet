import { browser, $, expect } from '@wdio/globals';

describe('Login Flow', () => {
  beforeEach(async () => {
    await browser.url('/login');
  });

  it('should login successfully', async () => {
    await $('#username').setValue('admin');
    await $('#password').setValue('pass123');
    await $('#login-btn').click();
    await expect(browser).toHaveUrl('http://localhost/dashboard');
    await expect($('#welcome')).toBeDisplayed();
    await expect($('#welcome')).toHaveText('Welcome, admin');
  });
});
