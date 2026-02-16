describe('Object method spying', () => {
  it('spies on console.log', () => {
    const spy = jest.spyOn(console, 'log').mockImplementation(() => {});
    console.log('test message');
    expect(spy).toHaveBeenCalledWith('test message');
    spy.mockRestore();
  });

  it('spies on console.error', () => {
    const spy = jest.spyOn(console, 'error').mockImplementation(() => {});
    console.error('error message');
    expect(spy).toHaveBeenCalledTimes(1);
    spy.mockRestore();
  });

  it('spies on Math.random', () => {
    const spy = jest.spyOn(Math, 'random').mockReturnValue(0.5);
    const result = Math.random();
    expect(result).toBe(0.5);
    spy.mockRestore();
  });
});
