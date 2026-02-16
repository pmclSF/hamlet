import { describe, it, expect, vi, beforeEach, afterEach, afterAll } from 'vitest';

describe('Mock lifecycle management', () => {
  const fetchData = vi.fn();
  const logger = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  afterAll(() => {
    vi.restoreAllMocks();
  });

  it('starts with clean mocks', () => {
    expect(fetchData).not.toHaveBeenCalled();
    fetchData('test');
    expect(fetchData).toHaveBeenCalledTimes(1);
  });

  it('has mocks cleared between tests', () => {
    expect(fetchData).not.toHaveBeenCalled();
    expect(logger).not.toHaveBeenCalled();
  });
});
