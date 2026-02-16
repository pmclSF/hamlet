describe('spies', () => {
  it('resets calls', () => {
    const spy = jasmine.createSpy('fn');
    spy();
    spy.calls.reset();
    expect(spy.calls.count()).toBe(0);
  });
});
