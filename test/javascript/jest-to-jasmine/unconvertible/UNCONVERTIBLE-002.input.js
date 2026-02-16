describe('test', () => {
  it('snapshot', () => {
    expect(1).toBe(1);
    expect({ a: 1 }).toMatchSnapshot();
  });
});
