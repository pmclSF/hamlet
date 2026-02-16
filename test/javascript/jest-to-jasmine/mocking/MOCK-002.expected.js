describe('mocks', () => {
  it('spies on method', () => {
    const obj = { foo: () => 42 };
    spyOn(obj, 'foo');
    obj.foo();
    expect(obj.foo).toHaveBeenCalled();
  });
});
