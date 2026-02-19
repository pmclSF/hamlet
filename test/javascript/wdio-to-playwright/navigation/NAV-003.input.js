describe('waits', () => {
  it('should wait', async () => {
    await browser.url('/slow');
    await browser.pause(2000);
  });
});
