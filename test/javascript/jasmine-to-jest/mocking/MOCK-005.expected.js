describe('spies', () => {
  it('calls fake', () => {
    const spy = jest.fn().mockImplementation(x => x * 2);
    expect(spy(5)).toBe(10);
  });
});
