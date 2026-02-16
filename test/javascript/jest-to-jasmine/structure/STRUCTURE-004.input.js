describe('suite', () => {
  it.only('focused', () => {
    expect(1).toBe(1);
  });
  it.skip('skipped', () => {
    expect(2).toBe(2);
  });
  it('normal', () => {
    expect(3).toBe(3);
  });
});
