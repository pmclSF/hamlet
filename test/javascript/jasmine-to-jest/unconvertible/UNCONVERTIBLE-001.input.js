describe('custom', () => {
  beforeEach(() => {
    jasmine.addMatchers(customMatchers);
  });
  it('works', () => {
    expect(5).toBe(5);
  });
});
