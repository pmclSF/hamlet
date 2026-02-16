describe('spyObj', () => {
  it('creates spy with 3 methods', () => {
    const service = { get: jest.fn(), put: jest.fn(), delete: jest.fn() };
    service.get();
    service.put();
    expect(service.get).toHaveBeenCalled();
    expect(service.put).toHaveBeenCalled();
    expect(service.delete).not.toHaveBeenCalled();
  });
});
