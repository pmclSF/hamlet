describe('selectors', () => {
  it('should find elements', async () => {
    await browser.url('/form');
    await $('#username').setValue('test');
    await $('#submit').click();
  });
});
