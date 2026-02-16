describe('Logger mock', () => {
  it('creates a mock constructor with mock methods', () => {
    const MockLogger = jest.fn().mockImplementation(() => ({
      log: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
    }));

    const logger = new MockLogger();
    logger.log('hello');
    logger.error('failure');

    expect(MockLogger).toHaveBeenCalledTimes(1);
    expect(logger.log).toHaveBeenCalledWith('hello');
    expect(logger.error).toHaveBeenCalledWith('failure');
    expect(logger.warn).not.toHaveBeenCalled();
  });
});
