describe('matchers', () => {
  it('uses any', () => {
    expect({ id: 1, name: 'test' }).toEqual({
      id: expect.any(Number),
      name: expect.any(String)
    });
  });
});
