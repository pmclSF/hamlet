describe('async', () => {
  it('uses promises', () => {
    return Promise.resolve(42).then(val => {
      expect(val).toBe(42);
    });
  });
});
