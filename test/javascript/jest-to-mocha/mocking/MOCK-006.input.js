describe('test', () => {
  it('rejects', async () => {
    const fn = jest.fn().mockRejectedValue(new Error('fail'));
    try {
      await fn();
    } catch (e) {
      expect(e.message).toBe('fail');
    }
  });
});
