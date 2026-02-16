describe('Mock lifecycle management', () => {
  const fetchData = jest.fn();
  const logger = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  afterAll(() => {
    jest.restoreAllMocks();
  });

  it('starts with clean mocks', () => {
    expect(fetchData).not.toHaveBeenCalled();
    fetchData('test');
    expect(fetchData).toHaveBeenCalledTimes(1);
  });

  it('has mocks cleared between tests', () => {
    expect(fetchData).not.toHaveBeenCalled();
    expect(logger).not.toHaveBeenCalled();
  });
});
