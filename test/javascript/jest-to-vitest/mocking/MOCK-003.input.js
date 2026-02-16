describe('Sequential responses', () => {
  it('returns different values on successive calls', () => {
    const fetchStatus = jest.fn()
      .mockReturnValueOnce('loading')
      .mockReturnValueOnce('ready')
      .mockReturnValueOnce('complete');

    expect(fetchStatus()).toBe('loading');
    expect(fetchStatus()).toBe('ready');
    expect(fetchStatus()).toBe('complete');
    expect(fetchStatus).toHaveBeenCalledTimes(3);
  });
});
