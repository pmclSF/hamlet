describe('click actions', () => {
  it('should click', async () => {
    await browser.url('/app');
    await $('#btn').click();
    await $('#dbl').doubleClick();
  });
});
