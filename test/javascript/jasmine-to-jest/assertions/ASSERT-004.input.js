describe('matchers', () => {
  it('uses arrayContaining', () => {
    expect([1, 2, 3]).toEqual(jasmine.arrayContaining([1, 2]));
  });
});
