describe('mixed', () => {
  it('has both', () => {
    const spy = jasmine.createSpy('fn');
    spy();
    expect(spy).toHaveBeenCalled();
  });
});
