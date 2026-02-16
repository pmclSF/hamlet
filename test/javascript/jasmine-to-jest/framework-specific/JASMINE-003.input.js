describe('matchers', () => {
  it('uses anything', () => {
    const spy = jasmine.createSpy('fn');
    spy(42);
    expect(spy).toHaveBeenCalledWith(jasmine.anything());
  });
});
