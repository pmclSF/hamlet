describe('spies', () => {
  it('spies on method', () => {
    const obj = { foo: () => 42 };
    jest.spyOn(obj, 'foo');
    obj.foo();
    expect(obj.foo).toHaveBeenCalled();
  });
});
