describe('spies', () => {
  it('gets first args', () => {
    const spy = jasmine.createSpy('fn');
    spy('first');
    spy('second');
    expect(spy.calls.first().args).toEqual(['first']);
  });
});
