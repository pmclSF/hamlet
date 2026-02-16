describe('matchers', () => {
  it('uses objectContaining', () => {
    expect({ a: 1, b: 2 }).toEqual(expect.objectContaining({ a: 1 }));
  });
});
