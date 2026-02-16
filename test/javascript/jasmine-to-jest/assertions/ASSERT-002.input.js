describe('matchers', () => {
  it('uses any', () => {
    expect({ id: 1, name: 'test' }).toEqual({
      id: jasmine.any(Number),
      name: jasmine.any(String)
    });
  });
});
