describe('spies', () => {
  it('creates a spy object', () => {
    const obj = { get: jest.fn(), set: jest.fn() };
    obj.get();
    expect(obj.get).toHaveBeenCalled();
  });
});
