describe('async', () => {
  it('uses done', (done) => {
    setTimeout(() => {
      expect(true).toBe(true);
      done();
    }, 100);
  });
});
