describe('Mock rejection', () => {
  it('mocks a rejected promise', async () => {
    const fetchData = jest.fn().mockRejectedValue(new Error('network failure'));
    await expect(fetchData()).rejects.toThrow('network failure');
    expect(fetchData).toHaveBeenCalledTimes(1);
  });

  it('mocks rejected value once then resolves', async () => {
    const fetchData = jest.fn()
      .mockRejectedValueOnce(new Error('temporary failure'))
      .mockResolvedValueOnce({ data: 'success' });

    await expect(fetchData()).rejects.toThrow('temporary failure');
    const result = await fetchData();
    expect(result.data).toBe('success');
  });

  it('mocks a resolved promise', async () => {
    const fetchData = jest.fn().mockResolvedValue({ id: 1, name: 'Alice' });
    const result = await fetchData();
    expect(result.name).toBe('Alice');
  });
});
