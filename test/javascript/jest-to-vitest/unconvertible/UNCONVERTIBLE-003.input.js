jest.retryTimes(3);

describe('Flaky test', () => {
  it('sometimes fails due to timing', () => {
    const value = Math.random();
    expect(value).toBeLessThan(0.9);
  });

  it('sometimes fails due to external dependency', async () => {
    const response = await fetchUnstableEndpoint();
    expect(response.status).toBe(200);
  });
});
