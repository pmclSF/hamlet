describe('Async assertions with resolves', () => {
  it('resolves with user data', async () => {
    await expect(fetchUser(1)).resolves.toEqual({
      id: 1,
      name: 'Alice',
    });
  });

  it('resolves to a truthy value', async () => {
    await expect(isServiceAvailable()).resolves.toBeTruthy();
  });

  it('resolves with partial match', async () => {
    await expect(fetchUser(1)).resolves.toMatchObject({
      name: 'Alice',
    });
  });
});
