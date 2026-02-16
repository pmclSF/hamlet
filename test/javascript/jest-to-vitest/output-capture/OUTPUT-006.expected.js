import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('BatchProcessor', () => {
  let consoleSpy;

  beforeEach(() => {
    consoleSpy = vi.spyOn(console, 'log').mockImplementation();
  });

  afterEach(() => {
    consoleSpy.mockRestore();
  });

  it('should log exactly once per processed item', () => {
    const processor = {
      processAll(items) {
        items.forEach((item) => {
          console.log(`Processed: ${item}`);
        });
      },
    };

    processor.processAll(['a', 'b', 'c']);

    expect(consoleSpy).toHaveBeenCalledTimes(3);
  });

  it('should not log when given empty input', () => {
    const processor = {
      processAll(items) {
        items.forEach((item) => {
          console.log(`Processed: ${item}`);
        });
      },
    };

    processor.processAll([]);

    expect(consoleSpy).toHaveBeenCalledTimes(0);
    expect(consoleSpy).not.toHaveBeenCalled();
  });

  it('should log progress at intervals', () => {
    const processor = {
      run(total) {
        for (let i = 1; i <= total; i++) {
          if (i % 10 === 0) {
            console.log(`Progress: ${i}/${total}`);
          }
        }
      },
    };

    processor.run(50);

    expect(consoleSpy).toHaveBeenCalledTimes(5);
    expect(consoleSpy).toHaveBeenLastCalledWith('Progress: 50/50');
  });
});
