import { describe, it, expect, vi } from 'vitest';

describe('Argument matching', () => {
  it('matches call with objectContaining', () => {
    const save = vi.fn();
    save({ id: 1, name: 'Alice', email: 'alice@example.com' });
    expect(save).toHaveBeenCalledWith(
      expect.objectContaining({ id: 1, name: 'Alice' })
    );
  });

  it('matches call with arrayContaining', () => {
    const process = vi.fn();
    process([1, 2, 3, 4, 5]);
    expect(process).toHaveBeenCalledWith(
      expect.arrayContaining([1, 3, 5])
    );
  });

  it('matches call with stringMatching', () => {
    const log = vi.fn();
    log('Error: connection timeout at 12:34:56');
    expect(log).toHaveBeenCalledWith(
      expect.stringMatching(/Error:.*timeout/)
    );
  });
});
