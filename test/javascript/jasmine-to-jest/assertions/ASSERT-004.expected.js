describe('matchers', () => {
  it('uses arrayContaining', () => {
    expect([1, 2, 3]).toEqual(expect.arrayContaining([1, 2]));
  });
});
