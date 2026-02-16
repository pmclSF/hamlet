describe('matchers', () => {
  it('uses stringMatching', () => {
    expect('hello world').toEqual(expect.stringMatching('hello'));
  });
});
