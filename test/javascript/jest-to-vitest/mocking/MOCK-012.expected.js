import { describe, it, expect, vi } from 'vitest';

describe('Object method spying', () => {
  it('spies on console.log', () => {
    const spy = vi.spyOn(console, 'log').mockImplementation(() => {});
    console.log('test message');
    expect(spy).toHaveBeenCalledWith('test message');
    spy.mockRestore();
  });

  it('spies on console.error', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    console.error('error message');
    expect(spy).toHaveBeenCalledTimes(1);
    spy.mockRestore();
  });

  it('spies on Math.random', () => {
    const spy = vi.spyOn(Math, 'random').mockReturnValue(0.5);
    const result = Math.random();
    expect(result).toBe(0.5);
    spy.mockRestore();
  });
});
