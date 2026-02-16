describe('spies', () => {
  it('multiple spies', () => {
    const spy1 = jasmine.createSpy('spy1');
    const spy2 = jasmine.createSpy('spy2');
    spy1('a');
    spy2('b');
    expect(spy1).toHaveBeenCalledWith('a');
    expect(spy2).toHaveBeenCalledWith('b');
  });
});
