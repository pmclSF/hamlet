describe('unicode handling', () => {
  it('should handle accented characters', () => {
    expect('cafe\u0301').toContain('e\u0301');
    expect('caf\u00e9').toContain('\u00e9');
  });

  it('should handle CJK characters in test values', () => {
    const greeting = '\u3053\u3093\u306b\u3061\u306f';
    expect(greeting).toHaveLength(5);
    expect(greeting).toBe('\u3053\u3093\u306b\u3061\u306f');
  });

  it('should handle emoji in assertions', () => {
    const message = 'Success \ud83c\udf89\ud83c\udf8a';
    expect(message).toContain('\ud83c\udf89');
    expect(message.length).toBeGreaterThan(8);
  });

  it('should preserve RTL characters', () => {
    const arabic = '\u0645\u0631\u062d\u0628\u0627';
    expect(arabic).toBeDefined();
    expect(arabic.length).toBe(5);
  });
});
