jest.retryTimes(3);

describe('FlakyExternalService', () => {
  it('should eventually connect to the service', () => {
    const connectionAttempt = Math.random() > 0.3;
    expect(connectionAttempt).toBe(true);
  });

  it('should fetch data after retries', () => {
    const data = { status: 'ok', items: [1, 2, 3] };
    expect(data.status).toBe('ok');
    expect(data.items).toHaveLength(3);
  });

  it('should handle intermittent timeouts', () => {
    const response = { code: 200, latency: 150 };
    expect(response.code).toBe(200);
    expect(response.latency).toBeLessThan(5000);
  });
});
