describe('test', () => {
  it('inline snapshot', () => {
    expect(1).toBe(1);
    expect('hello').toMatchInlineSnapshot('hello');
  });
});
