describe('async', () => {
// HAMLET-WARNING: done() callback detected. Jest supports done() but async/await is preferred. Consider refactoring to: async () => { await ... }
// Original: it('uses done', (done) => {
  it('uses done', (done) => {
    setTimeout(() => {
      expect(true).toBe(true);
      done();
    }, 100);
  });
});
