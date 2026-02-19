describe('text assertions', () => {
  it('should check text', async () => {
    await browser.url('/page');
    await expect($('#msg')).toHaveText('Hello');
    await expect($('#msg')).toHaveTextContaining('Hel');
    await expect($('#input')).toHaveValue('test');
  });
});
