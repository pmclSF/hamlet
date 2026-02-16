describe('ErrorReporter', () => {
  let errorSpy;

  beforeEach(() => {
    errorSpy = jest.spyOn(console, 'error').mockImplementation();
  });

  afterEach(() => {
    errorSpy.mockRestore();
  });

  it('should report errors to stderr', () => {
    const reporter = {
      report(err) { console.error('ERROR:', err.message); },
    };

    reporter.report(new Error('disk full'));

    expect(errorSpy).toHaveBeenCalledWith('ERROR:', 'disk full');
  });

  it('should report multiple errors', () => {
    const reporter = {
      report(err) { console.error('ERROR:', err.message); },
    };

    reporter.report(new Error('timeout'));
    reporter.report(new Error('connection refused'));

    expect(errorSpy).toHaveBeenCalledTimes(2);
    expect(errorSpy).toHaveBeenLastCalledWith('ERROR:', 'connection refused');
  });
});
