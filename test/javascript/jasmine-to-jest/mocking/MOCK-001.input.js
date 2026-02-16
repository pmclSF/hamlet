describe('spies', () => {
  it('creates a spy', () => {
    const spy = jasmine.createSpy('mySpy');
    spy();
    expect(spy).toHaveBeenCalled();
  });
});
