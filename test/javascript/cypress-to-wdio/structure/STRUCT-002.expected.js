describe('Login Flow', () => {
  beforeEach(async () => {
    await browser.url('/login');
  });

  it('should login', async () => {
    await $('#username').setValue('admin');
    await $('#password').setValue('pass123');
    await $('#login-btn').click();
    await expect(browser).toHaveUrl('http://localhost/dashboard');
    await expect($('#welcome')).toBeDisplayed();
  });
});
