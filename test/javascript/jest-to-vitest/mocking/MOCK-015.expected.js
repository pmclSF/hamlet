import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('./api');
vi.mock('./logger');

const { fetchUser } = require('./api');
const { log } = require('./logger');

describe('OrderService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('logs when a user is fetched', () => {
    fetchUser.mockReturnValue({ id: 1, name: 'Alice' });
    const handler = vi.fn();
    const user = fetchUser(1);
    handler(user);
    log(`Fetched user: ${user.name}`);
    expect(fetchUser).toHaveBeenCalledWith(1);
    expect(log).toHaveBeenCalledWith('Fetched user: Alice');
    expect(handler).toHaveBeenCalledWith({ id: 1, name: 'Alice' });
  });

  it('handles fetch failure', () => {
    fetchUser.mockReturnValue(null);
    const errorHandler = vi.fn();
    const user = fetchUser(999);
    if (!user) errorHandler('User not found');
    expect(errorHandler).toHaveBeenCalledWith('User not found');
  });
});
