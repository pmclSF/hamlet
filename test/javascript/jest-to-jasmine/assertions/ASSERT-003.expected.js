describe('matchers', () => {
  it('object containing', () => {
    expect({ a: 1, b: 2 }).toEqual(jasmine.objectContaining({ a: 1 }));
    expect(42).toEqual(jasmine.anything());
  });
});
