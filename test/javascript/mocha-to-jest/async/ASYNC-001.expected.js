describe('test', () => {
// HAMLET-WARNING: done() callback detected. Jest supports done() but async/await is preferred. Consider refactoring to: async () => { await ... }
// Original: it('async with done', (done) => {
  it('async with done', (done) => {
    setTimeout(() => {
      expect(true).toBe(true);
      done();
    }, 100);
  });
});
