describe('spyObj', () => {
  it('creates spy with 3 methods', () => {
    const service = jasmine.createSpyObj('service', ['get', 'put', 'delete']);
    service.get();
    service.put();
    expect(service.get).toHaveBeenCalled();
    expect(service.put).toHaveBeenCalled();
    expect(service.delete).not.toHaveBeenCalled();
  });
});
