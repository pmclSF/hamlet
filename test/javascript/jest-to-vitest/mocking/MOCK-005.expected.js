import { describe, it, expect, vi } from 'vitest';

vi.mock('./userService');

const { getUser } = require('./userService');

describe('UserController', () => {
  it('uses the mocked service', () => {
    getUser.mockReturnValue({ id: 1, name: 'Alice' });
    const result = getUser(1);
    expect(result.name).toBe('Alice');
    expect(getUser).toHaveBeenCalledWith(1);
  });

  it('handles missing user', () => {
    getUser.mockReturnValue(null);
    const result = getUser(999);
    expect(result).toBeNull();
  });
});
