describe('mocks', () => {
  afterEach(() => {
    // Jasmine spies are auto-cleaned between specs
  });

  it('works', () => {
    const fn = jasmine.createSpy();
    fn();
    expect(fn).toHaveBeenCalled();
  });
});
