describe.each([
  ['Chrome'],
  ['Firefox'],
  ['Safari'],
])('on %s', (browser) => {
  it.each([
    [200, 'OK'],
    [404, 'Not Found'],
    [500, 'Server Error'],
  ])('handles HTTP %i (%s)', (status, message) => {
    expect(browser).toBeDefined();
    expect(status).toBeGreaterThanOrEqual(100);
    expect(message).toBeTruthy();
  });

  it('should have a valid browser name', () => {
    expect(typeof browser).toBe('string');
    expect(browser.length).toBeGreaterThan(0);
  });
});
