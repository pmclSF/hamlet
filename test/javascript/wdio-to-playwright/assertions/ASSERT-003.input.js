describe('text and value assertions', () => {
  it('should check text and value', async () => {
    await browser.url('/form');
    await expect($('#label')).toHaveText('Hello');
    await expect($('#label')).toHaveTextContaining('Hel');
    await expect($('#input')).toHaveValue('test');
  });
});
