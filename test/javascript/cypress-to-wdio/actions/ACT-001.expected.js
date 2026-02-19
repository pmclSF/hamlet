describe('form actions', () => {
  it('should type values', async () => {
    await browser.url('/form');
    await $('#email').setValue('user@test.com');
    await $('#field').clearValue();
  });
});
