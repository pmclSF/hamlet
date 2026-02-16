describe('suite', () => {
  test.only('focused', () => {
    expect(1).toBe(1);
  });
  test.skip('skipped', () => {
    expect(2).toBe(2);
  });
});
