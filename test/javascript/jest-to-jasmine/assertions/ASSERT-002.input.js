describe('matchers', () => {
  it('asymmetric matchers', () => {
    expect({ id: 1 }).toEqual({
      id: expect.any(Number)
    });
    expect('hello').toEqual(expect.stringMatching('hel'));
    expect([1, 2]).toEqual(expect.arrayContaining([1]));
  });
});
