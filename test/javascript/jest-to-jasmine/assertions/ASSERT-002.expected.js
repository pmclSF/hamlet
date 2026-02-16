describe('matchers', () => {
  it('asymmetric matchers', () => {
    expect({ id: 1 }).toEqual({
      id: jasmine.any(Number)
    });
    expect('hello').toEqual(jasmine.stringMatching('hel'));
    expect([1, 2]).toEqual(jasmine.arrayContaining([1]));
  });
});
