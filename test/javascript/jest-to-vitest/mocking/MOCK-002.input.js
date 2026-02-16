describe('UserRepository', () => {
  it('returns a stubbed user', () => {
    const getUser = jest.fn().mockReturnValue({ name: 'Alice', role: 'admin' });
    const repo = new UserRepository(getUser);
    const user = repo.find(1);
    expect(user.name).toBe('Alice');
    expect(user.role).toBe('admin');
    expect(getUser).toHaveBeenCalledWith(1);
  });
});
