describe('spies', () => {
  it('checks call count', () => {
    const spy = jasmine.createSpy('fn');
    spy();
    spy();
    expect(spy.calls.count()).toBe(2);
  });
});
