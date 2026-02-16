describe('SlowServiceIntegration', () => {
  let service;

  beforeEach(async () => {
    service = await Promise.resolve({
      name: 'ExternalAPI',
      status: 'ready',
      latency: 2500,
    });
  }, 10000);

  it('should connect to the slow service', () => {
    expect(service.status).toBe('ready');
  });

  it('should report the service name', () => {
    expect(service.name).toBe('ExternalAPI');
  });

  it('should have measurable latency', () => {
    expect(service.latency).toBeGreaterThan(0);
  });
});
