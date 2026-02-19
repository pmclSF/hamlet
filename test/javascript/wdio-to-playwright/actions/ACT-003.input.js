describe('misc actions', () => {
  it('should perform various actions', async () => {
    await browser.url('/form');
    await $('#field').clearValue();
    await $('#hover-target').moveTo();
    await $('#dropdown').selectByVisibleText('Option A');
  });
});
