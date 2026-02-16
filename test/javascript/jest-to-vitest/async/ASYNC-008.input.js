describe('Async error assertions with rejects', () => {
  it('rejects with an error message', async () => {
    await expect(fetchUser(-1)).rejects.toThrow('not found');
  });

  it('rejects with a specific error type', async () => {
    await expect(fetchUser(null)).rejects.toThrow(TypeError);
  });

  it('rejects and matches error object', async () => {
    await expect(fetchUser(-1)).rejects.toMatchObject({
      message: 'not found',
      code: 404,
    });
  });
});
