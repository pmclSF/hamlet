describe('test', () => {
  it('throws', () => {
    expect(() => { throw new Error('fail'); }).toThrow();
  });
});
