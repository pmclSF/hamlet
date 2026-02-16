describe('mocks', () => {
  it('creates mock', () => {
    const fn = jasmine.createSpy();
    fn();
    expect(fn).toHaveBeenCalled();
  });
});
