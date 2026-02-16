describe('MultiLevelLogger', () => {
  let logSpy;
  let warnSpy;
  let errorSpy;

  beforeEach(() => {
    logSpy = jest.spyOn(console, 'log').mockImplementation();
    warnSpy = jest.spyOn(console, 'warn').mockImplementation();
    errorSpy = jest.spyOn(console, 'error').mockImplementation();
  });

  afterEach(() => {
    logSpy.mockRestore();
    warnSpy.mockRestore();
    errorSpy.mockRestore();
  });

  it('should route messages to correct log levels', () => {
    const logger = {
      log(level, msg) {
        if (level === 'info') console.log(msg);
        else if (level === 'warn') console.warn(msg);
        else if (level === 'error') console.error(msg);
      },
    };

    logger.log('info', 'Server started');
    logger.log('warn', 'Disk space low');
    logger.log('error', 'Connection lost');

    expect(logSpy).toHaveBeenCalledWith('Server started');
    expect(warnSpy).toHaveBeenCalledWith('Disk space low');
    expect(errorSpy).toHaveBeenCalledWith('Connection lost');
  });

  it('should not cross-contaminate log levels', () => {
    console.log('only info');

    expect(logSpy).toHaveBeenCalledTimes(1);
    expect(warnSpy).not.toHaveBeenCalled();
    expect(errorSpy).not.toHaveBeenCalled();
  });
});
