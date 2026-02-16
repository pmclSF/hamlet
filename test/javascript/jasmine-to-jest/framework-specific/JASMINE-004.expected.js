describe('matchers', () => {
  it('uses stringMatching', () => {
    const spy = jest.fn();
    spy('hello world');
    expect(spy).toHaveBeenCalledWith(expect.stringMatching('hello'));
  });
});
