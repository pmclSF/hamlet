jest.mock('./api', () => ({
  fetchUsers: jest.fn().mockResolvedValue([{ id: 1, name: 'Alice' }]),
}));

describe('Dashboard', () => {
  it('loads users', async () => {
    const users = await loadUsers();
    expect(users).toHaveLength(1);
  });
});
