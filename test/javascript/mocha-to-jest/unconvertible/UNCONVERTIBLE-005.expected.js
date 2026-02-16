describe('test', () => {
  it('include and match', () => {
    expect('hello world').toContain('hello');
    expect('hello').toMatch(/hel/);
  });
});
