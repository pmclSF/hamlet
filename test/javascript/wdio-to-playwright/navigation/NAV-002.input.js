describe('navigation', () => {
  it('should navigate back and forward', async () => {
    await browser.url('/page1');
    await browser.url('/page2');
    await browser.back();
    await browser.forward();
    await browser.refresh();
  });
});
