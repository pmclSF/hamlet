describe('Hook execution order', () => {
  const log = [];

  beforeAll(() => {
    log.push('outer beforeAll');
  });

  afterAll(() => {
    log.push('outer afterAll');
  });

  beforeEach(() => {
    log.push('outer beforeEach');
  });

  afterEach(() => {
    log.push('outer afterEach');
  });

  it('should run outer hooks for top-level test', () => {
    log.push('outer test');
    expect(log).toContain('outer beforeAll');
    expect(log).toContain('outer beforeEach');
  });

  describe('inner suite', () => {
    beforeEach(() => {
      log.push('inner beforeEach');
    });

    afterEach(() => {
      log.push('inner afterEach');
    });

    it('should run both outer and inner hooks', () => {
      log.push('inner test');
      // Outer beforeEach runs before inner beforeEach
      const beforeEachEntries = log.filter((e) => e.includes('beforeEach'));
      expect(beforeEachEntries.length).toBeGreaterThanOrEqual(2);
    });
  });
});
