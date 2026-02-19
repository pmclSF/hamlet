describe('click actions', () => {
  it('should interact', async () => {
    await browser.url('/app');
    await $('#double').doubleClick();
    await $('#target').moveTo();
  });
});
