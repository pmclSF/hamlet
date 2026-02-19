describe('async addition', () => {
  beforeEach(async () => {
    await browser.url('/setup');
  });

  it('should add async', async () => {
    await $('#btn').click();
  });
});
