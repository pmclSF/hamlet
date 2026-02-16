describe('MultiLayerCleanup', () => {
  const log = [];

  afterEach(() => {
    log.push('first-cleanup');
  });

  afterEach(() => {
    log.push('second-cleanup');
  });

  afterEach(() => {
    log.push('third-cleanup');
    log.length = 0;
  });

  it('should execute test logic', () => {
    log.push('test-ran');
    expect(log).toContain('test-ran');
  });

  it('should have clean state each time', () => {
    expect(log).toHaveLength(0);
    log.push('another-test');
    expect(log).toHaveLength(1);
  });
});
