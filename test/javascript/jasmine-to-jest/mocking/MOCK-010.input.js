describe('spies', () => {
  it('gets most recent', () => {
    const spy = jasmine.createSpy('fn');
    spy('first');
    spy('second');
    expect(spy.calls.mostRecent().args).toEqual(['second']);
  });
});
