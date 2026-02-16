describe('edge', () => {
  it('jasmine env', () => {
    jasmine.getEnv().addReporter(reporter);
    expect(true).toBe(true);
  });
});
