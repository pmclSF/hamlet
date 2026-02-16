describe('combined', () => {
  it('spy on with return and check', () => {
    const obj = { fetch: () => null };
    jest.spyOn(obj, 'fetch').mockReturnValue('data');
    const result = obj.fetch('url');
    expect(result).toBe('data');
    expect(obj.fetch).toHaveBeenCalledWith('url');
  });
});
