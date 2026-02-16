import { describe, it, expect, vi } from 'vitest';

vi.mock('./api', () => ({
  fetchUsers: vi.fn().mockResolvedValue([{ id: 1, name: 'Alice' }]),
}));

describe('Dashboard', () => {
  it('loads users', async () => {
    const users = await loadUsers();
    expect(users).toHaveLength(1);
  });
});
