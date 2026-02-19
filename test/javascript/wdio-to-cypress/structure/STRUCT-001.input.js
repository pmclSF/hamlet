describe('async removal', () => {
  beforeEach(async () => {
    await browser.url('/setup');
  });

  it('should remove async', async () => {
    await $('#btn').click();
  });
});
