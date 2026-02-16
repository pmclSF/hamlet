describe('spies', () => {
  it('calls through', () => {
    const obj = { foo: () => 42 };
    spyOn(obj, 'foo').and.callThrough();
    expect(obj.foo()).toBe(42);
  });
});
