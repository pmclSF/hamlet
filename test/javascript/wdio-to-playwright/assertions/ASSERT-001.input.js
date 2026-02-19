describe('url assertions', () => {
  it('should check url', async () => {
    await browser.url('/dashboard');
    await expect(browser).toHaveUrl('http://localhost/dashboard');
  });
});
