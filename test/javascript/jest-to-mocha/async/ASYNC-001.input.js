describe('test', () => {
  it('async', (done) => {
    setTimeout(() => {
      expect(true).toBe(true);
      done();
    }, 100);
  });
});
