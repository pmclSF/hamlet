describe('hooks', () => {
  beforeEach(async () => {
    await browser.url('/setup');
  });

  afterEach(async () => {
    await browser.url('/teardown');
  });

  it('should use hooks', async () => {
    await $('#btn').click();
  });
});
