describe('custom', () => {
  beforeEach(() => {
    // HAMLET-TODO [UNCONVERTIBLE-CUSTOM-MATCHER]: Jasmine custom matchers must be converted to expect.extend() in Jest
// Original: jasmine.addMatchers(customMatchers)
// Manual action required: Rewrite custom matchers using expect.extend()
// jasmine.addMatchers(customMatchers);
  });
  it('works', () => {
    expect(5).toBe(5);
  });
});
