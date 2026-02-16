describe('spies', () => {
  it('throws error', () => {
    const spy = jest.fn().mockImplementation(() => { throw 'oops'; });
    expect(() => spy()).toThrow();
  });
});
