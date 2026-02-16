describe('suite', () => {
  fit('focused', () => {
    expect(1).toBe(1);
  });
  xit('skipped', () => {
    expect(2).toBe(2);
  });
  it('normal', () => {
    expect(3).toBe(3);
  });
});
