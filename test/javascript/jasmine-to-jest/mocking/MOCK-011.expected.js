describe('spies', () => {
  it('resets calls', () => {
    const spy = jest.fn();
    spy();
    spy.mockClear();
    expect(spy.mock.calls.length).toBe(0);
  });
});
