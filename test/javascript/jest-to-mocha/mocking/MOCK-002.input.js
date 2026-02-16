describe('test', () => {
  it('spy', () => {
    const obj = { foo: () => 42 };
    jest.spyOn(obj, 'foo');
    obj.foo();
    expect(obj.foo).toHaveBeenCalled();
  });
});
