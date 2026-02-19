describe('existence assertions', () => {
  it('should check existence', async () => {
    await browser.url('/page');
    await expect($('#present')).toExist();
    await expect($('#absent')).not.toExist();
  });
});
