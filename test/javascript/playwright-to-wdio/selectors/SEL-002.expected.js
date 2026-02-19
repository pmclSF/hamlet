describe('text selectors', () => {
  it('should find by text', async () => {
    await browser.url('/home');
    await $(`*=Sign In`).click();
  });
});
