describe('Automatic mock restoration', () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it('spies on a method and restores automatically', () => {
    const spy = jest.spyOn(JSON, 'parse');
    JSON.parse('{"key": "value"}');
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy).toHaveBeenCalledWith('{"key": "value"}');
  });

  it('has a clean spy after restoration', () => {
    const spy = jest.spyOn(JSON, 'stringify');
    JSON.stringify({ a: 1 });
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it('spies with mock implementation', () => {
    jest.spyOn(console, 'warn').mockImplementation(() => {});
    console.warn('test warning');
    expect(console.warn).toHaveBeenCalledWith('test warning');
  });
});
