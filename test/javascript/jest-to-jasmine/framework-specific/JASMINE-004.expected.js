describe('suite', () => {
  fit('focused', () => {
    expect(1).toBe(1);
  });
  xit('skipped', () => {
    expect(2).toBe(2);
  });
});
