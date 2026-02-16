describe('mocks', () => {
  it('resolves', async () => {
    const fn = jasmine.createSpy().and.returnValue(Promise.resolve('data'));
    const result = await fn();
    expect(result).toBe('data');
  });
});
