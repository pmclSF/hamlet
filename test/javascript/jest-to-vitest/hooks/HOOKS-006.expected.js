import { describe, it, expect, beforeEach } from 'vitest';

describe('Logger', () => {
  let output;
  let logger;

  beforeEach(() => {
    output = [];
  });

  beforeEach(() => {
    logger = {
      log(msg) {
        output.push({ level: 'info', message: msg, timestamp: Date.now() });
      },
      warn(msg) {
        output.push({ level: 'warn', message: msg, timestamp: Date.now() });
      },
    };
  });

  it('should log info messages', () => {
    logger.log('Server started');
    expect(output).toHaveLength(1);
    expect(output[0].level).toBe('info');
  });

  it('should log warning messages', () => {
    logger.warn('Disk space low');
    expect(output).toHaveLength(1);
    expect(output[0].level).toBe('warn');
  });

  it('should track multiple messages', () => {
    logger.log('First');
    logger.warn('Second');
    logger.log('Third');
    expect(output).toHaveLength(3);
  });
});
