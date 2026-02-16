describe('matchers', () => {
  it('uses stringMatching', () => {
    const spy = jasmine.createSpy('fn');
    spy('hello world');
    expect(spy).toHaveBeenCalledWith(jasmine.stringMatching('hello'));
  });
});
