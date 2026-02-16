import { describe, it, expect } from 'vitest';

describe('UserAPI', () => {
  it('fetches user data', async () => {
    const data = await fetchUser(1);
    expect(data.name).toBe('Alice');
    expect(data.id).toBe(1);
  });

  it('fetches multiple users', async () => {
    const users = await fetchUsers();
    expect(users).toHaveLength(3);
    expect(users[0].name).toBe('Alice');
  });
});
