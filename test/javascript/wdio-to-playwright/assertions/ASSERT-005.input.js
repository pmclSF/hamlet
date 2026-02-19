describe('attribute and count assertions', () => {
  it('should check attributes and counts', async () => {
    await browser.url('/page');
    await expect($('#link')).toHaveAttribute('href', '/home');
    await expect($$('.item')).toBeElementsArrayOfSize(3);
  });
});
