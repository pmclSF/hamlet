import { describe, it, expect, vi } from 'vitest';

describe('Logger mock', () => {
  it('creates a mock constructor with mock methods', () => {
    const MockLogger = vi.fn().mockImplementation(() => ({
      log: vi.fn(),
      error: vi.fn(),
      warn: vi.fn(),
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
