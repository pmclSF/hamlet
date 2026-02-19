describe('navigation', () => {
  it('should navigate around', async () => {
    await browser.url('/page1');
    await browser.refresh();
    await browser.back();
    await browser.forward();
  });
});
