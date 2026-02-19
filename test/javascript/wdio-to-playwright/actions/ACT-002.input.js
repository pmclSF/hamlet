describe('click actions', () => {
  it('should click elements', async () => {
    await browser.url('/app');
    await $('#single').click();
    await $('#double').doubleClick();
  });
});
