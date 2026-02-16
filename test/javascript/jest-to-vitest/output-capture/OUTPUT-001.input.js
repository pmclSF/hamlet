describe('Logger', () => {
  it('should log info messages to stdout', () => {
    const spy = jest.spyOn(console, 'log').mockImplementation();
    const logger = {
      info(msg) { console.log('[INFO]', msg); },
    };

    logger.info('hello');

    expect(spy).toHaveBeenCalledWith('[INFO]', 'hello');
    expect(spy).toHaveBeenCalledTimes(1);
    spy.mockRestore();
  });

  it('should log multiple messages', () => {
    const spy = jest.spyOn(console, 'log').mockImplementation();
    const logger = {
      info(msg) { console.log('[INFO]', msg); },
    };

    logger.info('first');
    logger.info('second');

    expect(spy).toHaveBeenCalledTimes(2);
    expect(spy).toHaveBeenNthCalledWith(1, '[INFO]', 'first');
    expect(spy).toHaveBeenNthCalledWith(2, '[INFO]', 'second');
    spy.mockRestore();
  });
});
