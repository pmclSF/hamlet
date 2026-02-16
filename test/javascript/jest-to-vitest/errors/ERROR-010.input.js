describe('WarningSystem', () => {
  let consoleSpy;

  beforeEach(() => {
    consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleSpy.mockRestore();
  });

  it('should log an error when a deprecated function is called', () => {
    callDeprecatedFunction('oldMethod');
    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('deprecated')
    );
  });

  it('should log multiple warnings for chained deprecations', () => {
    callDeprecatedFunction('methodA');
    callDeprecatedFunction('methodB');
    expect(consoleSpy).toHaveBeenCalledTimes(2);
  });

  it('should capture console.warn for non-critical issues', () => {
    const warnSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});
    triggerWarning('low disk space');
    expect(warnSpy).toHaveBeenCalledWith(expect.stringMatching(/low disk space/i));
    warnSpy.mockRestore();
  });
});
