describe('test', () => {
  it('promise', () => {
    return Promise.resolve(42).then(val => {
      expect(val).toBe(42);
    });
  });
});
