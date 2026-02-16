describe('Logger', () => {
  it('should log with the correct level and message', () => {
    const transport = jest.fn();
    const logger = new Logger(transport);
    logger.info('Server started on port 3000');
    expect(transport).toHaveBeenCalledWith('info', 'Server started on port 3000');
  });

  it('should pass metadata to the transport', () => {
    const transport = jest.fn();
    const logger = new Logger(transport);
    logger.error('Connection failed', { retries: 3 });
    expect(transport).toHaveBeenCalledWith('error', 'Connection failed', { retries: 3 });
  });

  it('should call the formatter with the raw message', () => {
    const formatter = jest.fn((msg) => msg.toUpperCase());
    const logger = new Logger(jest.fn(), { formatter });
    logger.warn('disk space low');
    expect(formatter).toHaveBeenCalledWith('disk space low');
  });
});
