describe('visibility', () => {
  it('should check visibility', async () => {
    await browser.url('/page');
    await expect($('#visible')).toBeDisplayed();
    await expect($('#hidden')).not.toBeDisplayed();
  });
});
