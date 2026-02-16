describe('matchers', () => {
  it('uses stringMatching', () => {
    expect('hello world').toEqual(jasmine.stringMatching('hello'));
  });
});
