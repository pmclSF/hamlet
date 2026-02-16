describe('spies', () => {
  it('creates a spy object', () => {
    const obj = jasmine.createSpyObj('service', ['get', 'set']);
    obj.get();
    expect(obj.get).toHaveBeenCalled();
  });
});
