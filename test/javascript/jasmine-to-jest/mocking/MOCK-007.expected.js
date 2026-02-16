describe('spies', () => {
  it('calls through', () => {
    const obj = { foo: () => 42 };
    jest.spyOn(obj, 'foo');
    expect(obj.foo()).toBe(42);
  });
});
