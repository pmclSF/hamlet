describe('combined', () => {
  it('spy on with return and check', () => {
    const obj = { fetch: () => null };
    spyOn(obj, 'fetch').and.returnValue('data');
    const result = obj.fetch('url');
    expect(result).toBe('data');
    expect(obj.fetch).toHaveBeenCalledWith('url');
  });
});
